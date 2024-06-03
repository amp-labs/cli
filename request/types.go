package request

import (
	"time"

	"github.com/amp-labs/cli/openapi"
)

type Installation struct {
	Id            string      `json:"id"`
	ProjectId     string      `json:"projectId"`
	IntegrationId string      `json:"integrationId"`
	GroupRef      string      `json:"groupRef,omitempty"`
	Group         *Group      `json:"group"`
	ConnectionId  string      `json:"connectionId,omitempty"`
	Connection    *Connection `json:"connection"`
	CreatedBy     string      `json:"createdBy"`
	Config        *Config     `json:"config"`
	CreateTime    time.Time   `json:"createTime"`
	HealthStatus  string      `json:"healthStatus"`
}

type Group struct {
	GroupRef   string    `json:"groupRef"`
	GroupName  string    `json:"groupName"`
	ProjectId  string    `json:"projectId"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

type Connection struct {
	Id                   string             `json:"id"`
	ProjectId            string             `json:"projectId"`
	ProviderApp          *ProviderApp       `json:"providerApp,omitempty"`
	Group                *Group             `json:"group"`
	Consumer             *Consumer          `json:"consumer"`
	ProviderWorkspaceRef string             `json:"providerWorkspaceRef"`
	ProviderConsumerRef  string             `json:"providerConsumerRef"`
	CreateTime           time.Time          `json:"createTime"`
	UpdateTime           time.Time          `json:"updateTime"`
	Scopes               []string           `json:"scopes"`
	Status               string             `json:"status"`
	CatalogVars          *map[string]string `json:"catalogVars,omitempty"`
}

type ProviderApp struct {
	Id           string    `json:"id"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime   time.Time `json:"updateTime"`
	ExternalRef  string    `json:"externalRef"`
	Provider     string    `json:"provider"`
	ClientId     string    `json:"clientId"`
	ClientSecret string    `json:"clientSecret,omitempty"`
	Scopes       []string  `json:"scopes"`
	ProjectId    string    `json:"projectId"`
}

type Config struct {
	Id             string    `json:"id"`
	RevisionId     string    `json:"revisionId"`
	CreatedBy      string    `json:"createdBy"`
	Content        any       `json:"content"`
	InstallationId string    `json:"installationId"`
	CreateTime     time.Time `json:"createTime"`
}

type Consumer struct {
	ConsumerRef  string    `json:"consumerRef"`
	ConsumerName string    `json:"consumerName"`
	ProjectId    string    `json:"projectId"`
	CreateTime   time.Time `json:"createTime"`
	UpdateTime   time.Time `json:"updateTime"`
}

type Integration struct {
	Id             string    `json:"id"`
	Provider       string    `json:"provider"`
	ProjectId      string    `json:"projectId"`
	Name           string    `json:"name"`
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

type Project struct {
	Id         string    `json:"id"`
	AppName    string    `json:"appName"`
	Name       string    `json:"name"`
	CreateTime time.Time `json:"createTime"`
	OrgId      string    `json:"orgId"`
}
