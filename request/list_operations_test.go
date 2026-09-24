package request

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListOperations(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.August, 19, 0, 54, 26, 0, time.UTC)
	completedAt := startedAt.Add(time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet ||
			req.URL.Path != "/v1/projects/test-project/integrations/integration-id/installations/installation-id/operations" {
			t.Errorf("request = %s %s, want GET operations path", req.Method, req.URL.Path)
		}

		if req.Header.Get("X-Api-Key") != "test-api-key" {
			t.Errorf("X-Api-Key = %q, want test-api-key", req.Header.Get("X-Api-Key"))
		}

		writer.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(writer).Encode(map[string]any{
			"results": []Operation{
				{
					Id:             "operation-id",
					InstallationId: "installation-id",
					ActionType:     "read",
					Status:         "success",
					Resource:       "projects",
					ReadType:       "scheduled",
					CreateTime:     startedAt,
					UpdateTime:     &completedAt,
				},
			},
			"pagination": map[string]any{
				"done":          true,
				"nextPageToken": "",
			},
		})
		if err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	apiKey := "test-api-key"
	client := &APIClient{
		Root:      server.URL + "/v1",
		ProjectId: "test-project",
		APIKey:    &apiKey,
		Client:    &Client{Client: server.Client()},
	}

	operations, err := client.ListOperations(t.Context(), "integration-id", "installation-id")
	if err != nil {
		t.Fatalf("list operations: %v", err)
	}

	if len(operations) != 1 {
		t.Fatalf("operations count = %d, want 1", len(operations))
	}

	operation := operations[0]
	if operation.Id != "operation-id" || operation.Status != "success" || operation.Resource != "projects" {
		t.Fatalf("operation = %#v", operation)
	}
}
