package cmd

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/vars"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
	"golang.org/x/term"
)

const (
	ServerPort = 3535
)

var (
	errInvalidLoginCallback = errors.New("invalid callback URL")
	errMissingLoginPayload  = errors.New("callback URL is missing its login payload")
)

type handler struct{}

const (
	WaitBeforeExitSeconds = 3
	OSWindows             = "windows"
)

func getLoginURL() string {
	loginURL, ok := os.LookupEnv("AMP_LOGIN_URL_OVERRIDE")
	if ok {
		return loginURL
	}

	return vars.LoginURL
}

func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// This path is followed after the user logs in. The CLI Auth Client redirects to here.
	switch {
	case request.URL.Path == "/done" && request.Method == http.MethodGet:
		bts, _ := base64.StdEncoding.DecodeString(request.URL.Query().Get("p"))

		rsp, loginEmail, err := processLogin(request.Context(), bts, true)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			logger.FatalErr("error:", err)

			return
		}

		writer.WriteHeader(http.StatusOK)

		// rsp is the login-success page rendered by clerk.getHTML, whose only
		// interpolation is mustache's {{email}} -- the escaping form, so the claim
		// value cannot inject markup. gosec's taint analysis cannot see through the
		// template engine.
		// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		_, _ = writer.Write([]byte(rsp)) //nolint:gosec // G705: template-escaped, see above

		go func() {
			// Tell the user we're done and then forcefully exit the program.
			fmt.Fprint(os.Stdout, "Successfully logged in as "+loginEmail+"\n")
			time.Sleep(WaitBeforeExitSeconds * time.Second)
			os.Exit(0)
		}()

		return
	case request.URL.Path == "/" && request.Method == http.MethodGet:
		writer.Header().Set("Location", getLoginURL())
		writer.WriteHeader(http.StatusTemporaryRedirect)
	default:
		writer.WriteHeader(http.StatusNotFound)
	}
}

const JwtFilePermissions = 0o600

// processLogin takes the JWT token, verifies it, and then stores it in the jwt.json file.
func processLogin(ctx context.Context, payload []byte, write bool) (string, string, error) { //nolint:cyclop
	data := &clerk.LoginData{}

	err := json.Unmarshal(payload, data)
	if err != nil {
		return "", "", err
	}

	path := clerk.GetJwtPath()
	if write {
		// path is the XDG config path for this stage (clerk.GetJwtPath). The only
		// caller-influenced part is AMP_STAGE_OVERRIDE, an env var the user sets for
		// themselves on their own machine, so there is no cross-trust-boundary taint.
		err := os.WriteFile(path, pretty.Pretty(payload), JwtFilePermissions) //nolint:gosec // G703: user's own config path
		if err != nil {
			return "", "", err
		}
	}

	jwt, err := clerk.FetchJwt(ctx)
	if err != nil {
		return "", "", err
	}

	return clerk.DecodeJWT(jwt)
}

const ReadHeaderTimeoutSeconds = 3

func newLoginCmd(
	logout func(bool),
	login func(),
	headlessLogin func(context.Context),
) *cobra.Command {
	var headless bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log into an Ampersand account",
		Long:  "Log into an Ampersand account.",
		Run: func(cmd *cobra.Command, args []string) {
			logout(false)

			if headless {
				headlessLogin(cmd.Context())

				return
			}

			login()
		},
	}

	cmd.Flags().BoolVar(&headless, "headless", false, "Log in using a browser on another machine")

	return cmd
}

var loginCmd = newLoginCmd(DoLogout, doLogin, doHeadlessLogin) //nolint:gochecknoglobals

func doHeadlessLogin(ctx context.Context) {
	// Reuse the hosted page's existing callback so this flow needs no new auth endpoint.
	logger.Infof("Open %s in a browser.", getLoginURL())
	logger.Info("After signing in, copy the localhost URL from your browser and paste it here.")
	fmt.Fprint(os.Stdout, "Paste callback URL: ")

	callback, err := readHiddenInput(os.Stdin)

	fmt.Fprintln(os.Stdout)

	if err != nil {
		logger.FatalErr("Unable to read callback URL", err)
	}

	payload, err := parseLoginCallback(string(callback))
	if err != nil {
		logger.FatalErr("Unable to complete login", err)
	}

	_, email, err := processLogin(ctx, payload, true)
	if err != nil {
		logger.FatalErr("Unable to complete login", err)
	}

	logger.Info("Successfully logged in as " + email)
}

