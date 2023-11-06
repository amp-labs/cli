package cmd

import (
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "delete-integration <integrationId>",
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
		if apiKey == "" {
			logger.Fatal("Must provide an API key in the --key flag or via the AMP_API_KEY environment variable")
		}

		err := request.NewAPIClient(projectId, &apiKey).
			DeleteIntegration(cmd.Context(), integrationId)
		if err != nil {
			logger.FatalErr("Unable to delete integration", err)
		}

		logger.Info("Successfully deleted integration.")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
