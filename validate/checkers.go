// Package validate supplies the amp-yaml-validator library with project-specific
// data (destinations, provider apps) fetched from the Ampersand API. The library
// exposes these as dependency-injected "checker" hooks so that the same validation
// logic can run client-side (here, against the API) or server-side (against the
// database).
package validate

import (
	"context"

	"github.com/amp-labs/amp-yaml-validator/checker"
	"github.com/amp-labs/cli/request"
)

// DestinationChecker verifies that destinations referenced by a manifest exist in
// the project. It holds a snapshot of the project's destinations fetched once, so
// repeated CheckDestination calls during validation don't hit the API again.
type DestinationChecker struct {
	names map[string]struct{}
}

// Ensure DestinationChecker satisfies the library interface.
var _ checker.DestinationChecker = (*DestinationChecker)(nil)

// NewDestinationChecker fetches the project's destinations and returns a checker
// backed by that snapshot.
func NewDestinationChecker(ctx context.Context, client *request.APIClient) (*DestinationChecker, error) {
	destinations, err := client.ListDestinations(ctx)
	if err != nil {
		return nil, err
	}

	names := make(map[string]struct{}, len(destinations))
	for _, dest := range destinations {
		names[dest.Name] = struct{}{}
	}

	return &DestinationChecker{names: names}, nil
}

// CheckDestination reports checker.ErrDestinationNotFound if the named destination
// is not present in the project, and nil otherwise.
func (c *DestinationChecker) CheckDestination(_ context.Context, destinationName string) error {
	if _, ok := c.names[destinationName]; ok {
		return nil
	}

	return checker.ErrDestinationNotFound
}

// ProviderAppChecker verifies that a provider app (OAuth credentials) is configured
// for the providers referenced by a manifest. Like DestinationChecker it works off a
// one-time snapshot of the project's provider apps.
type ProviderAppChecker struct {
	providers map[string]struct{}
}

// Ensure ProviderAppChecker satisfies the library interface.
var _ checker.ProviderAppChecker = (*ProviderAppChecker)(nil)

// NewProviderAppChecker fetches the project's provider apps and returns a checker
// backed by that snapshot.
func NewProviderAppChecker(ctx context.Context, client *request.APIClient) (*ProviderAppChecker, error) {
	apps, err := client.ListProviderApps(ctx)
	if err != nil {
		return nil, err
	}

	providers := make(map[string]struct{}, len(apps))
	for _, app := range apps {
		providers[app.Provider] = struct{}{}
	}

	return &ProviderAppChecker{providers: providers}, nil
}

// CheckProviderApp reports checker.ErrProviderAppNotFound if no provider app is
// configured for the given provider, and nil otherwise.
func (c *ProviderAppChecker) CheckProviderApp(_ context.Context, providerName string) error {
	if _, ok := c.providers[providerName]; ok {
		return nil
	}

	return checker.ErrProviderAppNotFound
}
