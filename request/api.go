package request

import (
	"context"
	"fmt"
	"net/url"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/vars"
)

var API_VERSION = "v1" //nolint:gochecknoglobals

type APIClient struct {
	Root          string
	ProjectId     string
	APIKey        *string
	RequestClient *RequestClient
}

func NewAPIClient(projectId string, key *string) *APIClient {
	if projectId == "" {
		// TODO: add the ability to set projectId context via a command
		// so it doesn't need to be provided with each command.
		logger.Fatal("Must provide a project ID in the --project flag")
	}

	return &APIClient{
		Root:          fmt.Sprintf("%s/%s", vars.ApiURL, API_VERSION),
		ProjectId:     projectId,
		APIKey:        key,
		RequestClient: NewRequestClient(),
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

	_, err = c.RequestClient.Put(ctx, intURL, reqParams, &integrations, auth) //nolint:bodyclose
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
	genURL := fmt.Sprintf("%s/generate-upload-url", c.Root)

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

	_, err = c.RequestClient.Get(ctx, genURL, signed, auth) //nolint:bodyclose
	if err != nil {
		return SignedURL{}, err
	}

	return *signed, nil
}

func (c *APIClient) DeleteIntegration(ctx context.Context, integrationId string) error {
	delURL := fmt.Sprintf("%s/projects/%s/integrations/%s", c.Root, c.ProjectId, integrationId)

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return err
	}

	if _, err := c.RequestClient.Delete(ctx, delURL, auth); err != nil { //nolint:bodyclose
		return fmt.Errorf("error deleting integration: %w", err)
	}

	logger.Debugf("Deleted integration: %v", integrationId)

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
			Value: fmt.Sprintf("Bearer %s", jwt),
		}, nil
	}

	logger.Fatal("no authentication method found, please either log in (amp login) or provide an API key (--key)")

	panic("unreachable")
}
