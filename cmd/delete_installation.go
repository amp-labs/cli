package cmd

import (
	"errors"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/spf13/cobra"
)

var deleteInstallationCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "delete:installation <integrationId> <installationId>",
	Short: "Delete installation",
	Long:  "Delete installation",
	Args:  cobra.ExactArgs(2), //nolint:mnd
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
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to delete installation", err)
			}
		}

		logger.Info("Successfully deleted installation.")
	},
}

func init() {
	rootCmd.AddCommand(deleteInstallationCmd)
}