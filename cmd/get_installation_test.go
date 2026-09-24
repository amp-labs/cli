package cmd

import (
	"reflect"
	"testing"

	"github.com/amp-labs/cli/request"
)

func TestSummarizeInstallationShowsEnabledActions(t *testing.T) {
	t.Parallel()

	installation := &request.Installation{
		Id:            "installation-id",
		IntegrationId: "integration-id",
		HealthStatus:  "healthy",
		Group:         &request.Group{GroupRef: "group-ref"},
		Connection:    &request.Connection{Id: "connection-id"},
		Config: &request.Config{
			RevisionId: "revision-id",
			Content: map[string]any{
				"provider": "hubspot",
				"read": map[string]any{
					"objects": map[string]any{"contacts": map[string]any{}, "companies": map[string]any{}},
				},
				"write": map[string]any{
					"objects": map[string]any{"contacts": map[string]any{}},
				},
				"subscribe": map[string]any{
					"objects": map[string]any{"contacts": map[string]any{}},
				},
				"proxy": map[string]any{"enabled": true},
			},
		},
	}

	got := summarizeInstallation(installation)
	want := installationDetail{
		Id:            "installation-id",
		IntegrationId: "integration-id",
		GroupRef:      "group-ref",
		ConnectionId:  "connection-id",
		HealthStatus:  "healthy",
		RevisionId:    "revision-id",
		Provider:      "hubspot",
		Actions: installationActions{
			Read:      []string{"companies", "contacts"},
			Write:     []string{"contacts"},
			Subscribe: []string{"contacts"},
			Proxy:     true,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("summarizeInstallation() = %#v, want %#v", got, want)
	}
}

func TestSummarizeInstallationDoesNotExposeConnectionCredentials(t *testing.T) {
	t.Parallel()

	installation := &request.Installation{
		Connection: &request.Connection{
			Id: "connection-id",
			ProviderApp: &request.ProviderApp{
				ClientId:     "client-id",
				ClientSecret: "client-secret",
			},
		},
	}

	got := summarizeInstallation(installation)
	if got.ConnectionId != "connection-id" {
		t.Fatalf("connection ID = %q, want connection-id", got.ConnectionId)
	}

	if reflect.ValueOf(got).FieldByName("Connection").IsValid() {
		t.Fatal("installation detail includes the connection object")
	}
}
