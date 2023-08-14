package files

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/amp-labs/cli/utils"
)

var now = time.Now()

func Zip(path string) (zippedFolder string, zipError error) {
	wd := utils.GetWorkingDir()

	// TODO: create in temporary folder, not cwd
	dest := filepath.ToSlash(filepath.Join(wd, fmt.Sprintf("amp_%d.zip", now.Unix())))

	if err := zipSource(path, dest); err != nil {
		return "", err
	}
	return dest, nil

}

func zipSource(source string, dest string) error {
	file, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("cannot create destination for zipping: %v", err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error zipping: %v", err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("error adding file header while zipping: %v", err)
		}

		header.Method = zip.Deflate

		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return fmt.Errorf("error adding file header while zipping: %v", err)
		}
		if info.IsDir() {
			header.Name += "/"
		}

		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("error adding file header while zipping: %v", err)
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening file for zipping: %v", err)
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		if err != nil {
			return fmt.Errorf("error copying file for zipping: %v", err)
		}
		return nil
	})
}
