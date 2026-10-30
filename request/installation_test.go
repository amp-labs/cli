package request

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetInstallation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet ||
			req.URL.Path != "/v1/projects/test-project/integrations/integration-id/installations/installation-id" {
			http.NotFound(writer, req)

			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"id":"installation-id","integrationId":"integration-id"}`))
	}))
	t.Cleanup(server.Close)

	client := newInstallationTestClient(server.URL)

	installation, err := client.GetInstallation(t.Context(), "integration-id", "installation-id")
	if err != nil {
		t.Fatalf("GetInstallation() error = %v", err)
	}

	if installation.Id != "installation-id" {
		t.Fatalf("installation ID = %q, want installation-id", installation.Id)
	}
}

func TestPatchInstallation(t *testing.T) {
	t.Parallel()

	wantPatch := PatchInstallation{
		Installation: map[string]any{
			"config": map[string]any{
				"content": map[string]any{
					"proxy": map[string]any{"enabled": true},
				},
			},
		},
		UpdateMask: []string{"config.content.proxy.enabled"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPatch ||
			req.URL.Path != "/v1/projects/test-project/integrations/integration-id/installations/installation-id" {
			http.NotFound(writer, req)

			return
		}

		var gotPatch PatchInstallation

		err := json.NewDecoder(req.Body).Decode(&gotPatch)
		if err != nil {
			t.Fatalf("decode patch: %v", err)
		}

		if !reflect.DeepEqual(gotPatch, wantPatch) {
			t.Fatalf("patch = %#v, want %#v", gotPatch, wantPatch)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"id":"installation-id","integrationId":"integration-id"}`))
	}))
	t.Cleanup(server.Close)

	client := newInstallationTestClient(server.URL)

	installation, err := client.PatchInstallation(
		t.Context(), "integration-id", "installation-id", &wantPatch,
	)
	if err != nil {
		t.Fatalf("PatchInstallation() error = %v", err)
	}

	if installation.Id != "installation-id" {
		t.Fatalf("installation ID = %q, want installation-id", installation.Id)
	}
}

func newInstallationTestClient(root string) *APIClient {
	apiKey := "test-key"

	return &APIClient{
		Root:      root + "/v1",
		ProjectId: "test-project",
		APIKey:    &apiKey,
		Client:    NewRequestClient(),
	}
}
