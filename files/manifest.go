package files

import (
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/cli/openapi"
	"sigs.k8s.io/yaml"
)

const manifestVersion = "1.0.0"

func ParseManifest(yamlData []byte) (*openapi.Manifest, error) {
	manifest := &openapi.Manifest{}

	if err := yaml.Unmarshal(yamlData, manifest); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	return manifest, nil
}

// nolint: goerr113
func validationError(tracker *pathTracker, msg string, args ...any) error {
	err1 := fmt.Errorf(msg, args...)
	err2 := fmt.Errorf("The validation error happened at the %s", tracker.String()) //nolint:stylecheck

	return errors.Join(ErrBadManifest, err1, err2)
}

func ValidateManifest(manifest *openapi.Manifest) error {
	var tracker pathTracker

	if manifest.SpecVersion == "" {
		return validationError(tracker.PushObj("specVersion"), "The 'specVersion' field is required")
	}

	if manifest.SpecVersion != manifestVersion {
		return validationError(tracker.PushObj("specVersion"),
			"Invalid spec version: %s (only %s is supported)", manifest.SpecVersion, manifestVersion)
	}

	if len(manifest.Integrations) == 0 {
		return validationError(tracker.PushObj("integrations"),
			"No integrations found in manifest, please define at least one integration")
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
		return validationError(path.PushObj("enabled"), "The field 'enabled' is required")
	}

	return nil
}

func validateRead(read *openapi.IntegrationRead, path *pathTracker) error {
	path = path.PushObj("objects")

	if read.Objects == nil {
		return validationError(path, "The field 'objects' is required")
	}

	if len(*read.Objects) == 0 {
		return validationError(path, "The 'objects' field must contain at least one object")
	}

	for idx, obj := range *read.Objects {
		if obj.ObjectName == "" {
			return validationError(path.PushArr(idx).PushObj("objectName"), "The field 'objectName' is required")
		}

		if obj.Destination == "" {
			return validationError(path.PushArr(idx).PushObj("destination"), "The field 'destination' is required")
		}
	}

	return nil
}

func validateWrite(write *openapi.IntegrationWrite, path *pathTracker) error {
	path = path.PushObj("objects")

	if write.Objects == nil {
		return validationError(path, "The 'objects' field is required")
	}

	if len(*write.Objects) == 0 {
		return validationError(path, "The 'objects' field must contain at least one object")
	}

	for idx, obj := range *write.Objects {
		if obj.ObjectName == "" {
			return validationError(path.PushArr(idx).PushObj("objectName"),
				"The field 'objectName' is required")
		}
	}

	return nil
}

func validateSubscribeAssocChange(ace *openapi.AssociationChangeEvent, path *pathTracker) error {
	if ace.Enabled == nil {
		return validationError(path.PushObj("enabled"),
			"The field 'enabled' is required")
	}

	return nil
}

func validateSubscribeCreateEvent(event *openapi.CreateEvent, path *pathTracker) error {
	if event.Enabled == nil {
		return validationError(path.PushObj("enabled"),
			"The field 'enabled' is required")
	}

	return nil
}

func validateSubscribeUpdateEvent(event *openapi.UpdateEvent, path *pathTracker) error {
	if event.Enabled == nil {
		return validationError(path.PushObj("enabled"),
			"The field 'enabled' is required")
	}

	return nil
}

func validateSubscribeDeleteEvent(event *openapi.DeleteEvent, path *pathTracker) error {
	if event.Enabled == nil {
		return validationError(path.PushObj("enabled"),
			"The field 'enabled' is required")
	}

	return nil
}

//nolint:gocognit,cyclop,funlen
func validateSubscribe(sub *openapi.IntegrationSubscribe, path *pathTracker) error {
	path = path.PushObj("objects")

	if sub.Objects == nil {
		return validationError(path, "The 'objects' field is required")
	}

	if len(*sub.Objects) == 0 {
		return validationError(path, "The 'objects' field must contain at least one object")
	}

	for idx, obj := range *sub.Objects {
		if obj.ObjectName == "" {
			return validationError(path.PushArr(idx).PushObj("objectName"),
				"The field 'objectName' is required")
		}

		if obj.Destination == "" {
			return validationError(path.PushArr(idx).PushObj("destination"),
				"The field 'destination' is required")
		}

		if obj.AssociationChangeEvent != nil {
			if err := validateSubscribeAssocChange(obj.AssociationChangeEvent,
				path.PushArr(idx).PushObj("associationChangeEvent")); err != nil {
				return err
			}
		}

		if obj.CreateEvent != nil {
			if err := validateSubscribeCreateEvent(obj.CreateEvent,
				path.PushArr(idx).PushObj("createEvent")); err != nil {
				return err
			}
		}

		if obj.UpdateEvent != nil {
			if err := validateSubscribeUpdateEvent(obj.UpdateEvent,
				path.PushArr(idx).PushObj("updateEvent")); err != nil {
				return err
			}
		}

		if obj.DeleteEvent != nil {
			if err := validateSubscribeDeleteEvent(obj.DeleteEvent,
				path.PushArr(idx).PushObj("deleteEvent")); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateIntegration(integration openapi.Integration, path *pathTracker) error { //nolint:cyclop
	if integration.Name == "" {
		return validationError(path.PushObj("name"), "The field 'name' is required")
	}

	if integration.Provider == "" {
		return validationError(path.PushObj("provider"), "The field 'provider' is required")
	}

	if integration.Proxy == nil && integration.Read == nil && integration.Write == nil && integration.Subscribe == nil {
		return validationError(path, "At least one of proxy, read, write, or subscribe is required")
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

	if integration.Subscribe != nil {
		if err := validateSubscribe(integration.Subscribe, path.PushObj("subscribe")); err != nil {
			return err
		}
	}

	return nil
}

// GetRemovedReadObjects returns a list of read objects that were removed from
// the old revision but not in the new integration. The comparison is case-insensitive.
func GetRemovedReadObjects(oldRevision *openapi.Integration, newIntegration *openapi.Integration) []string {
	if oldRevision == nil || oldRevision.Read == nil || oldRevision.Read.Objects == nil {
		return nil
	}

	newIntegrationObjects := make(map[string]bool)

	if newIntegration.Read != nil && newIntegration.Read.Objects != nil {
		for _, obj := range *newIntegration.Read.Objects {
			newIntegrationObjects[strings.ToLower(obj.ObjectName)] = true
		}
	}

	var removedObjects []string

	for _, oldObj := range *oldRevision.Read.Objects {
		if !newIntegrationObjects[strings.ToLower(oldObj.ObjectName)] {
			removedObjects = append(removedObjects, oldObj.ObjectName)
		}
	}

	return removedObjects
}
