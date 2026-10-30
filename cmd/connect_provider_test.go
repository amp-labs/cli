package cmd

import (
	"context"
	"testing"

	"github.com/amp-labs/cli/request"
)

const (
	testAuthorizationURL = "https://provider.example/authorize"
	testOAuthProject     = "test-project"
)

type fakeOAuthClient struct {
	requestedWith *request.OAuthAuthorizationURLParams
}

func (f *fakeOAuthClient) GenerateOAuthAuthorizationURL(
	_ context.Context,
	params *request.OAuthAuthorizationURLParams,
) (string, error) {
	f.requestedWith = params

	return testAuthorizationURL, nil
}

func TestConnectProviderCommandGeneratesAndOpensAuthorizationURL(t *testing.T) {
	t.Parallel()

	client := &fakeOAuthClient{}
	openedURL := ""
	cmd := newConnectProviderCmd(
		func(projectId string, _ string) oauthClient {
			if projectId != testOAuthProject {
				t.Fatalf("project ID = %q, want test-project", projectId)
			}

			return client
		},
		func() string { return testOAuthProject },
		func() bool { return true },
		func(url string) { openedURL = url },
	)
	cmd.SetArgs([]string{
		"asana",
		"--group-ref", "test-group",
		"--consumer-ref", "test-consumer",
		"--provider-app", "provider-app-id",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute connect provider command: %v", err)
	}

	want := request.OAuthAuthorizationURLParams{
		ProjectIdOrName: testOAuthProject,
		Provider:        "asana",
		GroupRef:        "test-group",
		ConsumerRef:     "test-consumer",
		ProviderAppId:   "provider-app-id",
	}
	if client.requestedWith == nil || *client.requestedWith != want {
		t.Fatalf("authorization params = %#v, want %#v", client.requestedWith, want)
	}

	if openedURL != testAuthorizationURL {
		t.Fatalf("opened URL = %q, want %q", openedURL, testAuthorizationURL)
	}
}

func TestConnectProviderCommandDoesNotOpenBrowserWhenUnavailable(t *testing.T) {
	t.Parallel()

	client := &fakeOAuthClient{}
	cmd := newConnectProviderCmd(
		func(_ string, _ string) oauthClient { return client },
		func() string { return testOAuthProject },
		func() bool { return false },
		func(url string) { t.Fatalf("opened URL in headless mode: %s", url) },
	)
	cmd.SetArgs([]string{
		"asana",
		"--group-ref", "test-group",
		"--consumer-ref", "test-consumer",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute connect provider command: %v", err)
	}

	if client.requestedWith == nil {
		t.Fatal("authorization URL was not requested")
	}
}
