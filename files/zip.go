package files

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	mode        = 0o644
	yamlName    = "amp.yaml"
	yamlAltName = "amp.yml"
)

func chdir(dir string, function func() error) error {
	origDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %w", err)
	}

	err = os.Chdir(dir)
	if err != nil {
		return fmt.Errorf("error changing directory: %w", err)
	}

	var errs []error

	err = function()
	if err != nil {
		errs = append(errs, err)
	}

	err = os.Chdir(origDir)
	if err != nil {
		errs = append(errs, fmt.Errorf("error changing directory: %w", err))
	}

	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return errors.Join(errs...)
	}

	return nil
}

func getZipDir(source string) (string, error) {
	info, err := os.Stat(source)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: source %q does not exist: %w", ErrBadManifest, source, err)
		} else {
			return "", fmt.Errorf("%w: stat %q: %w", ErrBadManifest, source, err)
		}
	}

	if info == nil {
		return "", fmt.Errorf("%w: source %q does not exist: %w", ErrBadManifest, source, err)
	}

	var sourceDir string
	if info.IsDir() {
		sourceDir = source
	} else {
		name := filepath.Base(source)
		if name != yamlName && name != yamlAltName {
			return "", fmt.Errorf("%w: source %q is not a directory nor an %q file",
				ErrBadManifest, source, yamlName)
		}

		sourceDir = filepath.Dir(source)
	}

	return sourceDir, nil
}

func statYaml() (os.FileInfo, error) {
	yamlStat, err := os.Stat(yamlName)
	if err != nil { //nolint:nestif
		if os.IsNotExist(err) {
			yamlStat, err = os.Stat(yamlAltName)
			if err != nil {
				if os.IsNotExist(err) {
					return nil, fmt.Errorf("%s does not exist: %w", yamlName, err)
				} else {
					return nil, fmt.Errorf("stat %q: %w", yamlAltName, err)
				}
			}
		} else {
			return nil, fmt.Errorf("stat %q: %w", yamlName, err)
		}
	}

	return yamlStat, nil
}

func importYaml(writer *zip.Writer) error {
	yamlStat, err := statYaml()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(yamlStat)
	if err != nil {
		return fmt.Errorf("error adding file header while zipping: %w", err)
	}

	header.Method = zip.Deflate
	header.SetMode(mode)
	header.Name = yamlName

	headerWriter, err := writer.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("error adding file header while zipping: %w", err)
	}

	contents, err := os.ReadFile(yamlStat.Name())
	if err != nil {
		return fmt.Errorf("error opening %s file while zipping: %w", yamlStat.Name(), err)
	}

	manifest, err := ParseManifest(contents)
	if err != nil {
		return fmt.Errorf("error parsing manifest file %q: %w", yamlStat.Name(), err)
	}

	if err := ValidateManifest(manifest); err != nil {
		return fmt.Errorf("error validating manifest file %q: %w", yamlStat.Name(), err)
	}

	_, err = io.Copy(headerWriter, bytes.NewReader(contents))
	if err != nil {
		return fmt.Errorf("error copying file for zipping: %w", err)
	}

	return nil
}

// Zip creates a zip archive of the given directory in-memory.
func Zip(source string) ([]byte, error) { // nolint:funlen,cyclop
	sourceDir, err := getZipDir(source)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer

	chdirErr := chdir(sourceDir, func() error {
		writer := zip.NewWriter(&out)

		if err := importYaml(writer); err != nil {
			return err
		}

		if err := writer.Close(); err != nil {
			return fmt.Errorf("error closing zip writer: %w", err)
		}

		return nil
	})
	if chdirErr != nil {
		return nil, chdirErr
	}

	return out.Bytes(), nil
}
