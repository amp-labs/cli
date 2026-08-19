package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/amp-labs/cli/request"
)

type fakeInstallationCreator struct {
	integrationId string
	createdWith   *request.CreateInstallationParams
}

func (f *fakeInstallationCreator) CreateInstallation(
	_ context.Context,
	integrationId string,
	params *request.CreateInstallationParams,
) (*request.Installation, error) {
	f.integrationId = integrationId
	f.createdWith = params

	return &request.Installation{Id: "installation-id"}, nil
}

func TestCreateInstallationCommandReadsConfigFile(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	wantContent := json.RawMessage(`{"provider":"asana","read":{"objects":{"projects":{"objectName":"projects"}}}}`)

	err := os.WriteFile(configPath, wantContent, 0o600)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	client := &fakeInstallationCreator{}
	cmd := newCreateInstallationCmd(
		func(projectId string, _ string) installationCreator {
			if projectId != "test-project" {
				t.Fatalf("project ID = %q, want test-project", projectId)
			}

			return client
		},
		func() string { return "test-project" },
	)
	cmd.SetArgs([]string{
		"integration-id",
		"--group-ref", "test-group",
		"--connection-id", "connection-id",
		"--config", configPath,
	})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("execute create installation command: %v", err)
	}

	if client.integrationId != "integration-id" {
		t.Fatalf("integration ID = %q, want integration-id", client.integrationId)
	}

	if client.createdWith == nil {
		t.Fatal("installation was not created")
	}

	if client.createdWith.GroupRef != "test-group" || client.createdWith.ConnectionId != "connection-id" {
		t.Fatalf("create installation params = %#v", client.createdWith)
	}

	if string(client.createdWith.Config.Content) != string(wantContent) {
		t.Fatalf("config content = %s, want %s", client.createdWith.Config.Content, wantContent)
	}
}

func TestReadInstallationConfigRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")

	err := os.WriteFile(configPath, []byte(`{"provider":`), 0o600)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err = readInstallationConfig(configPath)
	if !errors.Is(err, errInvalidInstallationConfig) {
		t.Fatalf("error = %v, want errInvalidInstallationConfig", err)
	}
}

func TestReadInstallationConfigRejectsNonObject(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")

	err := os.WriteFile(configPath, []byte(`[]`), 0o600)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err = readInstallationConfig(configPath)
	if !errors.Is(err, errInvalidInstallationConfig) {
		t.Fatalf("error = %v, want errInvalidInstallationConfig", err)
	}
}
