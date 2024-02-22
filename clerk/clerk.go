package clerk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/adrg/xdg"
	"github.com/alexkappa/mustache"
	"github.com/amp-labs/cli/vars"
	"github.com/clerkinc/clerk-sdk-go/clerk"
)

// HTML is shown to the user after they log in.
var HTML = `<!doctype html>` + "\n" + //nolint:gochecknoglobals
	`<html>
        <head>
                <meta charset="utf-8"/>
                <title>Login</title>
                <!--[if lt IE 9]>
                <script src="//html5shim.googlecode.com/svn/trunk/html5.js"></script>
                <![endif]-->
        </head>
        <body>
            <h3>Successfully logged in as {{email}}</h3>
            <div>Please close this tab or page and return to the CLI</div>
        </body>
</html>`

const (
	ClientSessionPathDev  = "%s/v1/client?_clerk_js_version=4.50.1&__dev_session=%s"
	ClientSessionPathProd = "%s/v1/client?_clerk_js_version=4.50.1"
)

var clerkLogin *LoginData //nolint:gochecknoglobals

type LoginData struct {
	UserID    string            `json:"userId"`
	SessionID string            `json:"sessionId"`
	Token     string            `json:"token"`
	Cookies   map[string]string `json:"cookies"`
}

type token struct {
	Jwt string `json:"jwt"`
}

type verification struct {
	Status string `json:"status"`
}

type email struct {
	ID           string       `json:"id"`
	Address      string       `json:"email_address"` //nolint:tagliatelle
	Verification verification `json:"verification"`
}

type phone struct {
	ID           string       `json:"id"`
	Number       string       `json:"phone_number"` //nolint:tagliatelle
	Verification verification `json:"verification"`
}

type user struct {
	ID             string  `json:"id"`
	Username       string  `json:"username"`
	FirstName      string  `json:"first_name"`               //nolint:tagliatelle
	LastName       string  `json:"last_name"`                //nolint:tagliatelle
	ImageURL       string  `json:"image_url"`                //nolint:tagliatelle
	PrimaryEmail   string  `json:"primary_email_address_id"` //nolint:tagliatelle
	PrimaryPhone   string  `json:"primary_phone_number_id"`  //nolint:tagliatelle
	EmailAddresses []email `json:"email_addresses"`          //nolint:tagliatelle
	PhoneNumbers   []phone `json:"phone_numbers"`            //nolint:tagliatelle
}

type session struct {
	LastActiveToken token `json:"last_active_token"` //nolint:tagliatelle
	CreatedAt       int64 `json:"created_at"`        //nolint:tagliatelle
	UpdatedAt       int64 `json:"updated_at"`        //nolint:tagliatelle
	User            user  `json:"user"`
}

type response struct {
	Sessions []session `json:"sessions"`
}

type clientResponse struct {
	Response response `json:"response"`
}

func GetClerkRootURL() string {
	clerkRoot, ok := os.LookupEnv("AMP_CLERK_URL_OVERRIDE")
	if ok {
		return clerkRoot
	}

	return vars.ClerkRootURL
}

func GetSessionURL(data *LoginData) string {
	if vars.Stage == "prod" {
		return fmt.Sprintf(ClientSessionPathProd, GetClerkRootURL())
	}

	return fmt.Sprintf(ClientSessionPathDev, GetClerkRootURL(), data.Token)
}

func GetJwtFile() string {
	if vars.Stage == "prod" {
		return "amp/jwt.json"
	}

	return fmt.Sprintf("amp/jwt-%s.json", vars.Stage)
}

// GetJwtPath returns the path to the jwt.json file where the JWT token is stored.
func GetJwtPath() string {
	path, err := xdg.ConfigFile(GetJwtFile())
	if err != nil {
		log.Fatalln(err)
	}

	return path
}

func GetClerkDomain() string {
	u, err := url.Parse(GetClerkRootURL())
	if err != nil {
		log.Fatalln(err)
	}

	return u.Hostname()
}

func HasSession() (bool, error) {
	if clerkLogin != nil {
		return true, nil
	}

	st, err := os.Stat(GetJwtPath())
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if st != nil {
		return !st.IsDir(), nil
	}

	return false, nil
}

var ErrNoSessions = errors.New("no sessions found in response")

func FetchJwt(ctx context.Context) (string, error) { //nolint:funlen,cyclop
	if clerkLogin == nil {
		contents, err := os.ReadFile(GetJwtPath())
		if err != nil {
			return "", fmt.Errorf("error reading jwt file: %w", err)
		}

		data := &LoginData{}
		if err := json.Unmarshal(contents, data); err != nil {
			return "", fmt.Errorf("error unmarshalling jwt file: %w", err)
		}

		clerkLogin = data
	}

	// Call out to clerk and ask for session info using the JWT token.
	u := GetSessionURL(clerkLogin)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	for cookieName, cookieValue := range clerkLogin.Cookies {
		// This is doing the same thing that net/http does (filtering out
		// invalid characters), but it's doing it in a way that's not
		// going to log a noisy error message.
		var builder strings.Builder

		// See the function http.sanitizeCookieValue for where this
		// logic comes from. It's not a verbatim copy, although
		// validCookieValueRune is essentially identical to
		// validCookieValueByte.
		for _, char := range cookieValue {
			if validCookieValueRune(char) {
				builder.WriteRune(char)
			}
		}

		req.AddCookie(&http.Cookie{
			Name:     cookieName,
			Value:    builder.String(),
			Path:     "/",
			Domain:   GetClerkDomain(),
			Secure:   true,
			HttpOnly: true,
		})
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	bb, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	if rsp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http %d (%s)", rsp.StatusCode, string(bb)) //nolint:goerr113
	}

	cr := &clientResponse{}
	if err := json.Unmarshal(bb, cr); err != nil {
		return "", fmt.Errorf("error unmarshalling response body: %w", err)
	}

	if len(cr.Response.Sessions) == 0 {
		return "", ErrNoSessions
	}

	jwt := cr.Response.Sessions[0].LastActiveToken.Jwt

	return jwt, nil
}

var ErrMissingEmail = errors.New("couldn't find email address in claims")

func DecodeJWT(jwt string) (string, string, error) {
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
	emailStr, ok := claims.Extra["email"].(string)
	if !ok {
		return "", "", ErrMissingEmail
	}

	ht, err := getHTML(emailStr)
	if err != nil {
		return "", "", err
	}

	// Return the HTML and email
	return ht, emailStr, nil
}

func getHTML(emailStr string) (string, error) {
	// Render the HTML
	tmpl := mustache.New()
	if err := tmpl.ParseString(HTML); err != nil {
		return "", err
	}

	ht, err := tmpl.RenderString(map[string]string{
		"email": emailStr,
	})
	if err != nil {
		return "", err
	}

	return ht, nil
}

func validCookieValueRune(r rune) bool {
	return 0x20 <= r && r < 0x7f && r != '"' && r != ';' && r != '\\'
}
