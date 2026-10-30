package appdata

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

func setup(t *testing.T) string {
	t.Helper()

	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	t.Cleanup(xdg.Reload)

	return configHome
}

func writeConfig(t *testing.T, configHome string, contents string) {
	t.Helper()

	configDir := filepath.Join(configHome, "Ampersand")

	err := os.MkdirAll(configDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(configDir, "config.json"), []byte(contents), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}

//nolint:paralleltest // mutates the environment
func TestGetNoFile(t *testing.T) {
	setup(t)

	config, err := Get()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if config.Region != "" {
		t.Errorf("Get() = %+v, want zero config", config)
	}
}

//nolint:paralleltest // mutates the environment
func TestGetCorruptFile(t *testing.T) {
	configHome := setup(t)
	writeConfig(t, configHome, `{"region":`)

	_, err := Get()
	if err == nil {
		t.Error("Get() error = nil, want error")
	}
}

//nolint:paralleltest // mutates the environment
func TestSetCreatesFile(t *testing.T) {
	configHome := setup(t)

	err := Set(Config{Region: "eu"})
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	contents, err := os.ReadFile(filepath.Join(configHome, "Ampersand", "config.json"))
	if err != nil {
		t.Fatal(err)
	}

	if string(contents) != `{"region":"eu"}` {
		t.Errorf("config.json = %s, want %s", contents, `{"region":"eu"}`)
	}
}

//nolint:paralleltest // mutates the environment
func TestSetGetRoundTrip(t *testing.T) {
	setup(t)

	err := Set(Config{Region: "eu"})
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	config, err := Get()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if config.Region != "eu" {
		t.Errorf("Get().Region = %q, want %q", config.Region, "eu")
	}
}

//nolint:paralleltest // mutates the environment
func TestSetOverwrites(t *testing.T) {
	setup(t)

	err := Set(Config{Region: "eu"})
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	err = Set(Config{Region: "us"})
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	config, err := Get()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if config.Region != "us" {
		t.Errorf("Get().Region = %q, want %q", config.Region, "us")
	}
}

//nolint:paralleltest // mutates the environment
func TestSetCorruptFile(t *testing.T) {
	configHome := setup(t)
	writeConfig(t, configHome, `{"region":`)

	err := Set(Config{Region: "us"})
	if err == nil {
		t.Fatal("Set() error = nil, want error")
	}

	contents, err := os.ReadFile(filepath.Join(configHome, "Ampersand", "config.json"))
	if err != nil {
		t.Fatal(err)
	}

	if string(contents) != `{"region":` {
		t.Errorf("config.json = %s, want it unchanged", contents)
	}
}
