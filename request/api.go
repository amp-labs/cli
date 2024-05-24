package request

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/vars"
)

var ApiVersion = "v1" //nolint:gochecknoglobals

type APIClient struct {
	Root      string
	ProjectId string
	APIKey    *string
	Client    *Client
}

func NewAPIClient(projectId string, key *string) *APIClient {
	if projectId == "" {
		// TODO: add the ability to set projectId context via a command
		// so it doesn't need to be provided with each command.
		logger.Fatal("Must provide a project ID in the --project flag")
	}

	// For testing reasons, sometimes it's useful to override the API endpoint
	rootURL, ok := os.LookupEnv("AMP_API_URL")
	if !ok {
		rootURL = vars.ApiURL
	}

	return &APIClient{
		Root:      fmt.Sprintf("%s/%s", rootURL, ApiVersion),
		ProjectId: projectId,
		APIKey:    key,
		Client:    NewRequestClient(),
	}
}

type BatchUpsertIntegrationsParams struct {
	SourceZipURL string `json:"sourceZipUrl"`
}

type Integration struct {
	Name string `json:"name"`
}

func (c *APIClient) BatchUpsertIntegrations(
	ctx context.Context, reqParams BatchUpsertIntegrationsParams,
) ([]Integration, error) {
	intURL := fmt.Sprintf("%s/projects/%s/integrations:batch", c.Root, c.ProjectId)

	var integrations []Integration

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	_, err = c.Client.Put(ctx, intURL, reqParams, &integrations, auth) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return integrations, nil
}

type SignedURL struct {
	URL    string `json:"url"`
	Bucket string `json:"bucket"`
	Path   string `json:"path"`
}

func (c *APIClient) GetPreSignedUploadURL(ctx context.Context, md5 string) (SignedURL, error) {
	genURL := c.Root + "/generate-upload-url"

	if len(md5) > 0 {
		query := url.Values{}
		query.Add("md5", md5)
		genURL = fmt.Sprintf("%s?%s", genURL, query.Encode())
	}

	var err error

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return SignedURL{}, err
	}

	signed := &SignedURL{}

	_, err = c.Client.Get(ctx, genURL, signed, auth) //nolint:bodyclose
	if err != nil {
		return SignedURL{}, err
	}

	return *signed, nil
}

func (c *APIClient) GetMyInfo(ctx context.Context) (map[string]any, error) {
	myInfoURL := c.Root + "/my-info"

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	myInfo := make(map[string]any)

	_, err = c.Client.Get(ctx, myInfoURL, &myInfo, auth) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return myInfo, nil
}

func (c *APIClient) DeleteIntegration(ctx context.Context, integrationId string) error {
	delURL := fmt.Sprintf("%s/projects/%s/integrations/%s", c.Root, c.ProjectId, integrationId)

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return err
	}

	if _, err := c.Client.Delete(ctx, delURL, auth); err != nil { //nolint:bodyclose
		return fmt.Errorf("error deleting integration: %w", err)
	}

	logger.Debugf("Deleted integration: %v", integrationId)

	return nil
}

func (c *APIClient) DeleteInstallation(ctx context.Context, integrationId string, installationId string) error {
	delURL := fmt.Sprintf(
		"%s/projects/%s/integrations/%s/installations/%s", c.Root, c.ProjectId, integrationId, installationId,
	)

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return err
	}

	if _, err := c.Client.Delete(ctx, delURL, auth); err != nil { //nolint:bodyclose
		return fmt.Errorf("error deleting installation: %w", err)
	}

	logger.Debugf("Deleted installation: %v", installationId)

	return nil
}

func (c *APIClient) getAuthHeader(ctx context.Context) (Header, error) {
	if c.APIKey != nil && *c.APIKey != "" {
		header := Header{Key: "X-Api-Key", Value: *c.APIKey}

		return header, nil
	}

	haveJwtSession, err := clerk.HasSession()
	if err != nil {
		return Header{}, err
	}

	if haveJwtSession {
		jwt, err := clerk.FetchJwt(ctx)
		if err != nil {
			return Header{}, err
		}

		return Header{
			Key:   "Authorization",
			Value: "Bearer " + jwt,
		}, nil
	}

	logger.Fatal("no authentication method found, please run 'amp login' or provide an API key through the key flag")

	panic("unreachable")
}
