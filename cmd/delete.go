package cmd

import (
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deleteCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "integration:delete <integrationName>",
	Short: "Delete an integration",
	Long:  "Delete an integration.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectId := flags.GetProjectId()
		if projectId == "" {
			logger.Fatal("Must provide a project ID in the --project flag")
		}

		logger.Infof("Uploaded to %v", projectId)

		apiKey := viper.GetString("key")
		if apiKey == "" {
			logger.Fatal("Must provide an API key in the --key flag or via the AMP_API_KEY environment variable")
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	// deleteCmd.Flags().StringP("key", "k", "", "Ampersand API key")

	// if err := viper.BindPFlag("key", deployCmd.Flags().Lookup("key")); err != nil {
	// 	panic(err)
	// }

	// if err := viper.BindEnv("key", "AMP_API_KEY"); err != nil {
	// 	panic(err)
	// }
}
