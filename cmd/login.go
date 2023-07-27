package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/adrg/xdg"
	"github.com/alexkappa/mustache"
	"github.com/amp-labs/cli/vars"
	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

// Html is shown to the user after they log in
var Html = `<!doctype html>
<html>
        <head>
                <meta charset="utf-8"/>
                <title>Login</title>
                <!--[if lt IE 9]>
                <script src="//html5shim.googlecode.com/svn/trunk/html5.js"></script>
                <![endif]-->
        </head>
        <body>
            <h1>Successfully logged in as {{email}}</h1>
        <h2>Please close this tab or page and return to the CLI</h2>
        </body>
</html>`

// loginData is the data that is stored in the jwt.json file
type loginData struct {
	UserID    string `json:"userId"`
	SessionID string `json:"sessionId"`
	Token     string `json:"token"`
}

type token struct {
	Jwt string `json:"jwt"`
}

type verification struct {
	Status string `json:"status"`
}

type email struct {
	ID           string       `json:"id"`
	Address      string       `json:"email_address"`
	Verification verification `json:"verification"`
}

type phone struct {
	ID           string       `json:"id"`
	Number       string       `json:"phone_number"`
	Verification verification `json:"verification"`
}

type user struct {
	ID             string  `json:"id"`
	Username       string  `json:"username"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	ImageURL       string  `json:"image_url"`
	PrimaryEmail   string  `json:"primary_email_address_id"`
	PrimaryPhone   string  `json:"primary_phone_number_id"`
	EmailAddresses []email `json:"email_addresses"`
	PhoneNumbers   []phone `json:"phone_numbers"`
}

type session struct {
	LastActiveToken token `json:"last_active_token"`
	CreatedAt       int64 `json:"created_at"`
	UpdatedAt       int64 `json:"updated_at"`
	User            user  `json:"user"`
}

type response struct {
	Sessions []session `json:"sessions"`
}

type clientResponse struct {
	Response response `json:"response"`
}

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// This path is followed after the user logs in. The CLI Auth Client redirects to here.
	if r.URL.Path == "/done" && r.Method == "GET" {
		// Extract the JWT token
		bts, _ := base64.StdEncoding.DecodeString(r.URL.Query().Get("p"))

		// Process the JWT token
		rsp, loginEmail, err := processLogin(bts)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Printf("error: %v", err)
			return
		}

		w.WriteHeader(200)

		// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		_, _ = w.Write([]byte(rsp))

		go func() {
			// Tell the user we're done and then forcefully exit the program.
			fmt.Printf("Successfully logged in as %s\n", loginEmail)
			time.Sleep(3 * time.Second)
			os.Exit(0)
		}()

		return
	} else if r.URL.Path == "/" && r.Method == "GET" {
		// When the user first interacts with the login, this is what they see. (immediate redirect to the react app)
		w.Header().Set("Location", vars.LoginURL)
		w.WriteHeader(307) // redirect
	} else {
		w.WriteHeader(404)
	}
}

// processLogin takes the JWT token, verifies it, and then stores it in the jwt.json file.
func processLogin(payload []byte) (string, string, error) {
	path := getJwtPath()
	if err := os.WriteFile(path, pretty.Pretty(payload), 0600); err != nil {
		return "", "", err
	}

	dat := &loginData{}
	if err := json.Unmarshal(payload, dat); err != nil {
		return "", "", err
	}

	// Call out to clerk and ask for session info using the JWT token.
	hc := http.DefaultClient
	u := fmt.Sprintf("%s/v1/client?_clerk_js_version=4.50.1&__dev_session=%s",
		vars.ClerkRootURL, dat.Token)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Origin", "http://localhost:3535")

	rsp, err := hc.Do(req)
	if err != nil {
		return "", "", err
	}

	bb, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", "", err
	}

	if rsp.StatusCode != 200 {
		return "", "", fmt.Errorf("http %d", rsp.StatusCode)
	}

	cr := &clientResponse{}
	if err := json.Unmarshal(bb, cr); err != nil {
		return "", "", err
	}

	jwt := cr.Response.Sessions[0].LastActiveToken.Jwt

	// Using a dummy value here because DecodeToken doesn't actually use the secret.
	c, err := clerk.NewClient("dummy")
	if err != nil {
		return "", "", err
	}

	// Extract the claims (which includes the email address) from the JWT token.
	claims, err := c.DecodeToken(jwt)
	if err != nil {
		return "", "", err
	}

	// Grab the email address from the claims.
	em := claims.Extra["email"].(string)

	// Render the HTML
	tmpl := mustache.New()
	if err := tmpl.ParseString(Html); err != nil {
		return "", "", err
	}
	ht, err := tmpl.RenderString(map[string]string{
		"email": em,
	})
	if err != nil {
		return "", "", err
	}

	// Return the HTML and email
	return ht, em, nil
}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to an ampersand account",
	Long:  "Log in to an ampersand account.",
	Run: func(cmd *cobra.Command, args []string) {
		needLogin := false
		path := getJwtPath()
		fi, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				needLogin = true
			} else {
				log.Fatalln(err)
			}
		}

		if needLogin {
			http.Handle("/", &handler{})
			fmt.Println("http://localhost:3535")
			go func() {
				time.Sleep(1 * time.Second)
				openBrowser("http://localhost:3535")
			}()

			// nosemgrep: go.lang.security.audit.net.use-tls.use-tls
			log.Fatalln(http.ListenAndServe(":3535", nil))
		} else {
			if fi.IsDir() {
				log.Fatalln("jwt path isn't a regular file:", path)
			}
			log.Println("already logged in!")
			os.Exit(0)
		}
	},
}

func getJwtName() string {
	if vars.Stage == "prod" {
		return "amp/jwt.json"
	}
	return fmt.Sprintf("amp/jwt-%s.json", vars.Stage)
}

// getJwtPath returns the path to the jwt.json file where the JWT token is stored.
func getJwtPath() string {
	path, err := xdg.ConfigFile(getJwtName())
	if err != nil {
		log.Fatalln(err)
	}
	return path
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
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
