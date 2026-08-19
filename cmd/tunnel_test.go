package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/amp-labs/cli/request"
)

func TestRunDestinationTunnelRestoresOriginalURL(t *testing.T) {
	t.Parallel()

	var (
		patchesMu sync.Mutex
		patches   []*request.PatchDestination
	)

	updated := make(chan struct{}, 1)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPatch ||
			req.URL.Path != "/v1/projects/test-project/destinations/destination-id" {
			http.NotFound(writer, req)

			return
		}

		var patch request.PatchDestination

		err := json.NewDecoder(req.Body).Decode(&patch)
		if err != nil {
			t.Errorf("decode patch: %v", err)

			return
		}

		patchesMu.Lock()

		patches = append(patches, &patch)
		patchCount := len(patches)
		patchesMu.Unlock()

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"id":"destination-id","name":"local-webhook","type":"webhook"}`))

		if patchCount == 1 {
			updated <- struct{}{}
		}
	}))
	t.Cleanup(server.Close)

	apiKey := "test-key"
	client := &request.APIClient{
		Root:      server.URL + "/v1",
		ProjectId: "test-project",
		APIKey:    &apiKey,
		Client:    request.NewRequestClient(),
	}
	destination := Destination{
		Id:   "destination-id",
		Name: "local-webhook",
		URL:  "https://original.example.com/webhook?source=test",
		Type: "webhook",
	}

	ctx, cancel := context.WithCancel(t.Context())
	result := make(chan error, 1)
	tunnelDone := make(chan error)

	go func() {
		result <- runDestinationTunnel(ctx, client, destination, "http://127.0.0.1:4000",
			func(_ context.Context, targetURL string) (*runningTunnel, error) {
				if targetURL != "http://127.0.0.1:4000" {
					t.Errorf("target URL = %q", targetURL)
				}

				return &runningTunnel{
					publicURL: "https://temporary.trycloudflare.com",
					done:      tunnelDone,
				}, nil
			})
	}()

	<-updated
	cancel()

	err := <-result
	if err != nil {
		t.Fatalf("runDestinationTunnel() error = %v", err)
	}

	patchesMu.Lock()
	defer patchesMu.Unlock()

	if len(patches) != 2 {
		t.Fatalf("patch count = %d, want 2", len(patches))
	}

	wantURLs := []string{
		"https://temporary.trycloudflare.com/webhook?source=test",
		"https://original.example.com/webhook?source=test",
	}

	for index, patch := range patches {
		if !reflect.DeepEqual(patch.UpdateMask, []string{"metadata.url"}) {
			t.Errorf("patch %d update mask = %#v", index, patch.UpdateMask)
		}

		metadata, ok := patch.Destination["metadata"].(map[string]any)
		if !ok {
			t.Fatalf("patch %d metadata = %#v", index, patch.Destination["metadata"])
		}

		if metadata["url"] != wantURLs[index] {
			t.Errorf("patch %d URL = %q, want %q", index, metadata["url"], wantURLs[index])
		}
	}
}

func TestCloudflareTunnelURL(t *testing.T) {
	t.Parallel()

	line := "Your quick Tunnel has been created at https://small-test.trycloudflare.com"
	if got := cloudflareTunnelURL(line); got != "https://small-test.trycloudflare.com" {
		t.Fatalf("cloudflareTunnelURL() = %q", got)
	}
}
