package request

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httputil"

	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/utils"
)

const clientName = "amp-cli"

type Client struct {
	Client         *http.Client
	DefaultHeaders []Header
}

// Header is a key/value pair that can be added to a request.
type Header struct {
	Key   string
	Value string
}

type InputValidationIssue struct {
	In           string         `json:"in,omitempty"`
	Name         string         `json:"name,omitempty"`
	Value        any            `json:"value,omitempty"`
	Detail       string         `json:"detail,omitempty"`
	Href         string         `json:"href,omitempty"`
	Instance     string         `json:"instance,omitempty"`
	Status       int32          `json:"status,omitempty"`
	Title        string         `json:"title,omitempty"`
	Type         string         `json:"type,omitempty"`
	Subsystem    string         `json:"subsystem,omitempty"`
	Time         string         `json:"time,omitempty"`
	RequestID    string         `json:"requestId,omitempty"`
	Causes       []string       `json:"causes,omitempty"`
	Remedy       string         `json:"remedy,omitempty"`
	SupportEmail string         `json:"supportEmail,omitempty"`
	SupportPhone string         `json:"supportPhone,omitempty"`
	SupportURL   string         `json:"supportUrl,omitempty"`
	Retryable    *bool          `json:"retryable,omitempty"`
	RetryAfter   string         `json:"retryAfter,omitempty"`
	Context      map[string]any `json:"context,omitempty"`
}

type ProblemError struct {
	Detail       string                 `json:"detail,omitempty"`
	Href         string                 `json:"href,omitempty"`
	Instance     string                 `json:"instance,omitempty"`
	Status       int32                  `json:"status,omitempty"`
	Title        string                 `json:"title,omitempty"`
	Type         string                 `json:"type,omitempty"`
	Subsystem    string                 `json:"subsystem,omitempty"`
	Time         string                 `json:"time,omitempty"`
	RequestID    string                 `json:"requestId,omitempty"`
	Causes       []string               `json:"causes,omitempty"`
	Remedy       string                 `json:"remedy,omitempty"`
	SupportEmail string                 `json:"supportEmail,omitempty"`
	SupportPhone string                 `json:"supportPhone,omitempty"`
	SupportURL   string                 `json:"supportUrl,omitempty"`
	Retryable    *bool                  `json:"retryable,omitempty"`
	RetryAfter   string                 `json:"retryAfter,omitempty"`
	Context      map[string]any         `json:"context,omitempty"`
	Issues       []InputValidationIssue `json:"issues,omitempty"`
}

func (p *ProblemError) Error() string {
	if p == nil {
		return "<nil>"
	}

	js, _ := json.MarshalIndent(p, "", "  ") //nolint:errchkjson

	return string(js)
}

func NewRequestClient() *Client {
	versionInfo := utils.GetVersionInformation()

	headers := []Header{
		{Key: "X-Amp-Client", Value: clientName},
		{Key: "X-Amp-Client-Version", Value: versionInfo.Version},
		{Key: "X-Amp-Client-Commit", Value: versionInfo.CommitID},
		{Key: "X-Amp-Client-Branch", Value: versionInfo.Branch},
		{Key: "X-Amp-Client-Build-Date", Value: versionInfo.BuildDate},
	}

	if versionInfo.Stage != utils.Prod {
		// We really only care if these are non-prod clients. Otherwise
		// it's safe to just assume prod.
		headers = append(headers, Header{
			Key:   "X-Amp-Client-Stage",
			Value: string(versionInfo.Stage),
		})
	}

	return &Client{
		Client:         http.DefaultClient,
		DefaultHeaders: headers,
	}
}

// Get makes a GET request to the desired URL, and unmarshalls the
// response body into `result`.
func (c *Client) Get(ctx context.Context,
	url string, result any, headers ...Header,
) (*http.Response, error) {
	allHeaders := c.DefaultHeaders
	allHeaders = append(allHeaders, headers...)

	req, err := makeJSONGetRequest(ctx, url, allHeaders)
	if err != nil {
		return nil, err
	}

	return c.makeRequestAndParseJSONResult(req, result)
}

// Put makes a PUT request to the desired URL, and unmarshalls the
// response body into `result`.
func (c *Client) Put(ctx context.Context,
	url string, reqBody any, result any, headers ...Header,
) (*http.Response, error) {
	allHeaders := c.DefaultHeaders
	allHeaders = append(allHeaders, headers...)

	req, err := makeJSONPutRequest(ctx, url, allHeaders, reqBody)
	if err != nil {
		return nil, err
	}

	return c.makeRequestAndParseJSONResult(req, result)
}

