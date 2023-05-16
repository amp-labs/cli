package cmd

import (
  "fmt"
  "log"
  "os"
  "path/filepath"
  
  "github.com/amp-labs/cli/helpers/upload"
  //"github.com/amp-labs/cli/helpers/util"
  "github.com/amp-labs/cli/helpers/zip"
  "github.com/spf13/cobra"
)

var err error
var filePath = "/amp"

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
  Use:   "deploy",
  Short: "Deploy amp.yaml file",
  Long:  `Deploy changes to amp.yaml file.`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("deploy called")
    fmt.Println("Creating a local copy of the AMP file")
    
    //Resolve amp's location using current working directory path and folder name
    workingDir,_  := os.Getwd()
    var folderName = filepath.ToSlash(filepath.Join(workingDir,filePath))
    

    //Zips the amp folder and makes a temporary copy of zip in system temp directory
    fmt.Println("zipping FOLDER name:",folderName)
    zipPath,err := zip.Zip(folderName)
    if err != nil {
      fmt.Println("Exiting following unsuccesful zip operation")
      return 
    }

    fmt.Println("Uploading zipped file to storage bucket")
    uploadPath,err := upload.Upload(zipPath)
    
    clean_up(zipPath)
    
    if err!=nil{
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
