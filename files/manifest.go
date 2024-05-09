package files

import (
	"fmt"

	"github.com/amp-labs/cli/openapi"
	"sigs.k8s.io/yaml"
)

func ParseManifest(yamlData []byte) (*openapi.Manifest, error) {
	manifest := &openapi.Manifest{}

	if err := yaml.Unmarshal(yamlData, manifest); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	return manifest, nil
}

func ValidateManifest(manifest *openapi.Manifest) error {
	if manifest.SpecVersion != "1.0.0" {
		return fmt.Errorf("invalid spec version: %s (only 1.0.0 is supported)", manifest.SpecVersion)
	}

	if len(manifest.Integrations) == 0 {
		return fmt.Errorf("no integrations found in manifest, please define at least one integration")
	}

	return nil
}
