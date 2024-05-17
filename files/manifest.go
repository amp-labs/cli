package files

import (
	"errors"
	"fmt"

	"github.com/amp-labs/cli/openapi"
	"sigs.k8s.io/yaml"
)

const manifestVersion = "1.0.0"

var ErrMissingField = errors.New("missing field")

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

	for idx, integ := range manifest.Integrations {
		if err := ValidateIntegration(idx, integ); err != nil {
			return err
		}
	}

	return nil
}

func ValidateProxy(proxy *openapi.IntegrationProxy) error {
	if proxy.Enabled == nil {
		return fmt.Errorf("%w: the enabled field is required", ErrMissingField)
	}

	return nil
}

func ValidateRead(read *openapi.IntegrationRead) error {
	if read.Objects == nil {
		return fmt.Errorf("%w: standardObjects is required", ErrMissingField)
	}

	if len(*read.Objects) == 0 {
		return fmt.Errorf("%w: standardObjects must contain at least one object", ErrBadManifest)
	}

	for idx, obj := range *read.Objects {
		if obj.ObjectName == "" {
			return fmt.Errorf("%w: standardObjects[%d]: object name is required", ErrMissingField, idx)
		}

		if obj.Destination == "" {
			return fmt.Errorf("%w: standardObjects[%d]: destination is required", ErrMissingField, idx)
		}

		if obj.Schedule == "" {
			return fmt.Errorf("%w: standardObjects[%d]: schedule is required", ErrMissingField, idx)
		}

		if _, _, _, err := parseCronString(obj.Schedule); err != nil {
			return fmt.Errorf("%w: standardObjects[%d]: cron expression for schedule is invalid: %w",
				ErrBadManifest, idx, err)
		}
	}

	return nil
}

func ValidateWrite(write *openapi.IntegrationWrite) error {
	if write.Objects == nil {
		return fmt.Errorf("%w: objects is required", ErrMissingField)
	}

	if len(*write.Objects) == 0 {
		return fmt.Errorf("%w: objects must contain at least one object", ErrBadManifest)
	}

	for idx, obj := range *write.Objects {
		if obj.ObjectName == "" {
			return fmt.Errorf("%w: objects[%d]: object name is required", ErrMissingField, idx)
		}
	}

	return nil
}

func ValidateIntegration(index int, integration openapi.Integration) error { //nolint:cyclop
	if integration.Name == "" {
		return fmt.Errorf("%w: integration %d: name is required", ErrBadManifest, index)
	}

	if integration.Provider == "" {
		return fmt.Errorf("%w: integration %d: provider is required", ErrBadManifest, index)
	}

	if integration.Proxy == nil && integration.Read == nil && integration.Write == nil {
		return fmt.Errorf("%w: integration %d: at least one of proxy, read, or write is required", ErrBadManifest, index)
	}

	if integration.Proxy != nil {
		if err := ValidateProxy(integration.Proxy); err != nil {
			return fmt.Errorf("%w: integration %d: proxy is invalid: %w", ErrBadManifest, index, err)
		}
	}

	if integration.Read != nil {
		if err := ValidateRead(integration.Read); err != nil {
			return fmt.Errorf("%w: integration %d: read is invalid: %w", ErrBadManifest, index, err)
		}
	}

	if integration.Write != nil {
		if err := ValidateWrite(integration.Write); err != nil {
			return fmt.Errorf("%w: integration %d: write is invalid: %w", ErrBadManifest, index, err)
		}
	}

	return nil
}
