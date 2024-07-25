package cmd

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/spf13/cobra"
)

var listIntegrationsCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "list:integrations",
	Short: "List integrations",
	Long:  "List integrations",
	Run: func(cmd *cobra.Command, args []string) {
		projectId := flags.GetProjectOrFail()
		apiKey := flags.GetAPIKey()

		client := request.NewAPIClient(projectId, &apiKey)

		integs, err := client.ListIntegrations(cmd.Context())
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to list installations", err)
			}
		}

		sort.Slice(integs, func(i, j int) bool {
			return integs[i].Name < integs[j].Name
		})

		for _, integ := range integs {
			num := numInstallations(cmd.Context(), client, integ.Id)

			logger.Info(integ.Id + " " + integ.Name + " (" + num + ")")
		}
	},
}

func numInstallations(ctx context.Context, client *request.APIClient, id string) string {
	insts, err := client.ListInstallations(ctx, id)
	if err != nil {
		logger.FatalErr("Unable to list installations", err)
	}

	num := len(insts)

	switch num {
	case 0:
		return "no installations"
	case 1:
		return "1 installation"
	default:
		return fmt.Sprintf("%d installations", num)
	}
}

func init() {
	rootCmd.AddCommand(listIntegrationsCmd)
}
