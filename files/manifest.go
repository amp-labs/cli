package files

import (
	"errors"
	"fmt"

	"github.com/amp-labs/cli/openapi"
	"sigs.k8s.io/yaml"
)

const manifestVersion = "1.0.0"

var ErrBadManifest = errors.New("invalid manifest")

func ParseManifest(yamlData []byte) (*openapi.Manifest, error) {
	manifest := &openapi.Manifest{}

	if err := yaml.Unmarshal(yamlData, manifest); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	return manifest, nil
}

func ValidateManifest(manifest *openapi.Manifest) error {
	if manifest.SpecVersion != manifestVersion {
		return fmt.Errorf("%w: invalid spec version: %s (only %s is supported)",
			ErrBadManifest, manifest.SpecVersion, manifestVersion)
	}

	if len(manifest.Integrations) == 0 {
		return fmt.Errorf("%w: no integrations found in manifest, please define at least one integration",
			ErrBadManifest)
	}

	return nil
}