func readHiddenInput(input *os.File) ([]byte, error) {
	state, err := term.MakeRaw(int(input.Fd()))
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = term.Restore(int(input.Fd()), state)
	}()

	return readInputLine(input)
}

func readInputLine(reader io.Reader) ([]byte, error) {
	const deleteCharacter = '\x7f'

	var result []byte

	buffered := bufio.NewReader(reader)

	for {
		character, err := buffered.ReadByte()
		if err == nil {
			switch character {
			case '\r', '\n':
				return result, nil
			case '\b', deleteCharacter:
				if len(result) > 0 {
					result = result[:len(result)-1]
				}
			default:
				result = append(result, character)
			}

			continue
		}

		if errors.Is(err, io.EOF) && len(result) > 0 {
			return result, nil
		}

		return result, err
	}
}

func parseLoginCallback(callback string) ([]byte, error) {
	parsed, err := url.Parse(strings.TrimSpace(callback))
	if err != nil {
		return nil, fmt.Errorf("invalid callback URL: %w", err)
	}

	if parsed.Scheme != "http" ||
		parsed.Host != fmt.Sprintf("localhost:%d", ServerPort) ||
		parsed.Path != "/done" || parsed.Fragment != "" {
		return nil, errInvalidLoginCallback
	}

	encoded, found := strings.CutPrefix(parsed.RawQuery, "p=")
	if !found || encoded == "" || strings.Contains(encoded, "&") {
		return nil, errMissingLoginPayload
	}

	encoded, err = url.PathUnescape(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid callback payload: %w", err)
	}

	payload, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid callback payload: %w", err)
	}

	return payload, nil
}

func doLogin() {
	http.Handle("/", &handler{})

	hasBrowser := canOpenBrowser()

	runBrowser := func() {
		time.Sleep(1 * time.Second)

		if hasBrowser {
			openBrowser(fmt.Sprintf("http://localhost:%d", ServerPort))
		} else {
			logger.Info("No browser detected.")
			logger.Info("Stop this command and run `amp login --headless` to sign in from another machine.")
		}
	}

	go runBrowser()

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", ServerPort),
		ReadHeaderTimeout: ReadHeaderTimeoutSeconds * time.Second,
	}

	logger.FatalErr("error logging in:", server.ListenAndServe())
}

func isTerminal(fd uintptr) bool {
	// This uses golang.org/x/term
	return term.IsTerminal(int(fd))
}

func canOpenBrowser() bool {
	switch runtime.GOOS {
	case "linux":
		return canOpenBrowserLinux()
	case "darwin":
		return canOpenBrowserDarwin()
	case OSWindows:
		return canOpenBrowserWindows()
	default:
		return false
	}
}

func canOpenBrowserLinux() bool {
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		return false
	}

	_, err := exec.LookPath("xdg-open")
	if err != nil {
		logger.Info("'xdg-open' command not found, cannot open browser automatically.")

		return false
	}

	return true
}

func canOpenBrowserDarwin() bool {
	// Usually safe to assume macOS has GUI, but check if stdout is a terminal
	if !isTerminal(os.Stdout.Fd()) {
		return false
	}

	_, err := exec.LookPath("open")
	if err != nil {
		logger.Info("'open' command not found, cannot open browser automatically.")

		return false
	}

	return true
}

func canOpenBrowserWindows() bool {
	// There's no great way to detect headless here, so assume yes unless redirected
	if !isTerminal(os.Stdout.Fd()) {
		return false
	}

	_, err := exec.LookPath("rundll32")
	if err != nil {
		logger.Info("'rundll32' command not found, cannot open browser automatically.")

		return false
	}

	return true
}

// openBrowser tries to open the URL in a browser. Should work on most standard platforms.
func openBrowser(url string) {
	var err error

	// The browser launch is intentionally detached from any request context: it must
	// outlive this command, so we use context.Background() rather than a cancelable ctx.
	switch runtime.GOOS {
	case "linux":
		err = exec.CommandContext(context.Background(), "xdg-open", url).Start()
	case OSWindows:
		err = exec.CommandContext(context.Background(), "rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.CommandContext(context.Background(), "open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform: %s", runtime.GOOS) //nolint:err113
	}

	if err != nil {
		logger.Fatal(err.Error())
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
