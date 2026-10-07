package region

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/adrg/xdg"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    Region
		wantErr bool
	}{
		{name: "us", input: "us", want: US},
		{name: "eu", input: "eu", want: EU},
		{name: "uppercase", input: "EU", want: EU},
		{name: "mixed case", input: "eU", want: EU},
		{name: "whitespace", input: "  eu  ", want: EU},
		{name: "unknown region", input: "ap", want: Region("ap")},
		{name: "hyphen", input: "eu-west", want: Region("eu-west")},
		{name: "digit", input: "eu2", want: Region("eu2")},
		{name: "max length", input: strings.Repeat("a", maxLabelLen), want: Region(strings.Repeat("a", maxLabelLen))},
		{name: "empty", input: "", wantErr: true},
		{name: "inner space", input: "eu west", wantErr: true},
		{name: "leading hyphen", input: "-eu", wantErr: true},
		{name: "trailing hyphen", input: "eu-", wantErr: true},
		{name: "leading digit", input: "1eu", wantErr: true},
		{name: "punctuation", input: "eu!", wantErr: true},
		{name: "dot", input: "eu.west", wantErr: true},
		{name: "too long", input: strings.Repeat("a", maxLabelLen+1), wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(testCase.input)

			if testCase.wantErr {
				if !errors.Is(err, ErrInvalid) {
					t.Errorf("Parse(%q) error = %v, want ErrInvalid", testCase.input, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Parse(%q) error = %v", testCase.input, err)
			}

			if got != testCase.want {
				t.Errorf("Parse(%q) = %q, want %q", testCase.input, got, testCase.want)
			}
		})
	}
}

func TestKnown(t *testing.T) {
	t.Parallel()

	got := Known()
	want := []string{"us", "eu"}

	if len(got) != len(want) {
		t.Fatalf("Known() = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Known() = %v, want %v", got, want)
		}
	}
}

func TestKnownOrder(t *testing.T) {
	t.Parallel()

	got := Known()

	if len(got) == 0 {
		t.Fatal("Known() is empty")
	}

	if got[0] != string(Default) {
		t.Errorf("Known()[0] = %q, want %q", got[0], Default)
	}

	if !sort.StringsAreSorted(got[1:]) {
		t.Errorf("Known()[1:] = %v, want sorted", got[1:])
	}
}

func TestIsKnown(t *testing.T) {
	t.Parallel()

	for _, name := range Known() {
		if !IsKnown(Region(name)) {
			t.Errorf("IsKnown(%q) = false, want true", name)
		}
	}

	if IsKnown(Region("ap")) {
		t.Error(`IsKnown("ap") = true, want false`)
	}
}

func TestRegionalizeURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		rawURL string
		region Region
		want   string
	}{
		{
			name:   "us",
			rawURL: "https://api.withampersand.com",
			region: US,
			want:   "https://api.withampersand.com",
		},
		{
			name:   "eu",
			rawURL: "https://api.withampersand.com",
			region: EU,
			want:   "https://api.eu.withampersand.com",
		},
		{
			name:   "unknown region",
			rawURL: "https://api.withampersand.com",
			region: Region("ap"),
			want:   "https://api.ap.withampersand.com",
		},
		{
			name:   "stage prefix",
			rawURL: "https://staging-api.withampersand.com",
			region: EU,
			want:   "https://staging-api.eu.withampersand.com",
		},
		{
			name:   "sign-in host",
			rawURL: "https://cli-signin.withampersand.com",
			region: EU,
			want:   "https://cli-signin.eu.withampersand.com",
		},
		{
			name:   "path and query",
			rawURL: "https://api.withampersand.com/v1?debug=true",
			region: EU,
			want:   "https://api.eu.withampersand.com/v1?debug=true",
		},
		{
			name:   "port",
			rawURL: "https://api.withampersand.com:8443",
			region: EU,
			want:   "https://api.eu.withampersand.com:8443",
		},
		{
			name:   "already regionalized",
			rawURL: "https://api.eu.withampersand.com",
			region: EU,
			want:   "https://api.eu.withampersand.com",
		},
		{
			name:   "localhost",
			rawURL: "http://localhost:8080",
			region: EU,
			want:   "http://localhost:8080",
		},
		{
			name:   "loopback",
			rawURL: "http://127.0.0.1:4010",
			region: EU,
			want:   "http://127.0.0.1:4010",
		},
		{
			name:   "other domain",
			rawURL: "https://api.example.com",
			region: EU,
			want:   "https://api.example.com",
		},
		{
			name:   "bare domain",
			rawURL: "https://withampersand.com",
			region: EU,
			want:   "https://withampersand.com",
		},
		{
			name:   "suffix lookalike",
			rawURL: "https://api.notwithampersand.com",
			region: EU,
			want:   "https://api.notwithampersand.com",
		},
		{
			name:   "unparseable",
			rawURL: "://not a url",
			region: EU,
			want:   "://not a url",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.region.RegionalizeURL(testCase.rawURL)
			if got != testCase.want {
				t.Errorf("RegionalizeURL(%q) = %q, want %q", testCase.rawURL, got, testCase.want)
			}
		})
	}
}

func setupConfig(t *testing.T, contents string) {
	t.Helper()

	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	xdg.Reload()

	t.Cleanup(xdg.Reload)

	if contents == "" {
		return
	}

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
func TestResolve(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		env       string
		flagValue string
		want      Region
		wantErr   bool
	}{
		{name: "default", config: "", flagValue: "", want: Default},
		{name: "saved", config: `{"region":"eu"}`, flagValue: "", want: EU},
		{name: "env", config: "", env: "eu", flagValue: "", want: EU},
		{name: "flag over saved", config: `{"region":"eu"}`, flagValue: "us", want: US},
		{name: "flag over env", config: "", env: "eu", flagValue: "us", want: US},
		{name: "saved over env", config: `{"region":"eu"}`, env: "ap", flagValue: "", want: EU},
		{name: "flag unknown", config: "", flagValue: "ap", want: Region("ap")},
		{name: "flag malformed", config: "", flagValue: "eu west", wantErr: true},
		{name: "saved malformed", config: `{"region":"eu west"}`, flagValue: "", wantErr: true},
		{name: "env malformed", config: "", env: "eu west", flagValue: "", wantErr: true},
		{name: "corrupt config", config: `{"region":`, flagValue: "", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			setupConfig(t, testCase.config)
			// Always set, so that AMP_REGION in the developer's own shell cannot leak in.
			t.Setenv(EnvVar, testCase.env)

			got, err := Resolve(testCase.flagValue)

			if testCase.wantErr {
				if err == nil {
					t.Errorf("Resolve(%q) = %q, want error", testCase.flagValue, got)
				}

				return
			}

			if err != nil {
				t.Fatalf("Resolve(%q) error = %v", testCase.flagValue, err)
			}

			if got != testCase.want {
				t.Errorf("Resolve(%q) = %q, want %q", testCase.flagValue, got, testCase.want)
			}
		})
	}
}
