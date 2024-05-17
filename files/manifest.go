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
	var tracker pathTracker

	if manifest.SpecVersion != manifestVersion {
		return fmt.Errorf("%w: %s: invalid spec version: %s (only %s is supported)",
			ErrBadManifest, tracker.PushObj("specVersion").String(), manifest.SpecVersion, manifestVersion)
	}

	if len(manifest.Integrations) == 0 {
		return fmt.Errorf("%w: %s: no integrations found in manifest, please define at least one integration",
			ErrBadManifest, tracker.PushObj("integrations").String())
	}

	for idx, integ := range manifest.Integrations {
		if err := validateIntegration(integ, tracker.PushObj("integrations").PushArr(idx)); err != nil {
			return err
		}
	}

	return nil
}

func validateProxy(proxy *openapi.IntegrationProxy, path *pathTracker) error {
	if proxy.Enabled == nil {
		return fmt.Errorf("%w: %s: the enabled field is required",
			ErrMissingField, path.PushObj("enabled").String())
	}

	return nil
}

func validateRead(read *openapi.IntegrationRead, path *pathTracker) error {
	path = path.PushObj("objects")

	if read.Objects == nil {
		return fmt.Errorf("%w: %s: objects is required",
			ErrMissingField, path.String())
	}

	if len(*read.Objects) == 0 {
		return fmt.Errorf("%w: %s: objects must contain at least one object",
			ErrBadManifest, path.String())
	}

	for idx, obj := range *read.Objects {
		if obj.ObjectName == "" {
			return fmt.Errorf("%w: %s: objectName is required",
				ErrMissingField, path.PushArr(idx).PushObj("objectName").String())
		}

		if obj.Destination == "" {
			return fmt.Errorf("%w: %s: destination is required",
				ErrMissingField, path.PushArr(idx).PushObj("destination").String())
		}

		if obj.Schedule == "" {
			return fmt.Errorf("%w: %s: schedule is required",
				ErrMissingField, path.PushArr(idx).PushObj("schedule").String())
		}
	}

	return nil
}

func validateWrite(write *openapi.IntegrationWrite, path *pathTracker) error {
	path = path.PushObj("objects")

	if write.Objects == nil {
		return fmt.Errorf("%w: %s: objects is required",
			ErrMissingField, path.String())
	}

	if len(*write.Objects) == 0 {
		return fmt.Errorf("%w: %s: objects must contain at least one object",
			ErrBadManifest, path.String())
	}

	for idx, obj := range *write.Objects {
		if obj.ObjectName == "" {
			return fmt.Errorf("%w: %s: object name is required",
				ErrMissingField, path.PushArr(idx).PushObj("objectName").String())
		}
	}

	return nil
}

func validateIntegration(integration openapi.Integration, path *pathTracker) error { //nolint:cyclop
	if integration.Name == "" {
		return fmt.Errorf("%w: %s: name is required",
			ErrBadManifest, path.PushObj("name").String())
	}

	if integration.Provider == "" {
		return fmt.Errorf("%w: %s: provider is required",
			ErrBadManifest, path.PushObj("provider").String())
	}

	if integration.Proxy == nil && integration.Read == nil && integration.Write == nil {
		return fmt.Errorf("%w: %s: at least one of proxy, read, or write is required",
			ErrBadManifest, path.String())
	}

	if integration.Proxy != nil {
		if err := validateProxy(integration.Proxy, path.PushObj("proxy")); err != nil {
			return err
		}
	}

	if integration.Read != nil {
		if err := validateRead(integration.Read, path.PushObj("read")); err != nil {
			return err
		}
	}

	if integration.Write != nil {
		if err := validateWrite(integration.Write, path.PushObj("write")); err != nil {
			return err
		}
	}

	return nil
}
