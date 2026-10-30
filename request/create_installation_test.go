package request

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateInstallation(t *testing.T) {
	t.Parallel()

	wantParams := CreateInstallationParams{
		GroupRef:     "test-group",
		ConnectionId: "connection-id",
		Config: CreateInstallationConfig{
			Content: json.RawMessage(`{"provider":"asana"}`),
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost ||
			req.URL.Path != "/v1/projects/test-project/integrations/integration-id/installations" {
			t.Errorf("request = %s %s, want POST installation path", req.Method, req.URL.Path)
		}

		if req.Header.Get("X-Api-Key") != "test-api-key" {
			t.Errorf("X-Api-Key = %q, want test-api-key", req.Header.Get("X-Api-Key"))
		}

		var params CreateInstallationParams

		err := json.NewDecoder(req.Body).Decode(&params)
		if err != nil {
			t.Errorf("decode request: %v", err)
		}

		if params.GroupRef != wantParams.GroupRef || params.ConnectionId != wantParams.ConnectionId {
			t.Errorf("request body = %#v, want %#v", params, wantParams)
		}

		if string(params.Config.Content) != string(wantParams.Config.Content) {
			t.Errorf("config content = %s, want %s", params.Config.Content, wantParams.Config.Content)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(Installation{
			Id:            "installation-id",
			ProjectId:     "test-project",
			IntegrationId: "integration-id",
			GroupRef:      params.GroupRef,
			ConnectionId:  params.ConnectionId,
		})
	}))
	defer server.Close()

	apiKey := "test-api-key"
	client := &APIClient{
		Root:      server.URL + "/v1",
		ProjectId: "test-project",
		APIKey:    &apiKey,
		Client:    &Client{Client: server.Client()},
	}

	installation, err := client.CreateInstallation(t.Context(), "integration-id", &wantParams)
	if err != nil {
		t.Fatalf("create installation: %v", err)
	}

	if installation.Id != "installation-id" || installation.GroupRef != wantParams.GroupRef {
		t.Fatalf("installation = %#v", installation)
	}
}
