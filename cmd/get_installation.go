package cmd

import (
	"errors"
	"os"
	"sort"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/amp-labs/cli/utils"
	"github.com/spf13/cobra"
)

type installationActions struct {
	Read      []string `json:"read"`
	Write     []string `json:"write"`
	Subscribe []string `json:"subscribe"`
	Proxy     bool     `json:"proxy"`
}

type installationDetail struct {
	Id            string              `json:"id"`
	IntegrationId string              `json:"integrationId"`
	GroupRef      string              `json:"groupRef"`
	ConnectionId  string              `json:"connectionId"`
	HealthStatus  string              `json:"healthStatus"`
	RevisionId    string              `json:"revisionId"`
	Provider      string              `json:"provider"`
	Actions       installationActions `json:"actions"`
}

func summarizeInstallation(installation *request.Installation) installationDetail {
	detail := installationDetail{
		Id:            installation.Id,
		IntegrationId: installation.IntegrationId,
		GroupRef:      installation.GroupRef,
		ConnectionId:  installation.ConnectionId,
		HealthStatus:  installation.HealthStatus,
		Actions: installationActions{
			Read:      []string{},
			Write:     []string{},
			Subscribe: []string{},
		},
	}

	if installation.Group != nil {
		detail.GroupRef = installation.Group.GroupRef
	}

	if installation.Connection != nil {
		detail.ConnectionId = installation.Connection.Id
	}

	if installation.Config == nil {
		return detail
	}

	detail.RevisionId = installation.Config.RevisionId

	content, ok := installation.Config.Content.(map[string]any)
	if !ok {
		return detail
	}

	detail.Provider, _ = content["provider"].(string)
	detail.Actions.Read = configuredObjects(content, "read")
	detail.Actions.Write = configuredObjects(content, "write")
	detail.Actions.Subscribe = configuredObjects(content, "subscribe")
	detail.Actions.Proxy = proxyEnabled(content)

	return detail
}

func configuredObjects(content map[string]any, action string) []string {
	actionConfig, ok := content[action].(map[string]any)
	if !ok {
		return []string{}
	}

	objects, ok := actionConfig["objects"].(map[string]any)
	if !ok {
		return []string{}
	}

	names := make([]string, 0, len(objects))
	for name := range objects {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func proxyEnabled(content map[string]any) bool {
	proxy, ok := content["proxy"].(map[string]any)
	if !ok {
		return false
	}

	enabled, _ := proxy["enabled"].(bool)

	return enabled
}

var getInstallationCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "get:installation <integrationId> <installationId>",
	Short: "Show an installation",
	Long:  "Show an installation's provider, health, and configured actions without connection credentials.",
	Args:  cobra.ExactArgs(2), //nolint:mnd
	Run: func(cmd *cobra.Command, args []string) {
		projectId := flags.GetProjectOrFail()
		apiKey := flags.GetAPIKey()
		client := request.NewAPIClient(projectId, &apiKey)

		installation, err := client.GetInstallation(cmd.Context(), args[0], args[1])
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to get installation", err)
			}
		}

		err = utils.WriteStruct(os.Stdout, flags.GetOutputFormatForCommand(cmd), summarizeInstallation(installation))
		if err != nil {
			logger.FatalErr("Unable to write installation", err)
		}
	},
}

func init() {
	err := flags.InitAndBindFormatFlag(getInstallationCmd)
	if err != nil {
		logger.FatalErr("unable to initialize flags", err)
	}

	rootCmd.AddCommand(getInstallationCmd)
}
