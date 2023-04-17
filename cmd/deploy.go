package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"context"
	"io/ioutil"
	"log"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var customer = "customer_external_identifier"
var now = time.Now()
var year = now.Year()
var month = int(now.Month())
var day = now.Day()

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy amp.yaml file",
	Long:  `Deploy changes to amp.yaml file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deploy called")
		fmt.Println("Zipping amp.yaml file........")
		zzip()

		fmt.Println("sending file to remote data store......")
		upload()

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

func zzip() {
	fmt.Println("Beginning to zip amp.yaml.....")

	// Open the file to be zipped
	file, err := os.Open("cmd/amp/amp.yaml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create the output zip file
	zipFile, err := os.Create("cmd/amp/amp.zip")
	if err != nil {
		panic(err)
	}
	defer zipFile.Close()

	// Create a new zip archive
	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	// Add the file to the zip archive
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		panic(err)
	}

	header.Name = file.Name()

	writer, err := archive.CreateHeader(header)
	if err != nil {
		panic(err)
	}

	if _, err = io.Copy(writer, file); err != nil {
		panic(err)
	}

	fmt.Println("amp.yaml File zipped successfully!.....")
}

func upload() {
	ctx := context.Background()
	apiKey := "AIzaSyB1zaLK-0rQebuF5-g-7wt3qwg3WQhQrls"
	client, err := storage.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}

	bucketName := "ampersand-dev-deploy-uploads"
	zipFile := "cmd/amp/amp.zip"

	zipData, err := ioutil.ReadFile(zipFile)
	if err != nil {
		log.Fatal(err)
	}

	destination := fmt.Sprintf("%s/%d/%02d/%02d", customer, year, month, day)

	// Create a new writer object to write the zip file contents to the bucket
	zipObject := client.Bucket(bucketName).Object(destination)
	writer := zipObject.NewWriter(ctx)

	// Write the zip file contents to the bucket using the writer object
	if _, err := writer.Write(zipData); err != nil {
		log.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		log.Fatal(err)
	}

	// Print a success message if the zip file was uploaded successfully
	fmt.Println("Zip file uploaded successfully!")
}
