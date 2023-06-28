package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/alexkappa/mustache"
	"github.com/amp-labs/cli/www"
	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var Html = `<!doctype html>
<html>
        <head>
                <meta charset="utf-8"/>
                <title>Login</title>
                <!--[if lt IE 9]>
                <script src="//html5shim.googlecode.com/svn/trunk/html5.js"></s\
cript>
                <![endif]-->
        </head>
        <body>
            <h1>Successfully logged in as {{email}}</h1>
        <h2>Please close this tab or page and return to the CLI</h2>
        </body>
</html>`

type loginData struct {
	UserID    string `json:"userId"`
	SessionID string `json:"sessionId"`
	Token     string `json:"token"`
}

type token struct {
	Jwt string `json:"jwt"`
}

type session struct {
	LastActiveToken token `json:"last_active_token"`
}

type response struct {
	Sessions []session `json:"sessions"`
}

type clientResponse struct {
	Response response `json:"response"`
}

type handler struct {
	realHandler http.Handler
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/done" && r.Method == "GET" {
		bts, _ := base64.StdEncoding.DecodeString(r.URL.Query().Get("p"))
		rsp, err := processLogin(bts)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Printf("error: %v", err)
			return
		}

		w.WriteHeader(200)
		_, _ = w.Write([]byte(rsp))
		go func() {
			fmt.Println("login successful")
			time.Sleep(3 * time.Second)
			os.Exit(0)
		}()
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")
	h.realHandler.ServeHTTP(w, r)
}

func processLogin(payload []byte) (string, error) {
	path := getJwtPath()
	if err := os.WriteFile(path, pretty.Pretty(payload), 0600); err != nil {
		return "", err
	}

	tmpl := mustache.New()
	if err := tmpl.ParseString(Html); err != nil {
		return "", err
	}

	dat := &loginData{}
	if err := json.Unmarshal(payload, dat); err != nil {
		return "", err
	}

	hc := http.DefaultClient

	u := "https://mighty-kingfish-66.clerk.accounts.dev/v1/client?_clerk_js_version=4.50.1&__dev_session=" + dat.Token

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Origin", "http://localhost:3535")

	rsp, err := hc.Do(req)
	if err != nil {
		return "", err
	}

	bb, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	if rsp.StatusCode != 200 {
		return "", fmt.Errorf("http %d", rsp.StatusCode)
	}

	cr := &clientResponse{}
	if err := json.Unmarshal(bb, cr); err != nil {
		return "", err
	}

	jwt := cr.Response.Sessions[0].LastActiveToken.Jwt

	c, err := clerk.NewClient("dummy")
	if err != nil {
		return "", err
	}

	claims, err := c.DecodeToken(jwt)
	if err != nil {
		return "", err
	}

	em := claims.Extra["email"].(string)
	ht, err := tmpl.RenderString(map[string]string{
		"email": em,
	})
	if err != nil {
		return "", err
	}

	return ht, nil
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
			h := &handler{
				realHandler: http.FileServer(http.FS(www.FS())),
			}
			http.Handle("/", h)
			fmt.Println("http://localhost:3535")
			go func() {
				time.Sleep(1 * time.Second)
				openBrowser("http://localhost:3535")
			}()
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

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of an ampersand account",
	Long:  "Log out of an ampersand account.",
	Run: func(cmd *cobra.Command, args []string) {
		path := getJwtPath()
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return
			} else {
				log.Fatalln(err)
			}
		}

		if err := os.Remove(path); err != nil {
			log.Fatalln(err)
		}
	},
}

var loginTest = &cobra.Command{
	Use: "test",
	Run: func(cmd *cobra.Command, args []string) {
		// Never ever use this credential in production. This is here for demo purposes only.
		c, err := clerk.NewClient("sk_test_RrsNOFiMbTBZXANx7hL1wj8LMQbJRVKEdrpTyQg7a6")
		if err != nil {
			log.Fatalln(err)
		}

		bts, err := os.ReadFile(getJwtPath())
		if err != nil {
			log.Fatalln(err)
		}

		dat := &loginData{}
		if err := json.Unmarshal(bts, dat); err != nil {
			log.Fatalln(err)
		}

		_, err = c.DecodeToken(dat.Token)
		if err != nil {
			log.Fatalln(err)
		}

		hc := http.DefaultClient

		u := "https://mighty-kingfish-66.clerk.accounts.dev/v1/client?_clerk_js_version=4.50.1&__dev_session=" + dat.Token

		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			log.Fatalln(err)
		}

		req.Header.Set("Origin", "http://localhost:3535")

		rsp, err := hc.Do(req)
		if err != nil {
			log.Fatalln(err)
		}

		bb, err := io.ReadAll(rsp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		if rsp.StatusCode != 200 {
			log.Fatalf("http %d", rsp.StatusCode)
		}

		cr := &clientResponse{}
		if err := json.Unmarshal(bb, cr); err != nil {
			log.Fatalln(err)
		}

		jwt := cr.Response.Sessions[0].LastActiveToken.Jwt

		sc, err := c.VerifyToken(jwt)
		if err != nil {
			log.Fatalln(err)
		}

		s, _ := json.MarshalIndent(sc, "", "  ")
		fmt.Println(string(s))
	},
}

func getJwtPath() string {
	path, err := xdg.ConfigFile("amp/jwt.json")
	if err != nil {
		log.Fatalln(err)
	}
	return path
}

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
	rootCmd.AddCommand(loginTest)
	rootCmd.AddCommand(logoutCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
