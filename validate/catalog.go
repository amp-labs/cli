package validate

import (
	"context"
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/catalog"
	"github.com/amp-labs/cli/request"
	"github.com/amp-labs/connectors/providers"
)

// CatalogProvider implements the validator's catalog.CatalogProvider interface using
// the Ampersand API's live provider catalog (the "dynamic catalog"), which is updated
// several times a day. Backing validation with this keeps provider, module, and
// capability checks current instead of relying on the catalog compiled into the
// connectors library, which goes stale between releases.
//
// It holds a snapshot fetched once so repeated lookups during a single validation run
// don't re-hit the API. When the fetch fails, callers should fall back to the embedded
// catalog (catalog.NewDefaultCatalogProvider) so validation still works offline.
type CatalogProvider struct {
	catalog map[string]providers.ProviderInfo
}

// Ensure CatalogProvider satisfies the library interface.
var _ catalog.CatalogProvider = (*CatalogProvider)(nil)

// NewCatalogProvider fetches the live provider catalog from the API and returns a
// provider backed by that snapshot. The catalog endpoint is public, so this works
// even without a configured project or an active login session.
func NewCatalogProvider(ctx context.Context) (*CatalogProvider, error) {
	var cat providers.CatalogType

	err := request.FetchProviderCatalog(ctx, &cat)
	if err != nil {
		return nil, err
	}

	return &CatalogProvider{catalog: cat}, nil
}

// GetProviderInfo retrieves provider information from the fetched catalog.
func (c *CatalogProvider) GetProviderInfo(
	_ context.Context,
	providerName string,
) (*providers.ProviderInfo, error) {
	info, ok := c.catalog[providerName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", catalog.ErrProviderNotFound, providerName)
	}

	return &info, nil
}

// GetProviderSupport retrieves the provider's capabilities from the fetched catalog.
func (c *CatalogProvider) GetProviderSupport(
	ctx context.Context,
	providerName string,
) (*providers.Support, error) {
	info, err := c.GetProviderInfo(ctx, providerName)
	if err != nil {
		return nil, err
	}

	return &info.Support, nil
}

// GetModuleInfo retrieves module information for a given provider and module ID.
func (c *CatalogProvider) GetModuleInfo(
	ctx context.Context,
	providerName string,
	moduleID string,
) (*providers.ModuleInfo, error) {
	info, err := c.GetProviderInfo(ctx, providerName)
	if err != nil {
		return nil, err
	}

	if info.Modules == nil {
		return nil, fmt.Errorf("%w: %s", catalog.ErrProviderNoModules, providerName)
	}

	mods := *info.Modules

	mod, ok := mods[moduleID]
	if !ok {
		return nil, fmt.Errorf("%w: %s for provider %s", catalog.ErrModuleNotFound, moduleID, providerName)
	}

	return &mod, nil
}

// ListObjects reports ErrNotSupported: the provider catalog does not expose object
// schemas, matching the connectors default implementation.
func (c *CatalogProvider) ListObjects(
	_ context.Context,
	_ string,
	_ string,
) ([]string, error) {
	return nil, catalog.ErrNotSupported
}

// Ping reports whether the fetched catalog is usable. The validator gates all
// provider-specific validation (including provider-app checks) on this, so a
// successfully fetched, non-empty catalog must report healthy.
func (c *CatalogProvider) Ping(_ context.Context) error {
	if len(c.catalog) == 0 {
		return catalog.ErrProviderNotFound
	}

	return nil
}
