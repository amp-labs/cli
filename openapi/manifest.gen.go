// Package openapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package openapi

import (
	"encoding/json"

	"github.com/oapi-codegen/runtime"
)

// Defines values for DeliveryMode.
const (
	Auto      DeliveryMode = "auto"
	OnRequest DeliveryMode = "onRequest"
)

// Defines values for OptionalFieldsAutoOption.
const (
	All OptionalFieldsAutoOption = "all"
)

// Backfill defines model for Backfill.
type Backfill struct {
	DefaultPeriod DefaultPeriod `json:"defaultPeriod"`
}

// DefaultPeriod defines model for DefaultPeriod.
type DefaultPeriod struct {
	// Days Number of days in past to backfill from. 0 is no backfill. e.g) if 10, then backfill last 10 days of data. Required if fullHistory is not set.
	Days *int `json:"days,omitempty" validate:"required_without=FullHistory,omitempty,min=0"`

	// FullHistory If true, backfill all history. Required if days is not set.
	FullHistory *bool `json:"fullHistory,omitempty" validate:"required_without=Days"`
}

// Delivery defines model for Delivery.
type Delivery struct {
	// Mode The data delivery mode for this object. If not specified, defaults to automatic.
	Mode *DeliveryMode `json:"mode,omitempty"`

	// PageSize The number of records to receive per data delivery.
	PageSize *int `json:"pageSize,omitempty"`
}

// DeliveryMode The data delivery mode for this object. If not specified, defaults to automatic.
type DeliveryMode string

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
	// AllFields This is a list of all fields on the object for a particular SaaS instance. This is used to populate the UI during configuration.
	AllFields   *[]HydratedIntegrationField `json:"allFields,omitempty"`
	Backfill    *Backfill                   `json:"backfill,omitempty"`
	Destination string                      `json:"destination"`
	DisplayName string                      `json:"displayName"`

	// MapToDisplayName A display name to map to in the destination.
	MapToDisplayName string `json:"mapToDisplayName,omitempty"`

	// MapToName An object name to map to in the destination.
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
}

// Integration defines model for Integration.
type Integration struct {
	DisplayName *string           `json:"displayName,omitempty"`
	Name        string            `json:"name"`
	Provider    string            `json:"provider"`
	Proxy       *IntegrationProxy `json:"proxy,omitempty"`
	Read        *IntegrationRead  `json:"read,omitempty"`
	Write       *IntegrationWrite `json:"write,omitempty"`
}

// IntegrationField defines model for IntegrationField.
type IntegrationField struct {
	union json.RawMessage
}

// IntegrationFieldExistent defines model for IntegrationFieldExistent.
type IntegrationFieldExistent struct {
	FieldName string `json:"fieldName"`

	// MapToDisplayName The display name to map to in the destination.
	MapToDisplayName string `json:"mapToDisplayName,omitempty"`

	// MapToName The field name to map to in the destination.
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

	// MapToDisplayName A display name to map to in the destination.
	MapToDisplayName string `json:"mapToDisplayName,omitempty"`

	// MapToName An object name to map to in the destination.
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

// IntegrationWrite defines model for IntegrationWrite.
type IntegrationWrite struct {
	Objects *[]IntegrationWriteObject `json:"objects,omitempty"`
}

// IntegrationWriteObject defines model for IntegrationWriteObject.
type IntegrationWriteObject struct {
	// InheritMappingFromRead If true, the write object will inherit the mapping from the read object. If false, the write object will have no mapping.
	InheritMappingFromRead *bool  `json:"inheritMappingFromRead,omitempty"`
	ObjectName             string `json:"objectName"`
}

// Manifest This is the schema of the manifest file that is used to define the integrations of the project.
type Manifest struct {
	Integrations []Integration `json:"integrations"`

	// SpecVersion The version of the manifest spec that this file conforms to.
	SpecVersion string `json:"specVersion"`
}

// OptionalFieldsAutoOption defines model for OptionalFieldsAutoOption.
type OptionalFieldsAutoOption string

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
