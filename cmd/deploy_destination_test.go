package cmd

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/amp-labs/cli/request"
)

func TestGeneratePatchUpdatesOnlyDestinationURL(t *testing.T) {
	t.Parallel()

	oldDestination := &request.Destination{
		Metadata: &request.WebhookMetadata{
			URL:            "https://old.example.com/webhook",
			Headers:        map[string]string{"Authorization": "secret"},
			SvixAppId:      "svix-app-id",
			SvixEndpointId: "svix-endpoint-id",
		},
	}
	newDestination := &request.Destination{
		Metadata: &request.WebhookMetadata{URL: "https://new.example.com/webhook"},
	}

	patch := generatePatch(oldDestination, newDestination)

	wantDestination := map[string]any{
		"metadata": map[string]any{"url": "https://new.example.com/webhook"},
	}
	if !reflect.DeepEqual(patch.Destination, wantDestination) {
		t.Fatalf("patch destination = %#v, want %#v", patch.Destination, wantDestination)
	}

	wantUpdateMask := []string{"metadata.url"}
	if !reflect.DeepEqual(patch.UpdateMask, wantUpdateMask) {
		t.Fatalf("patch update mask = %#v, want %#v", patch.UpdateMask, wantUpdateMask)
	}
}

func TestGetOldDestFindsExistingDestinationByName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/v1/projects/test-project/destinations" {
			http.NotFound(writer, req)

			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`[{"id":"destination-id","name":"local-webhook","type":"webhook"}]`))
	}))
	t.Cleanup(server.Close)

	apiKey := "test-key"
	client := &request.APIClient{
		Root:      server.URL + "/v1",
		ProjectId: "test-project",
		APIKey:    &apiKey,
		Client:    request.NewRequestClient(),
	}

	destination := getOldDest(t.Context(), client, &request.Destination{Name: "local-webhook"})
	if destination == nil || destination.Id != "destination-id" {
		t.Fatalf("getOldDest() = %#v, want existing destination", destination)
	}
}
