package appdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/adrg/xdg"
	"github.com/imdario/mergo"
)

const fileName = "Ampersand/config.json"

// Config is the user's persisted CLI configuration.
//
// IMPORTANT: Do not modify the JSON labels in this struct without ensuring backwards
// compatibility, since those strings are written to the user's config file on their computer.
type Config struct {
	// Region is the Ampersand deployment region the CLI talks to, e.g. "us" or "eu".
	// An empty value means no region has been selected and the default ("us") applies.
	Region string `json:"region"`
}

// Get returns the user's existing config, or an empty config if the file doesn't exist.
func Get() (Config, error) {
	path, err := configFilePath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// no config file exists yet, which is not an error:
			// the caller is returned an empty Config object, and Set() will create the file
			return Config{}, nil
		}

		return Config{}, fmt.Errorf("can't read config file at %s: %w", path, err)
	}

	var c Config

	err = json.Unmarshal(data, &c)
	if err != nil {
		return Config{}, fmt.Errorf("can't parse config at %s: %w", path, err)
	}

	return c, nil
}

// Set writes the given config to the user's config file.
// The config can be a partial config, and it is merged with the existing config.
func Set(config Config) error {
	existing, err := Get()
	if err != nil {
		return err
	}

	merged := config

	err = mergo.Merge(&merged, existing)
	if err != nil {
		return fmt.Errorf("can't merge new config with existing config: %w", err)
	}

	return setEntireConfig(merged)
}

func setEntireConfig(config Config) error {
	path, err := configFilePath()
	if err != nil {
		return err
	}

	js, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("can't marshal config into JSON: %w", err)
	}

	return writeFile(path, js)
}

// configFilePath returns the path of the user's config file in the XDG config home, creating
// its parent directory if needed. This mirrors how clerk.GetJwtPath locates jwt.json.
func configFilePath() (string, error) {
	path, err := xdg.ConfigFile(fileName)
	if err != nil {
		return "", fmt.Errorf("can't determine config file path: %w", err)
	}

	return path, nil
}

const perm = 0o600 // Regular file with read/write permission for owner

func writeFile(path string, data []byte) error {
	return os.WriteFile(
		path,
		data,
		perm,
	)
}
