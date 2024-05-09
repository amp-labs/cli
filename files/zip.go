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
	mode        = 420
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

// Zip creates a zip archive of the given directory in-memory.
func Zip(source string) ([]byte, error) { // nolint:funlen,cyclop
	info, err := os.Stat(source)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source %q does not exist: %w", source, err)
		} else {
			return nil, fmt.Errorf("stat %q: %w", source, err)
		}
	}

	if info == nil {
		return nil, fmt.Errorf("source %q does not exist: %w", source, err)
	}

	var sourceDir string
	if info.IsDir() {
		sourceDir = source
	} else {
		name := filepath.Base(source)
		if name != yamlName && name != yamlAltName {
			return nil, fmt.Errorf("source %q is not a directory nor an %q file", source, yamlName)
		}

		sourceDir = filepath.Dir(source)
	}

	var out bytes.Buffer

	chdirErr := chdir(sourceDir, func() error {
		yamlStat, err := os.Stat(yamlName)
		if err != nil {
			if os.IsNotExist(err) {
				yamlStat, err = os.Stat(yamlAltName)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("%s does not exist in %q: %w", yamlName, sourceDir, err)
					} else {
						return fmt.Errorf("stat %q: %w", yamlAltName, err)
					}
				}
			} else {
				return fmt.Errorf("stat %q: %w", yamlName, err)
			}
		}

		writer := zip.NewWriter(&out)

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
			if err != nil {
				return fmt.Errorf("error opening %s file while zipping: %w", yamlStat.Name(), err)
			}
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
