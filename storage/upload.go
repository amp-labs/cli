package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/amp-labs/cli/vars"
	"google.golang.org/api/option"
)

var (
	now              = time.Now()                              //nolint:gochecknoglobals
	year, month, day = now.Year(), int(now.Month()), now.Day() //nolint:gochecknoglobals
)

// TODO: this should get moved to the server instead, so API keys & bucket names
// get managed there.
var (
	apiKey     = vars.GCSKey    //nolint:gochecknoglobals
	bucketName = vars.GCSBucket //nolint:gochecknoglobals
)

// Upload takes in a byte array and uploads it to GCS as per the given name
func Upload(data []byte, as string) (gcsURL string, err error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("error initializing GCS client: %w", err)
	}

	destinationPath := fmt.Sprintf("%d/%02d/%02d/%s", year, month, day, filepath.Base(as))

	// Create a new writer object to write the zip file contents to the bucket
	zipObject := client.Bucket(bucketName).Object(destinationPath)
	writer := zipObject.NewWriter(ctx)

	// Write the zip file contents to the bucket using the writer object
	if _, err := writer.Write(data); err != nil {
		return "", fmt.Errorf("error writing to GCS bucket: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("error closing writer: %w", err)
	}

	finalPath := fmt.Sprintf("gs://%s/%s", bucketName, destinationPath)

	return finalPath, nil
}
