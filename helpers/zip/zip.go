package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const errorkey = "ERROR:Ampersand-Cli:cli/helpers/zip: "

var err error
var now = time.Now()

func Zip(folderName string) (string, error) {

	var workingDir, err = os.Getwd()
	if err != nil {
		return errorkey, err
	}

	var zippedFolder = fmt.Sprintf("amp_%d.zip", now.Unix())

	var zippedDir = filepath.ToSlash(filepath.Join(workingDir, zippedFolder))

	if err := zipSource(folderName, zippedDir); err != nil {
		return errorkey, err
	}
	return zippedFolder, nil

}

func zipSource(source, target string) error {
	file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Method = zip.Deflate

		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}
