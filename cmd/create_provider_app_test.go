package cmd

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/amp-labs/cli/request"
)

const testClientSecret = "test-client-secret"

type fakeProviderAppClient struct {
	createdWith *request.CreateProviderAppParams
}

func (f *fakeProviderAppClient) CreateProviderApp(
	_ context.Context,
	params *request.CreateProviderAppParams,
) (*request.ProviderApp, error) {
	f.createdWith = params

	return &request.ProviderApp{
		Id:           "provider-app-id",
		Provider:     params.Provider,
		ClientId:     params.ClientId,
		ClientSecret: params.ClientSecret,
		Scopes:       params.Scopes,
	}, nil
}

func TestCreateProviderAppCommandReadsClientSecretFromStdin(t *testing.T) {
	t.Parallel()

	client := &fakeProviderAppClient{}
	cmd := newCreateProviderAppCmd(func(projectId string, _ string) providerAppClient {
		if projectId != "test-project" {
			t.Fatalf("project ID = %q, want test-project", projectId)
		}

		return client
	}, func() string {
		return "test-project"
	})
	cmd.SetIn(strings.NewReader(testClientSecret + "\n"))
	cmd.SetArgs([]string{
		"asana",
		"--client-id", "test-client-id",
		"--scope", "projects:read",
		"--scope", "projects:write",
		"--client-secret-stdin",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute create provider app command: %v", err)
	}

	want := request.CreateProviderAppParams{
		Provider:     "asana",
		ClientId:     "test-client-id",
		ClientSecret: testClientSecret,
		Scopes:       []string{"projects:read", "projects:write"},
	}
	if client.createdWith == nil || !providerAppParamsEqual(client.createdWith, &want) {
		t.Fatalf("create provider app params = %#v, want %#v", client.createdWith, want)
	}
}

func TestReadClientSecretRejectsMultipleLines(t *testing.T) {
	t.Parallel()

	_, err := readClientSecret(strings.NewReader("first\nsecond\n"))
	if err == nil {
		t.Fatal("read client secret succeeded, want error")
	}
}

func TestProviderAppCreatedMessageOmitsCredentials(t *testing.T) {
	t.Parallel()

	message := providerAppCreatedMessage(&request.ProviderApp{
		Id:           "provider-app-id",
		Provider:     "asana",
		ClientId:     "test-client-id",
		ClientSecret: testClientSecret,
		Scopes:       []string{"projects:read"},
	})

	if strings.Contains(message, "test-client-id") || strings.Contains(message, testClientSecret) {
		t.Fatalf("message contains credentials: %q", message)
	}
}

func providerAppParamsEqual(left *request.CreateProviderAppParams, right *request.CreateProviderAppParams) bool {
	return left.Provider == right.Provider &&
		left.ClientId == right.ClientId &&
		left.ClientSecret == right.ClientSecret &&
		slices.Equal(left.Scopes, right.Scopes)
}
