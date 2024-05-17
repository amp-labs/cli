package cmd

import (
	"errors"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "delete:integration <integrationId>",
	Short: "Delete integration",
	Long:  "Delete integration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("Deleting Integration")

		integrationId := args[0]
		projectId := flags.GetProjectId()
		if projectId == "" {
			logger.Fatal("Must provide a project ID in the --project flag")
		}

		apiKey := flags.GetAPIKey()

		err := request.NewAPIClient(projectId, &apiKey).
			DeleteIntegration(cmd.Context(), integrationId)
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Clerk session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to delete integration", err)
			}
		}

		logger.Info("Successfully deleted integration.")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
