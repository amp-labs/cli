package cmd

import (
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/spf13/cobra"
)

var deleteInstallationCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "delete:installation <integrationId> <installationId>",
	Short: "Delete installation",
	Long:  "Delete installation",
	Args:  cobra.ExactArgs(2), //nolint:gomnd
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("Deleting installation")

		integrationId := args[0]
		installationId := args[1]
		projectId := flags.GetProjectId()
		if projectId == "" {
			logger.Fatal("Must provide a project ID in the --project flag")
		}

		apiKey := flags.GetAPIKey()

		err := request.NewAPIClient(projectId, &apiKey).
			DeleteInstallation(cmd.Context(), integrationId, installationId)
		if err != nil {
			logger.FatalErr("Unable to delete installation", err)
		}

		logger.Info("Successfully deleted installation.")
	},
}

func init() {
	rootCmd.AddCommand(deleteInstallationCmd)
}
