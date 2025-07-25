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
)

const (
	ServerPort = 3535
)

type handler struct{}

const WaitBeforeExitSeconds = 3

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
		loginURL, ok := os.LookupEnv("AMP_LOGIN_URL_OVERRIDE")
		if ok {
			writer.Header().Set("Location", loginURL)
		} else {
			writer.Header().Set("Location", vars.LoginURL)
		}

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

	runBrowser := func() {
		time.Sleep(1 * time.Second)
		openBrowser(fmt.Sprintf("http://localhost:%d", ServerPort))
	}

	go runBrowser()

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", ServerPort),
		ReadHeaderTimeout: ReadHeaderTimeoutSeconds * time.Second,
	}

	logger.FatalErr("error logging in:", server.ListenAndServe())
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
