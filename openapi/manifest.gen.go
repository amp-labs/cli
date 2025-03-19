// Package openapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package openapi

import (
	"encoding/json"

	"github.com/oapi-codegen/runtime"
)

// Defines values for AssociationChangeEventEnabled.
const (
	AssociationChangeEventEnabledAlways AssociationChangeEventEnabled = "always"
)

// Defines values for CreateEventEnabled.
const (
	CreateEventEnabledAlways CreateEventEnabled = "always"
)

// Defines values for DeleteEventEnabled.
const (
	DeleteEventEnabledAlways DeleteEventEnabled = "always"
)

// Defines values for DeliveryMode.
const (
	Auto      DeliveryMode = "auto"
	OnRequest DeliveryMode = "onRequest"
)

// Defines values for FieldMetadataValueType.
const (
	Boolean      FieldMetadataValueType = "boolean"
	Date         FieldMetadataValueType = "date"
	Datetime     FieldMetadataValueType = "datetime"
	Float        FieldMetadataValueType = "float"
	Int          FieldMetadataValueType = "int"
	MultiSelect  FieldMetadataValueType = "multiSelect"
	Other        FieldMetadataValueType = "other"
	SingleSelect FieldMetadataValueType = "singleSelect"
	String       FieldMetadataValueType = "string"
)

// Defines values for OptionalFieldsAutoOption.
const (
	All OptionalFieldsAutoOption = "all"
)

// Defines values for UpdateEventEnabled.
const (
	UpdateEventEnabledAlways UpdateEventEnabled = "always"
)

// AssociationChangeEvent defines model for AssociationChangeEvent.
type AssociationChangeEvent struct {
	// Enabled If always, the integration will subscribe to association change events.
	Enabled *AssociationChangeEventEnabled `json:"enabled,omitempty"`

	// IncludeFullRecords If true, the integration will include full records in the event payload.
	IncludeFullRecords *bool `json:"includeFullRecords,omitempty"`
}

// AssociationChangeEventEnabled If always, the integration will subscribe to association change events.
type AssociationChangeEventEnabled string

// Backfill defines model for Backfill.
type Backfill struct {
	DefaultPeriod DefaultPeriod `json:"defaultPeriod"`
}

// CreateEvent defines model for CreateEvent.
type CreateEvent struct {
	// Enabled If always, the integration will subscribe to create events.
	Enabled *CreateEventEnabled `json:"enabled,omitempty"`
}

// CreateEventEnabled If always, the integration will subscribe to create events.
type CreateEventEnabled string

// DefaultPeriod defines model for DefaultPeriod.
type DefaultPeriod struct {
	// Days Number of days in past to backfill from. 0 is no backfill. e.g) if 10, then backfill last 10 days of data. Required if fullHistory is not set.
	Days *int `json:"days,omitempty" validate:"required_without=FullHistory,omitempty,min=0"`

	// FullHistory If true, backfill all history. Required if days is not set.
	FullHistory *bool `json:"fullHistory,omitempty" validate:"required_without=Days"`
}

// DeleteEvent defines model for DeleteEvent.
type DeleteEvent struct {
	// Enabled If always, the integration will subscribe to delete events.
	Enabled *DeleteEventEnabled `json:"enabled,omitempty"`
}

// DeleteEventEnabled If always, the integration will subscribe to delete events.
type DeleteEventEnabled string

// Delivery defines model for Delivery.
type Delivery struct {
	// Mode The data delivery mode for this object. If not specified, defaults to automatic.
	Mode *DeliveryMode `json:"mode,omitempty"`

	// PageSize The number of records to receive per data delivery.
	PageSize *int `json:"pageSize,omitempty"`
}

// DeliveryMode The data delivery mode for this object. If not specified, defaults to automatic.
type DeliveryMode string

