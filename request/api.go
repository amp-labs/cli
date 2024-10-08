package request

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/openapi"
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

type IntegrationName struct {
	Name string `json:"name"`
}

func (c *APIClient) BatchUpsertIntegrations(
	ctx context.Context, reqParams BatchUpsertIntegrationsParams,
) ([]IntegrationName, error) {
	intURL := fmt.Sprintf("%s/projects/%s/integrations:batch", c.Root, c.ProjectId)

	var integrations []IntegrationName

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

func (c *APIClient) ListIntegrations(ctx context.Context) ([]*Integration, error) {
	listURL := fmt.Sprintf("%s/projects/%s/integrations", c.Root, c.ProjectId)

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	var installations []*Integration

	_, err = c.Client.Get(ctx, listURL, &installations, auth) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return installations, nil
}

func (c *APIClient) ListInstallations(ctx context.Context, integrationId string) ([]*Installation, error) {
	listURL := fmt.Sprintf("%s/projects/%s/integrations/%s/installations", c.Root, c.ProjectId, integrationId)

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	var installations []*Installation

	_, err = c.Client.Get(ctx, listURL, &installations, auth) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return installations, nil
}

func (c *APIClient) ListConnections(ctx context.Context) ([]*Connection, error) {
	listURL := fmt.Sprintf("%s/projects/%s/connections", c.Root, c.ProjectId)

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	var connections []*Connection

	_, err = c.Client.Get(ctx, listURL, &connections, auth) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return connections, nil
}

func (c *APIClient) GetCatalog(ctx context.Context) (openapi.CatalogType, error) {
	catalogURL := c.Root + "/providers"

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	var catalog openapi.CatalogType

	_, err = c.Client.Get(ctx, catalogURL, &catalog, auth) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return catalog, nil
}

func (c *APIClient) ListProviderApps(ctx context.Context) ([]*ProviderApp, error) {
	listURL := fmt.Sprintf("%s/projects/%s/provider-apps", c.Root, c.ProjectId)

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	var providerApps []*ProviderApp

	_, err = c.Client.Get(ctx, listURL, &providerApps, auth) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return providerApps, nil
}

func (c *APIClient) ListProjects(ctx context.Context) ([]*Project, error) {
	listURL := c.Root + "/projects"

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	var projects []*Project

	_, err = c.Client.Get(ctx, listURL, &projects, auth) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (c *APIClient) ListDestinations(ctx context.Context) ([]*Destination, error) {
	listURL := fmt.Sprintf("%s/projects/%s/destinations", c.Root, c.ProjectId)

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	var destinations []*Destination

	_, err = c.Client.Get(ctx, listURL, &destinations, auth) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return destinations, nil
}

func (c *APIClient) CreateDestination(ctx context.Context, dest *Destination) (*Destination, error) {
	createURL := fmt.Sprintf("%s/projects/%s/destinations", c.Root, c.ProjectId)

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	var out Destination

	_, err = c.Client.Post(ctx, createURL, dest, &out, auth) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (c *APIClient) GetDestination(ctx context.Context, id string) (*Destination, error) {
	getURL := fmt.Sprintf("%s/projects/%s/destinations/%s", c.Root, c.ProjectId, id)

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	var out Destination

	_, err = c.Client.Get(ctx, getURL, &out, auth) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (c *APIClient) PatchDestination(ctx context.Context, id string, patch *PatchDestination) (*Destination, error) {
	patchURL := fmt.Sprintf("%s/projects/%s/destinations/%s", c.Root, c.ProjectId, id)

	auth, err := c.getAuthHeader(ctx)
	if err != nil {
		return nil, err
	}

	var out Destination

	_, err = c.Client.Patch(ctx, patchURL, patch, &out, auth) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return &out, nil
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
