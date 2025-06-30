package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
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

type handler struct{}

const WaitBeforeExitSeconds = 3

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

		// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		_, _ = writer.Write([]byte(rsp))

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
	if err := json.Unmarshal(payload, data); err != nil {
		return "", "", err
	}

	path := clerk.GetJwtPath()
	if write {
		if err := os.WriteFile(path, pretty.Pretty(payload), JwtFilePermissions); err != nil {
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

// loginCmd represents the login command.
var loginCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "login",
	Short: "Log into an Ampersand account",
	Long:  "Log into an Ampersand account.",
	Run: func(cmd *cobra.Command, args []string) {
		DoLogout(false)
		doLogin()
	},
}

func doLogin() {
	http.Handle("/", &handler{})

	hasBrowser := canOpenBrowser()

	runBrowser := func() {
		time.Sleep(1 * time.Second)

		if hasBrowser {
			openBrowser(fmt.Sprintf("http://localhost:%d", ServerPort))
		} else {
			link := getLoginURL()

			linkMsg := fmt.Sprintf("No browser detected, please open %s in your browser to log in.", link)
			localhostMsg := fmt.Sprintf("NOTE: the login page will redirect to http://localhost:%d/...", ServerPort)

			logger.Info(linkMsg)
			logger.Info()
			logger.Info(localhostMsg)
			logger.Info("If this URL isn't accessible (e.g. you're using a remote server),")
			logger.Info("the credentials won't be saved. It's best to run this command")
			logger.Info("on a machine with a browser, but you can also overcome this using")
			logger.Info("SSH port forwarding or a proxy.")
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
		if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
			return false
		}

		if _, err := exec.LookPath("xdg-open"); err != nil {
			logger.Info("'xdg-open' command not found, cannot open browser automatically.")

			return false
		}
	case "darwin":
		// Usually safe to assume macOS has GUI, but check if stdout is a terminal
		if !isTerminal(os.Stdout.Fd()) {
			return false
		}

		if _, err := exec.LookPath("open"); err != nil {
			logger.Info("'open' command not found, cannot open browser automatically.")

			return false
		}
	case "windows":
		// There's no great way to detect headless here, so assume yes unless redirected
		if !isTerminal(os.Stdout.Fd()) {
			return false
		}

		if _, err := exec.LookPath("rundll32"); err != nil {
			logger.Info("'rundll32' command not found, cannot open browser automatically.")

			return false
		}
	default:
		return false
	}

	return true
}

// openBrowser tries to open the URL in a browser. Should work on most standard platforms.
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform: %s", runtime.GOOS) //nolint:goerr113
	}

	if err != nil {
		logger.Fatal(err.Error())
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
