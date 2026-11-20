package request

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateOAuthAuthorizationURL(t *testing.T) {
	t.Parallel()

	wantParams := OAuthAuthorizationURLParams{
		ProjectIdOrName: "test-project",
		Provider:        "asana",
		GroupRef:        "test-group",
		ConsumerRef:     "test-consumer",
		ProviderAppId:   "provider-app-id",
	}
	wantURL := "https://provider.example/authorize"

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost || req.URL.Path != "/v1/oauth-connect" {
			t.Errorf("request = %s %s, want POST /v1/oauth-connect", req.Method, req.URL.Path)
		}

		if req.Header.Get("X-Api-Key") != "test-api-key" {
			t.Errorf("X-Api-Key = %q, want test-api-key", req.Header.Get("X-Api-Key"))
		}

		var params OAuthAuthorizationURLParams

		err := json.NewDecoder(req.Body).Decode(&params)
		if err != nil {
			t.Errorf("decode request: %v", err)
		}

		if params != wantParams {
			t.Errorf("request body = %#v, want %#v", params, wantParams)
		}

		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = writer.Write([]byte(wantURL))
	}))
	defer server.Close()

	apiKey := "test-api-key"
	client := &APIClient{
		Root:      server.URL + "/v1",
		ProjectId: "test-project",
		APIKey:    &apiKey,
		Client:    &Client{Client: server.Client()},
	}

	url, err := client.GenerateOAuthAuthorizationURL(t.Context(), &wantParams)
	if err != nil {
		t.Fatalf("generate OAuth authorization URL: %v", err)
	}

	if url != wantURL {
		t.Fatalf("authorization URL = %q, want %q", url, wantURL)
	}
}