// FieldMetadata defines model for FieldMetadata.
type FieldMetadata struct {
	// DisplayName The display name of the field from the provider API.
	DisplayName string `json:"displayName"`

	// FieldName The name of the field from the provider API.
	FieldName string `json:"fieldName"`

	// ProviderType Raw field type from the provider API.
	ProviderType string `json:"providerType,omitempty"`

	// ReadOnly Whether the field is read-only.
	ReadOnly bool `json:"readOnly,omitempty"`

	// ValueType A normalized field type
	ValueType FieldMetadataValueType `json:"valueType,omitempty"`

	// Values If the valueType is singleSelect or multiSelect, this is a list of possible values
	Values []FieldValue `json:"values,omitempty"`
}

// FieldMetadataValueType A normalized field type
type FieldMetadataValueType string

// FieldValue Represents a field value
type FieldValue struct {
	// DisplayValue The human-readable display value
	DisplayValue string `json:"displayValue"`

	// Value The internal value used by the system
	Value string `json:"value"`
}

// HydratedIntegration defines model for HydratedIntegration.
type HydratedIntegration struct {
	DisplayName *string                   `json:"displayName,omitempty"`
	Name        string                    `json:"name"`
	Provider    string                    `json:"provider"`
	Proxy       *HydratedIntegrationProxy `json:"proxy,omitempty"`
	Read        *HydratedIntegrationRead  `json:"read,omitempty"`
	Write       *HydratedIntegrationWrite `json:"write,omitempty"`
}

// HydratedIntegrationField defines model for HydratedIntegrationField.
type HydratedIntegrationField struct {
	union json.RawMessage
}

// HydratedIntegrationFieldExistent defines model for HydratedIntegrationFieldExistent.
type HydratedIntegrationFieldExistent struct {
	DisplayName string `json:"displayName"`
	FieldName   string `json:"fieldName"`

	// MapToDisplayName The display name to map to in the destination.
	MapToDisplayName string `json:"mapToDisplayName,omitempty"`

	// MapToName The field name to map to in the destination.
	MapToName string `json:"mapToName,omitempty"`
}

// HydratedIntegrationObject defines model for HydratedIntegrationObject.
type HydratedIntegrationObject struct {
	// AllFields This is a list of all fields on the object for a particular SaaS instance, including their display names.
	AllFields *[]HydratedIntegrationField `json:"allFields,omitempty"`

	// AllFieldsMetadata This is a map of all fields on the object including their metadata (such as display name and type), the keys of the map are the field names.
	AllFieldsMetadata *map[string]FieldMetadata `json:"allFieldsMetadata,omitempty"`
	Backfill          *Backfill                 `json:"backfill,omitempty"`
	Destination       string                    `json:"destination"`
	DisplayName       string                    `json:"displayName"`

	// MapToDisplayName A display name to map to.
	MapToDisplayName string `json:"mapToDisplayName,omitempty"`

	// MapToName An object name to map to.
	MapToName          string                      `json:"mapToName,omitempty"`
	ObjectName         string                      `json:"objectName"`
	OptionalFields     *[]HydratedIntegrationField `json:"optionalFields,omitempty"`
	OptionalFieldsAuto *OptionalFieldsAutoOption   `json:"optionalFieldsAuto,omitempty"`
	RequiredFields     *[]HydratedIntegrationField `json:"requiredFields,omitempty"`
	Schedule           string                      `json:"schedule"`
}

// HydratedIntegrationProxy defines model for HydratedIntegrationProxy.
type HydratedIntegrationProxy struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// HydratedIntegrationRead defines model for HydratedIntegrationRead.
type HydratedIntegrationRead struct {
	Objects *[]HydratedIntegrationObject `json:"objects,omitempty"`
}

// HydratedIntegrationWrite defines model for HydratedIntegrationWrite.
type HydratedIntegrationWrite struct {
	Objects *[]HydratedIntegrationWriteObject `json:"objects,omitempty"`
}

// HydratedIntegrationWriteObject defines model for HydratedIntegrationWriteObject.
type HydratedIntegrationWriteObject struct {
	DisplayName string `json:"displayName"`
	ObjectName  string `json:"objectName"`

	// ValueDefaults Configuration to set default write values for object fields.
	ValueDefaults *ValueDefaults `json:"valueDefaults,omitempty"`
}

