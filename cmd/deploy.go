package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/amp-labs/cli/cmd/helpers/upload"
	"github.com/amp-labs/cli/cmd/helpers/util"
	"github.com/amp-labs/cli/cmd/helpers/zip"
	"github.com/spf13/cobra"
)

var now = time.Now()
var uploadPath string
var uploadErrors error
var filePath = "cmd/amp/amp.yaml"

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy amp.yaml file",
	Long:  `Deploy changes to amp.yaml file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deploy called")
		fmt.Println("Creating a local copy of the AMP file")

		//Resolve amp.yaml's location using current working directory path and constant directory extension
		workingDir,_  := os.Getwd()
		var filename = filepath.ToSlash(filepath.Join(workingDir,filePath))

		//make a local copy of the amp.yaml by attaching a timestamp value to its name

		var localVersion = fmt.Sprintf("amp_%d.yaml", now.Unix())
		_copied := util.Copy(filename,localVersion)

		if _copied{
			fmt.Println("zipping file name:",localVersion)
			
			var zipPath = fmt.Sprintf("amp_%d.zip", now.Unix())
			
			_zipped := zip.Zip(localVersion, zipPath)
			
			if _zipped{
				fmt.Println("Uploading zipped file to storage bucket")
				uploadPath,uploadErrors = upload.Upload(zipPath)
			}else {
				fmt.Println("Exiting following unsuccesful zip operation")
				
				return
			}
			fmt.Println("Deleting local amp zipped file")
			
			clean_up(zipPath)
		}
		
		clean_up(localVersion)

		if uploadErrors!=nil{
			fmt.Println("Exiting following unsuccesful upload operation")
		}else{
			fmt.Println("Succesfully uploaded zipped file to remote storage..",uploadPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func clean_up(filename string){
	err := os.Remove(filename)
	if err != nil{
		log.Fatal(err)
		return
	}
}
