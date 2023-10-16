package cmd

import (
	"path/filepath"
	"strings"

	"github.com/amp-labs/cli/files"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/amp-labs/cli/storage"
	"github.com/amp-labs/cli/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deployCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "deploy <sourceFolderPath>",
	Short: "Deploy amp.yaml file",
	Long:  "Deploy changes to amp.yaml file.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectId := flags.GetProjectId()
		if projectId == "" {
			logger.Fatal("Must provide a project ID in the --project flag")

			return
		}

		path := args[0]
		workingDir := utils.GetWorkingDir()
		if workingDir == "" {
			logger.Fatal("Unable to get working directory")

			return
		}

		folderName := filepath.Join(workingDir, path)

		zipPath, err := files.Zip(folderName)
		defer files.Remove(zipPath)

		if err != nil {
			logger.FatalErr("Unable to zip folder", err)

			return
		}

		gcsURL, err := storage.Upload(zipPath)
		if err != nil {
			logger.FatalErr("Unable to upload to Google Cloud Storage", err)

			return
		}
		logger.Debugf("Uploaded to %v", gcsURL)

		apiKey := viper.GetString("key")
		if apiKey == "" {
			logger.Fatal("Must provide an API key in the --key flag")

			return
		}

		integrations, err := request.NewAPIClient(projectId, &apiKey).
			BatchUpsertIntegrations(cmd.Context(), request.BatchUpsertIntegrationsParams{SourceZipURL: gcsURL})
		if err != nil {
			logger.FatalErr("Unable to deploy integrations", err)

			return
		}

		names := make([]string, len(integrations))
		for idx, i := range integrations {
			names[idx] = i.Name
		}

		if len(names) == 0 {
			logger.Infof("No integrations were found in the source file.\n")
		} else if len(names) == 1 {
			logger.Infof("Successfully deployed your integration %s.\n", names[0])
		} else {
			logger.Infof("Successfully deployed your integrations %s.\n", strings.Join(names, ", "))
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("key", "k", "", "Ampersand API key")

	if err := viper.BindPFlag("key", deployCmd.Flags().Lookup("key")); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("key", "AMP_API_KEY"); err != nil {
		panic(err)
	}
}
