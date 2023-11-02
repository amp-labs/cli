package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrGcsFailure = errors.New("error uploading to GCS")

// Upload takes in bytes and uploads it to GCS as per the given name.
func Upload(ctx context.Context, data []byte, url, md5 string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	if len(md5) > 0 {
		req.Header.Set("Content-MD5", md5)
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}

	defer rsp.Body.Close()

	if rsp.StatusCode >= 200 && rsp.StatusCode < 300 {
		return nil
	} else {
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}

		return fmt.Errorf("%w: %s", ErrGcsFailure, string(body))
	}
}
