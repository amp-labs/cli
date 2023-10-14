package files

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/utils"
)

var now = time.Now() //nolint:gochecknoglobals

const (
	TmpDir = ".amp" // This is used for zipping and uploading
)

func Zip(path string) (zippedFolder string, zipError error) {
	wd := utils.GetWorkingDir()
	if wd == "" {
		logger.Fatal("unable to get working directory")

		return
	}

	// TODO: create in temporary folder, not cwd
	dest := filepath.ToSlash(filepath.Join(wd, TmpDir, fmt.Sprintf("amp_%d.zip", now.Unix())))

	if err := zipSource(path, dest); err != nil {
		return "", err
	}

	return dest, nil
}

func zipSource(source string, dest string) error {
	file, err := ensureDirectoryAndCreateFile(dest)
	if err != nil {
		return fmt.Errorf("cannot create destination for zipping: %w", err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	return filepath.WalkDir(source, func(path string, dir fs.DirEntry, err error) error {
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

		header.Name, err = filepath.Rel(source, path)

		if err != nil {
			return fmt.Errorf("error adding file header while zipping: %w", err)
		}

		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("error adding file header while zipping: %w", err)
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening file for zipping: %w", err)
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		if err != nil {
			return fmt.Errorf("error copying file for zipping: %w", err)
		}

		return nil
	})
}

// ensureDirectoryAndCreateFile creates the directory for the file if it doesn't exist
// and then returns the created file.
func ensureDirectoryAndCreateFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("cannot create destination dir for zipping: %w", err)
	}

	return os.Create(path)
}
