package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// TODO: read from appdata
const customer = "TODO_PREFIX"

var now = time.Now()
var year, month, day = now.Year(), int(now.Month()), now.Day()

// TODO: this should get moved to the server instead, so API keys & bucket names
// get managed there.
const apiKey = "AIzaSyB1zaLK-0rQebuF5-g-7wt3qwg3WQhQrls"
const bucketName = "ampersand-dev-deploy-uploads"

func Upload(zipPath string) (gcsUrl string, err error) {

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("error initializing GCS client: %v", err)
	}

	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		return "", fmt.Errorf("error reading zipped file: %v", err)
	}

	destination := fmt.Sprintf("%s/%d/%02d/%02d/%s", customer, year, month, day, filepath.Base(zipPath))

	// Create a new writer object to write the zip file contents to the bucket
	zipObject := client.Bucket(bucketName).Object(destination)
	writer := zipObject.NewWriter(ctx)

	// Write the zip file contents to the bucket using the writer object
	if _, err := writer.Write(zipData); err != nil {
		return "", fmt.Errorf("error writing to GCS bucket: %v", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("error closing writer: %v", err)
	}
	finalPath := fmt.Sprintf("gs://%s/%s", bucketName, destination)

	return finalPath, nil
}
