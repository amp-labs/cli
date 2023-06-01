package upload

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// TEMP placement ; TODO: Add a parsing logic to read customer identifier from yaml file
var customer = "customer_external_identifier"
var now = time.Now()
var year = now.Year()
var month = int(now.Month())
var day = now.Day()

// TEMP placement ; should move to a .env file or keymanager
const apiKey = "AIzaSyB1zaLK-0rQebuF5-g-7wt3qwg3WQhQrls"

var bucketName = "ampersand-dev-deploy-uploads"

const errorkey = "ERROR:Ampersand-Cli:cli/upload: "

func Upload(zipPath string) error {

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(errorkey, err)
		return err
	}

	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		log.Fatal(errorkey, err)
		return err
	}

	destination := fmt.Sprintf("%s/%d/%02d/%02d/%s", customer, year, month, day, zipPath)

	// Create a new writer object to write the zip file contents to the bucket
	zipObject := client.Bucket(bucketName).Object(destination)
	writer := zipObject.NewWriter(ctx)

	// Write the zip file contents to the bucket using the writer object
	if _, err := writer.Write(zipData); err != nil {
		log.Fatal(errorkey, err)
		return err
	}

	if err := writer.Close(); err != nil {
		log.Fatal(errorkey, err)
		return err
	}

	// Print a success message if the zip file was uploaded successfully
	// Leaving the path name as a comment in case this needs to be returned to deploy module to build dependencies
	//finalPath := fmt.Sprintf("gs://%s/%s", bucketName, destination)

	return nil
}
