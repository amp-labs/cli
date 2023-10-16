package files

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const mode = 420

// Zip creates a zip archive of the given directory in-memory.
func Zip(sourceDir string) ([]byte, error) { // nolint:funlen
	var out bytes.Buffer
	writer := zip.NewWriter(&out)

	err := filepath.WalkDir(sourceDir, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error zipping: %w", err)
		}

		info, err := dir.Info()
		if err != nil {
			return fmt.Errorf("error getting file info while zipping: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("error adding file header while zipping: %w", err)
		}

		header.Method = zip.Deflate
		header.SetMode(mode)

		header.Name, err = filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("error adding file header while zipping: %w", err)
		}

		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("error adding file header while zipping: %w", err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening file while zipping: %w", err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
			}
		}(file)

		_, err = io.Copy(headerWriter, file)
		if err != nil {
			return fmt.Errorf("error copying file for zipping: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to zip directory: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error closing zip writer: %w", err)
	}

	return out.Bytes(), nil
}