// Post makes a POST request to the desired URL, and unmarshalls the
// response body into `result`.
func (c *Client) Post(ctx context.Context,
	url string, reqBody any, result any, headers ...Header,
) (*http.Response, error) {
	allHeaders := c.DefaultHeaders
	allHeaders = append(allHeaders, headers...)

	req, err := makeJSONPostRequest(ctx, url, allHeaders, reqBody)
	if err != nil {
		return nil, err
	}

	return c.makeRequestAndParseJSONResult(req, result)
}

// Delete makes a Delete request to the desired URL for plain text requests.
func (c *Client) Delete(ctx context.Context,
	url string, headers ...Header,
) (*http.Response, error) {
	allHeaders := c.DefaultHeaders
	allHeaders = append(allHeaders, headers...)

	req, err := makeDeleteRequest(ctx, url, allHeaders)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return c.makeRequest(req)
}

var ErrNon200Status = errors.New("error response from API")

func (c *Client) makeRequestAndParseJSONResult(req *http.Request, result any) (*http.Response, error) {
	dump, _ := httputil.DumpRequest(req, false)
	logger.Debugf("\n>>> API REQUEST:\n%v>>> END OF API REQUEST\n", string(dump))

	res, payload, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 { //nolint:nestif
		ct := res.Header.Get("Content-Type")
		if len(ct) > 0 {
			mt, _, err := mime.ParseMediaType(ct)
			if err == nil {
				if mt == "application/problem+json" {
					prob := &ProblemError{}
					if err := json.Unmarshal(payload, prob); err == nil {
						return res, fmt.Errorf("%w: %w", ErrNon200Status, prob)
					}
				}
			}
		}

		return res, fmt.Errorf("%w: HTTP Status %s", ErrNon200Status, res.Status)
	}

	if err := json.Unmarshal(payload, result); err != nil {
		return nil, err
	}

	return res, nil
}

func makeJSONGetRequest(ctx context.Context, url string, headers []Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return addAcceptJSONHeaders(req, headers)
}

func (c *Client) makeRequest(req *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequest(req, false)
	logger.Debugf("\n>>> API REQUEST:\n%v>>> END OF API REQUEST\n", string(dump))

	res, _, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return res, fmt.Errorf("%w: HTTP Status %s", ErrNon200Status, res.Status)
	}

	return res, nil
}

func makeJSONPostRequest(ctx context.Context, url string, headers []Header, body any) (*http.Request, error) {
	jBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("request body is not valid JSON, body is %v:\n%w", body, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	headers = append(headers, Header{Key: "Content-Type", Value: "application/json"})
	req.ContentLength = int64(len(jBody))

	return addAcceptJSONHeaders(req, headers)
}

func makeJSONPutRequest(ctx context.Context, url string, headers []Header, body any) (*http.Request, error) {
	jBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("request body is not valid JSON, body is %v:\n%w", body, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(jBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	headers = append(headers, Header{Key: "Content-Type", Value: "application/json"})
	req.ContentLength = int64(len(jBody))

	return addAcceptJSONHeaders(req, headers)
}

func makeDeleteRequest(ctx context.Context, url string, headers []Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.ContentLength = 0

	return addHeaders(req, headers), nil
}

func addHeaders(req *http.Request, headers []Header) *http.Request {
	// Apply any custom headers
	for _, hdr := range headers {
		req.Header.Add(hdr.Key, hdr.Value)
	}

	return req
}

func addAcceptJSONHeaders(req *http.Request, headers []Header) (*http.Request, error) {
	// Request JSON
	req.Header.Add("Accept", "application/json")

	// Apply any custom headers
	for _, hdr := range headers {
		req.Header.Add(hdr.Key, hdr.Value)
	}

	return req, nil
}

func (c *Client) sendRequest(req *http.Request) (*http.Response, []byte, error) {
	// Send the request
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error sending request: %w", err)
	}

	dump, _ := httputil.DumpResponse(res, true)
	logger.Debugf("\n<<< API RESPONSE:\n%v\n<<< END OF API RESPONSE\n", string(dump))

	// Read the response body
	body, err := io.ReadAll(res.Body)

	defer func() {
		if res != nil && res.Body != nil {
			if closeErr := res.Body.Close(); closeErr != nil {
				logger.Debugf("unable to close response body %v", closeErr)
			}
		}
	}()

	if err != nil {
		return nil, nil, fmt.Errorf("error reading response body: %w", err)
	}

	return res, body, nil
}