// Integration defines model for Integration.
type Integration struct {
	DisplayName *string               `json:"displayName,omitempty"`
	Name        string                `json:"name"`
	Provider    string                `json:"provider"`
	Proxy       *IntegrationProxy     `json:"proxy,omitempty"`
	Read        *IntegrationRead      `json:"read,omitempty"`
	Subscribe   *IntegrationSubscribe `json:"subscribe,omitempty"`
	Write       *IntegrationWrite     `json:"write,omitempty"`
}

// IntegrationField defines model for IntegrationField.
type IntegrationField struct {
	union json.RawMessage
}

// IntegrationFieldExistent defines model for IntegrationFieldExistent.
type IntegrationFieldExistent struct {
	FieldName string `json:"fieldName"`

	// MapToDisplayName The display name to map to.
	MapToDisplayName string `json:"mapToDisplayName,omitempty"`

	// MapToName The field name to map to.
	MapToName string `json:"mapToName,omitempty"`
}

// IntegrationFieldMapping defines model for IntegrationFieldMapping.
type IntegrationFieldMapping struct {
	Default          *string `json:"default,omitempty"`
	MapToDisplayName *string `json:"mapToDisplayName,omitempty"`
	MapToName        string  `json:"mapToName"`
	Prompt           *string `json:"prompt,omitempty"`
}

// IntegrationObject defines model for IntegrationObject.
type IntegrationObject struct {
	Backfill    *Backfill `json:"backfill,omitempty"`
	Delivery    *Delivery `json:"delivery,omitempty"`
	Destination string    `json:"destination"`

	// MapToDisplayName A display name to map to.
	MapToDisplayName string `json:"mapToDisplayName,omitempty"`

	// MapToName An object name to map to.
	MapToName          string                    `json:"mapToName,omitempty"`
	ObjectName         string                    `json:"objectName"`
	OptionalFields     *[]IntegrationField       `json:"optionalFields,omitempty"`
	OptionalFieldsAuto *OptionalFieldsAutoOption `json:"optionalFieldsAuto,omitempty"`
	RequiredFields     *[]IntegrationField       `json:"requiredFields,omitempty"`
	Schedule           string                    `json:"schedule"`
}

// IntegrationProxy defines model for IntegrationProxy.
type IntegrationProxy struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// IntegrationRead defines model for IntegrationRead.
type IntegrationRead struct {
	Objects *[]IntegrationObject `json:"objects,omitempty"`
}

// IntegrationSubscribe defines model for IntegrationSubscribe.
type IntegrationSubscribe struct {
	Objects *[]IntegrationSubscribeObject `json:"objects,omitempty"`
}

// IntegrationSubscribeObject defines model for IntegrationSubscribeObject.
type IntegrationSubscribeObject struct {
	AssociationChangeEvent *AssociationChangeEvent `json:"associationChangeEvent,omitempty"`
	CreateEvent            *CreateEvent            `json:"createEvent,omitempty"`
	DeleteEvent            *DeleteEvent            `json:"deleteEvent,omitempty"`
	Destination            string                  `json:"destination"`
	ObjectName             string                  `json:"objectName"`
	OtherEvents            *OtherEvents            `json:"otherEvents,omitempty"`
	UpdateEvent            *UpdateEvent            `json:"updateEvent,omitempty"`
}

// IntegrationWrite defines model for IntegrationWrite.
type IntegrationWrite struct {
	Objects *[]IntegrationWriteObject `json:"objects,omitempty"`
}

// IntegrationWriteObject defines model for IntegrationWriteObject.
type IntegrationWriteObject struct {
	// InheritMapping If true, the write object will inherit the mapping from the read object. If false, the write object will have no mapping.
	InheritMapping *bool  `json:"inheritMapping,omitempty"`
	ObjectName     string `json:"objectName"`

	// ValueDefaults Configuration to set default write values for object fields.
	ValueDefaults *ValueDefaults `json:"valueDefaults,omitempty"`
}

// Manifest This is the schema of the manifest file that is used to define the integrations of the project.
type Manifest struct {
	Integrations []Integration `json:"integrations"`

	// SpecVersion The version of the manifest spec that this file conforms to.
	SpecVersion string `json:"specVersion"`
}

