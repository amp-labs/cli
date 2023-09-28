package request

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/amp-labs/cli/logger"
	"io"
	"log/slog"
	"net/http"
)

type RequestClient struct {
	Client *http.Client
}

// Header is a key/value pair that can be added to a request.
type Header struct {
	Key   string
	Value string
}

func NewRequestClient() *RequestClient {
	return &RequestClient{Client: http.DefaultClient}
}

func (c *RequestClient) Put(ctx context.Context,
	url string, reqBody any, result any, headers ...Header,
) (*http.Response, error) {
	req, err := makeJSONPutRequest(ctx, url, headers, reqBody)
	if err != nil {
		return nil, err
	}
	return c.makeRequestAndParseResult(req, result)
}

func (c *RequestClient) Post(ctx context.Context,
	url string, reqBody any, result any, headers ...Header,
) (*http.Response, error) {
	req, err := makeJSONPostRequest(ctx, url, headers, reqBody)
	if err != nil {
		return nil, err
	}
	return c.makeRequestAndParseResult(req, result)
}

var ErrNone200Status = errors.New("error response from API")

func (c *RequestClient) makeRequestAndParseResult(req *http.Request, result any) (*http.Response, error) {
	logger.Debugf(">>> API REQUEST:\n%+v\n", req)
	res, payload, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}
	logger.Debugf("<<< API RESPONSE:\n%+v\n", res)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return res, fmt.Errorf("%w: HTTP Status %s", ErrNone200Status, res.Status)
	}
	if err := json.Unmarshal(payload, result); err != nil {
		return nil, err
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jBody))

	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	headers = append(headers, Header{Key: "Content-Type", Value: "application/json"})
	req.ContentLength = int64(len(jBody))

	return addAcceptJSONHeaders(req, headers)
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

func (c *RequestClient) sendRequest(req *http.Request) (*http.Response, []byte, error) {
	// Send the request
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error sending request: %w", err)
	}

	// Read the response body
	body, err := io.ReadAll(res.Body)

	defer func() {
		if res != nil && res.Body != nil {
			if closeErr := res.Body.Close(); closeErr != nil {
				slog.Warn("unable to close response body", "error", closeErr)
			}
		}
	}()

	if err != nil {
		return nil, nil, fmt.Errorf("error reading response body: %w", err)
	}

	return res, body, nil
}
