package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/amp-labs/cli/vars"
	"sigs.k8s.io/yaml"
)

type Format string

const (
	Unknown Format = ""
	JSON    Format = "json"
	YAML    Format = "yaml"
)

var (
	ErrReadFile      = errors.New("unable to read file")
	ErrUnknownFormat = errors.New("unknown format")
)

func GetWorkingDir() string {
	workingDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	return workingDir
}

func ReadStruct(r io.Reader, out any) (Format, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Unknown, err
	}

	if err := json.Unmarshal(data, out); err != nil {
		var se *json.SyntaxError
		if !errors.As(err, &se) {
			return Unknown, err
		}
	} else {
		return JSON, nil
	}

	if err := yaml.Unmarshal(data, out); err != nil {
		return Unknown, err
	}

	return YAML, nil
}

func ReadStructFromFile(filePath string, out any) (Format, error) {
	if filePath == "-" {
		return ReadStruct(os.Stdin, out)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return Unknown, err
	}

	defer func() {
		_ = f.Close()
	}()

	if format, err := ReadStruct(f, out); err == nil {
		return format, nil
	}

	return Unknown,
		fmt.Errorf("%w: failed to unmarshal %s (not valid JSON or YAML)",
			ErrReadFile, filePath)
}

func WriteStruct(writer io.Writer, format Format, data any) error {
	switch format { //nolint:exhaustive
	case JSON:
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)

		if err := enc.Encode(data); err != nil {
			return err
		}

		return nil
	case YAML:
		bts, err := yaml.Marshal(data)
		if err != nil {
			return err
		}

		_, err = writer.Write(bts)

		return err
	default:
		return ErrUnknownFormat
	}
}

func GetStage() string {
	stage, ok := os.LookupEnv("AMP_STAGE_OVERRIDE")
	if ok {
		return stage
	}

	return vars.Stage
}

func WriteStructToFile(filePath string, format Format, data any) error {
	if filePath == "-" {
		return WriteStruct(os.Stdout, format, data)
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	return WriteStruct(f, format, data)
}