// OptionalFieldsAutoOption defines model for OptionalFieldsAutoOption.
type OptionalFieldsAutoOption string

// OtherEvents defines model for OtherEvents.
type OtherEvents = []string

// UpdateEvent defines model for UpdateEvent.
type UpdateEvent struct {
	// Enabled If always, the integration will subscribe to update events.
	Enabled             *UpdateEventEnabled `json:"enabled,omitempty"`
	RequiredWatchFields *[]string           `json:"requiredWatchFields,omitempty"`
}

// UpdateEventEnabled If always, the integration will subscribe to update events.
type UpdateEventEnabled string

// ValueDefaults Configuration to set default write values for object fields.
type ValueDefaults struct {
	// AllowAnyFields If true, users can set default values for any field.
	AllowAnyFields *bool `json:"allowAnyFields,omitempty"`
}

// AsHydratedIntegrationFieldExistent returns the union data inside the HydratedIntegrationField as a HydratedIntegrationFieldExistent
func (t HydratedIntegrationField) AsHydratedIntegrationFieldExistent() (HydratedIntegrationFieldExistent, error) {
	var body HydratedIntegrationFieldExistent
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromHydratedIntegrationFieldExistent overwrites any union data inside the HydratedIntegrationField as the provided HydratedIntegrationFieldExistent
func (t *HydratedIntegrationField) FromHydratedIntegrationFieldExistent(v HydratedIntegrationFieldExistent) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeHydratedIntegrationFieldExistent performs a merge with any union data inside the HydratedIntegrationField, using the provided HydratedIntegrationFieldExistent
func (t *HydratedIntegrationField) MergeHydratedIntegrationFieldExistent(v HydratedIntegrationFieldExistent) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

// AsIntegrationFieldMapping returns the union data inside the HydratedIntegrationField as a IntegrationFieldMapping
func (t HydratedIntegrationField) AsIntegrationFieldMapping() (IntegrationFieldMapping, error) {
	var body IntegrationFieldMapping
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromIntegrationFieldMapping overwrites any union data inside the HydratedIntegrationField as the provided IntegrationFieldMapping
func (t *HydratedIntegrationField) FromIntegrationFieldMapping(v IntegrationFieldMapping) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeIntegrationFieldMapping performs a merge with any union data inside the HydratedIntegrationField, using the provided IntegrationFieldMapping
func (t *HydratedIntegrationField) MergeIntegrationFieldMapping(v IntegrationFieldMapping) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

func (t HydratedIntegrationField) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

func (t *HydratedIntegrationField) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err
}

// AsIntegrationFieldExistent returns the union data inside the IntegrationField as a IntegrationFieldExistent
func (t IntegrationField) AsIntegrationFieldExistent() (IntegrationFieldExistent, error) {
	var body IntegrationFieldExistent
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromIntegrationFieldExistent overwrites any union data inside the IntegrationField as the provided IntegrationFieldExistent
func (t *IntegrationField) FromIntegrationFieldExistent(v IntegrationFieldExistent) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeIntegrationFieldExistent performs a merge with any union data inside the IntegrationField, using the provided IntegrationFieldExistent
func (t *IntegrationField) MergeIntegrationFieldExistent(v IntegrationFieldExistent) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

// AsIntegrationFieldMapping returns the union data inside the IntegrationField as a IntegrationFieldMapping
func (t IntegrationField) AsIntegrationFieldMapping() (IntegrationFieldMapping, error) {
	var body IntegrationFieldMapping
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromIntegrationFieldMapping overwrites any union data inside the IntegrationField as the provided IntegrationFieldMapping
func (t *IntegrationField) FromIntegrationFieldMapping(v IntegrationFieldMapping) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeIntegrationFieldMapping performs a merge with any union data inside the IntegrationField, using the provided IntegrationFieldMapping
func (t *IntegrationField) MergeIntegrationFieldMapping(v IntegrationFieldMapping) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

func (t IntegrationField) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

func (t *IntegrationField) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err
}
