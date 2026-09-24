package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHandleWebhookWithoutForwarding(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{"action":"subscribe","result":[]}`))
	response := httptest.NewRecorder()

	var output bytes.Buffer

	handleWebhookWithOptions(response, req, &output, "", false)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestLogWebhookDefaultsToMetadataAndOmitsFieldValues(t *testing.T) {
	t.Parallel()

	body := []byte(`{"action":"subscribe","result":[{"fields":{"email":"secret@example.com"}}]}`)
	req := &http.Request{
		Method: http.MethodPost,
		URL:    &url.URL{Path: "/webhook", RawQuery: "token=secret"},
		Header: http.Header{"Content-Type": []string{"application/json"}},
	}

	var output bytes.Buffer

	err := logWebhook(&output, req, body, false)
	if err != nil {
		t.Fatalf("logWebhook() error = %v", err)
	}

	got := output.String()
	for _, want := range []string{
		`"method":"POST"`,
		`"path":"/webhook"`,
		`"contentType":"application/json"`,
		`"byteCount":`,
		`"action":"subscribe"`,
		`"itemCount":1`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("metadata output %q does not contain %q", got, want)
		}
	}

	for _, secret := range []string{"secret@example.com", "token=secret"} {
		if strings.Contains(got, secret) {
			t.Fatalf("metadata output contains %q", secret)
		}
	}
}

func TestLogWebhookIncludesPayloadWhenRequested(t *testing.T) {
	t.Parallel()

	body := []byte(`{"action":"subscribe","result":[{"fields":{"email":"visible@example.com"}}]}`)
	req := &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/webhook"}}

	var output bytes.Buffer

	err := logWebhook(&output, req, body, true)
	if err != nil {
		t.Fatalf("logWebhook() error = %v", err)
	}

	if !strings.Contains(output.String(), "visible@example.com") {
		t.Fatalf("payload output = %q", output.String())
	}
}

func TestSummarizeWebhookUsesResultInfoCount(t *testing.T) {
	t.Parallel()

	body := []byte(`{"action":"read","resultInfo":{"type":"url","numRecords":3}}`)
	req := &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/webhook"}}

	got := summarizeWebhook(req, body)
	if got.ItemCount == nil || *got.ItemCount != 3 {
		t.Fatalf("item count = %v, want 3", got.ItemCount)
	}
}
