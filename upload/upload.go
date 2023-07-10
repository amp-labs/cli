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

// TODO: read from appdata
const customer = "customer_external_identifier"
var now = time.Now()
var year, month, day = now.Year(), int(now.Month()), now.Day()

// TODO: stop harcoding these.
const apiKey = "AIzaSyB1zaLK-0rQebuF5-g-7wt3qwg3WQhQrls"
const bucketName = "ampersand-dev-deploy-uploads"

func Upload(zipPath string) (uploadPath string, err error) {

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	destination := fmt.Sprintf("%s/%d/%02d/%02d/%s", customer, year, month, day, zipPath)

	// Create a new writer object to write the zip file contents to the bucket
	zipObject := client.Bucket(bucketName).Object(destination)
	writer := zipObject.NewWriter(ctx)

	// Write the zip file contents to the bucket using the writer object
	if _, err := writer.Write(zipData); err != nil {
		log.Fatal(err)
		return "", err
	}

	if err := writer.Close(); err != nil {
		log.Fatal(err)
		return "", err
	}
	finalPath := fmt.Sprintf("gs://%s/%s", bucketName, destination)

	return finalPath, nil
}
