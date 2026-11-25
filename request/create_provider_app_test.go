package request

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"
)

const providerAppTestAPIKey = "test-api-key"

func newProviderAppTestAPIClient(server *httptest.Server) *APIClient {
	apiKey := providerAppTestAPIKey

	return &APIClient{
		Root:      server.URL + "/v1",
		ProjectId: "skills-onboarding-cli-dev",
		APIKey:    &apiKey,
		Client:    &Client{Client: server.Client()},
	}
}

func TestCreateProviderApp(t *testing.T) {
	t.Parallel()

	wantParams := CreateProviderAppParams{
		Provider:     "asana",
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		Scopes:       []string{"projects:read", "projects:write"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost ||
			req.URL.Path != "/v1/projects/skills-onboarding-cli-dev/provider-apps" {
			t.Errorf("request = %s %s, want POST provider-apps path", req.Method, req.URL.Path)
		}

		if req.Header.Get("X-Api-Key") != providerAppTestAPIKey {
			t.Errorf("X-Api-Key = %q, want test-api-key", req.Header.Get("X-Api-Key"))
		}

		var params CreateProviderAppParams

		err := json.NewDecoder(req.Body).Decode(&params)
		if err != nil {
			t.Errorf("decode request: %v", err)
		}

		if !createProviderAppParamsEqual(params, wantParams) {
			t.Errorf("request body = %#v, want %#v", params, wantParams)
		}

		writer.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(writer).Encode(ProviderApp{
			Id:         "provider-app-id",
			Provider:   params.Provider,
			ClientId:   params.ClientId,
			Scopes:     params.Scopes,
			ProjectId:  "project-id",
			CreateTime: time.Date(2026, time.August, 18, 12, 0, 0, 0, time.UTC),
		})
		if err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	providerApp, err := newProviderAppTestAPIClient(server).CreateProviderApp(t.Context(), &wantParams)
	if err != nil {
		t.Fatalf("create provider app: %v", err)
	}

	if providerApp.Id != "provider-app-id" || providerApp.Provider != "asana" {
		t.Fatalf("provider app = %#v", providerApp)
	}

	if providerApp.ClientSecret != "" {
		t.Fatalf("provider app response contains client secret")
	}
}

func createProviderAppParamsEqual(left CreateProviderAppParams, right CreateProviderAppParams) bool {
	return left.Provider == right.Provider &&
		left.ClientId == right.ClientId &&
		left.ClientSecret == right.ClientSecret &&
		slices.Equal(left.Scopes, right.Scopes)
}
