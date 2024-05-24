package request

import (
	"time"

	"github.com/amp-labs/cli/openapi"
)

type Installation struct {
	Id            string      `json:"id"                     mapstructure:"Id"`
	ProjectId     string      `json:"projectId"              mapstructure:"ProjectId"`
	IntegrationId string      `json:"integrationId"          mapstructure:"IntegrationId"`
	GroupRef      string      `json:"groupRef,omitempty"     mapstructure:"GroupRef"`
	Group         *Group      `json:"group"                  mapstructure:"Group"`
	ConnectionId  string      `json:"connectionId,omitempty" mapstructure:"ConnectionId"`
	Connection    *Connection `json:"connection"             mapstructure:"Connection"`
	CreatedBy     string      `json:"createdBy"              mapstructure:"CreatedBy"`
	Config        *Config     `json:"config"                 mapstructure:"Config"`
	CreateTime    time.Time   `json:"createTime"             mapstructure:"CreateTime"`
	HealthStatus  string      `json:"healthStatus"           mapstructure:"HealthStatus"`
}

type Group struct {
	GroupRef   string    `json:"groupRef"   mapstructure:"GroupRef"   validate:"required"`
	GroupName  string    `json:"groupName"  mapstructure:"GroupName"  validate:"required"`
	ProjectId  string    `json:"projectId"  mapstructure:"ProjectId"`
	CreateTime time.Time `json:"createTime" mapstructure:"CreateTime"`
	UpdateTime time.Time `json:"updateTime" mapstructure:"UpdateTime"`
}

type Connection struct {
	Id                   string             `json:"id"                    mapstructure:"Id"`
	ProjectId            string             `json:"projectId"             mapstructure:"ProjectId"`
	ProviderApp          *ProviderApp       `json:"providerApp,omitempty" mapstructure:"ProviderApp"`
	Group                *Group             `json:"group"                 mapstructure:"Group"`
	Consumer             *Consumer          `json:"consumer"              mapstructure:"Consumer"`
	ProviderWorkspaceRef string             `json:"providerWorkspaceRef"  mapstructure:"ProviderWorkspaceRef"`
	ProviderConsumerRef  string             `json:"providerConsumerRef"   mapstructure:"ProviderConsumerRef"`
	CreateTime           time.Time          `json:"createTime"            mapstructure:"CreateTime"`
	UpdateTime           time.Time          `json:"updateTime"            mapstructure:"UpdateTime"`
	Scopes               []string           `json:"scopes"                mapstructure:"Scopes"`
	Status               string             `json:"status"                mapstructure:"Status"`
	CatalogVars          *map[string]string `json:"catalogVars,omitempty" mapstructure:"CatalogVars"`
}

type ProviderApp struct {
	Id           string    `json:"id"                     mapstructure:"Id"`
	CreateTime   time.Time `json:"createTime"             mapstructure:"CreateTime"`
	UpdateTime   time.Time `json:"updateTime"             mapstructure:"UpdateTime"`
	ExternalRef  string    `json:"externalRef"            mapstructure:"ExternalRef"`
	Provider     string    `json:"provider"               mapstructure:"Provider"     validate:"required"`
	ClientId     string    `json:"clientId"               mapstructure:"ClientId"     validate:"required"`
	ClientSecret string    `json:"clientSecret,omitempty" mapstructure:"ClientSecret" validate:"required"`
	Scopes       []string  `json:"scopes"                 mapstructure:"Scopes"`
	ProjectId    string    `json:"projectId"              mapstructure:"ProjectId"`
}

type Config struct {
	Id             string    `json:"id"             mapstructure:"Id"`
	RevisionId     string    `json:"revisionId"     mapstructure:"RevisionId"`
	CreatedBy      string    `json:"createdBy"      mapstructure:"CreatedBy"      validate:"required"`
	Content        any       `json:"content"        mapstructure:"Content"        validate:"required"`
	InstallationId string    `json:"installationId" mapstructure:"InstallationId"`
	CreateTime     time.Time `json:"createTime"     mapstructure:"CreateTime"`
}

type Consumer struct {
	ConsumerRef  string    `json:"consumerRef"  mapstructure:"ConsumerRef"  validate:"required"`
	ConsumerName string    `json:"consumerName" mapstructure:"ConsumerName" validate:"required"`
	ProjectId    string    `json:"projectId"    mapstructure:"ProjectId"`
	CreateTime   time.Time `json:"createTime"   mapstructure:"CreateTime"`
	UpdateTime   time.Time `json:"updateTime"   mapstructure:"UpdateTime"`
}

type Integration struct {
	Id             string    `json:"id"`
	Provider       string    `json:"provider"       validate:"required"`
	ProjectId      string    `json:"projectId"`
	Name           string    `json:"name"           validate:"required"`
	CreateTime     time.Time `json:"createTime"`
	UpdateTime     time.Time `json:"updateTime"`
	LatestRevision *Revision `json:"latestRevision"`
}

type Revision struct {
	Id            string              `json:"id"`
	IntegrationId string              `json:"integrationId"`
	CreateTime    time.Time           `json:"createTime"`
	Content       openapi.Integration `json:"content"`
	SpecVersion   string              `json:"specVersion"`
}
