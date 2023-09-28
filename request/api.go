package request

import (
	"context"
	"fmt"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/vars"
)

var API_VERSION = "v1"

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

func (c *APIClient) BatchUpsertIntegrations(ctx context.Context, reqParams BatchUpsertIntegrationsParams) ([]Integration, error) {
	url := fmt.Sprintf("%s/projects/%s/integrations:batch", c.Root, c.ProjectId)
	var integrations []Integration
	var err error
	if c.APIKey != nil && *c.APIKey != "" {
		header := Header{Key: "X-Api-Key", Value: *c.APIKey}
		_, err = c.RequestClient.Put(ctx, url, reqParams, &integrations, header)
	} else {
		// TODO: Default to token authentication and set Authorization header, instead of failing.
		logger.Fatal("Must provide an API key in the --key flag")
	}
	if err != nil {
		return nil, err
	}
	return integrations, nil
}
