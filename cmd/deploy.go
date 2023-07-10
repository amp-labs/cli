package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/amp-labs/cli/upload"
	"github.com/amp-labs/cli/zip"
	"github.com/spf13/cobra"
)

var filePath = "/amp"

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy amp.yaml file",
	Long:  `Deploy changes to amp.yaml file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deploy called")

		//Resolve amp's location using current working directory path and folder name
		workingDir, _ := os.Getwd()
		var folderName = filepath.ToSlash(filepath.Join(workingDir, filePath))

		//Zips the amp folder and makes a temporary copy of zip in system temp directory
		zipPath, err := zip.Zip(folderName)
		if err != nil {
			log.Fatal(err)
			return
		}

		if _, err := upload.Upload(zipPath); err != nil {
			log.Fatal(err)
			cleanUp(zipPath)
			return
		}
		cleanUp(zipPath)

		fmt.Println("Successfully deployed changes to amp.yaml....")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func cleanUp(filename string) {
	err := os.Remove(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
}
