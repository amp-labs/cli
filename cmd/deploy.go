package cmd

import (
	"path/filepath"

	"github.com/amp-labs/cli/files"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/storage"
	"github.com/amp-labs/cli/utils"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy <sourceFolderPath>",
	Short: "Deploy amp.yaml file",
	Long:  "Deploy changes to amp.yaml file.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		workingDir := utils.GetWorkingDir()
		folderName := filepath.ToSlash(filepath.Join(workingDir, path))

		zipPath, err := files.Zip(folderName)
		defer files.Remove(zipPath)

		if err != nil {
			logger.FatalErr("Unable to zip folder", err)
		}

		gcsUrl, err := storage.Upload(zipPath)
		if err != nil {
			logger.FatalErr("Unable to upload to Google Cloud Storage", err)
		}
		logger.Debugf("Uploaded to %v", gcsUrl)
		logger.Info("Successfully deployed changes to your integrations ....")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
