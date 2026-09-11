package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

const (
	testProject     = "test-project"
	testDestination = "webhook"
	testProvider    = "salesforce"
)

// testCatalog is the response the stub API serves for the public provider catalog
// endpoint: enough of a provider entry for the validator's provider, module, and
// capability checks to run.
const testCatalog = `{
  "salesforce": {
    "name": "salesforce",
    "displayName": "Salesforce",
    "authType": "oauth2",
    "baseURL": "https://login.salesforce.com",
    "support": {
      "read": true,
      "write": true,
      "subscribe": true,
      "proxy": true
    }
  }
}`

// validManifest references the destination and provider that apiStub serves by
// default, so it validates cleanly both offline and against a project.
const validManifest = `specVersion: 1.0.0
integrations:
  - name: readSalesforce
    provider: salesforce
    read:
      objects:
        - objectName: account
          destination: webhook
          schedule: "*/10 * * * *"
`

// invalidManifest has schema errors: a misspelled top-level key and a misspelled
// object key.
const invalidManifest = `specVersion: 1.0.0
descrption: top-level key that is not part of the schema
integrations:
  - name: readSalesforce
    provider: salesforce
    read:
      objects:
        - objectName: account
          destination: webhook
          scheduel: "*/10 * * * *"
`

// apiStub serves the subset of the Ampersand API that validation reads. Each
// endpoint can be made to fail so the degraded paths are covered too.
type apiStub struct {
	noCatalog        bool
	noDestinations   bool
	noProviderApps   bool
	destinationNames []string
	providerNames    []string
}

// start brings up the stub and returns its root URL.
func (s apiStub) start(t *testing.T) string {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/providers", func(writer http.ResponseWriter, _ *http.Request) {
		if s.noCatalog {
			http.Error(writer, "catalog unavailable", http.StatusInternalServerError)

			return
		}

		writeJSON(t, writer, testCatalog)
	})

	mux.HandleFunc("/v1/projects/"+testProject+"/destinations",
		func(writer http.ResponseWriter, _ *http.Request) {
			if s.noDestinations {
				http.Error(writer, "destinations unavailable", http.StatusInternalServerError)

				return
			}

			writeJSON(t, writer, jsonList(s.destinationNames, "name"))
		})

	mux.HandleFunc("/v1/projects/"+testProject+"/provider-apps",
		func(writer http.ResponseWriter, _ *http.Request) {
			if s.noProviderApps {
				http.Error(writer, "provider apps unavailable", http.StatusInternalServerError)

				return
			}

			writeJSON(t, writer, jsonList(s.providerNames, "provider"))
		})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server.URL
}

// jsonList renders names as a JSON array of objects carrying them under field.
func jsonList(names []string, field string) string {
	items := make([]string, 0, len(names))
	for _, name := range names {
		items = append(items, `{"`+field+`":"`+name+`"}`)
	}

	return "[" + strings.Join(items, ",") + "]"
}

func writeJSON(t *testing.T, writer http.ResponseWriter, body string) {
	t.Helper()

	writer.Header().Set("Content-Type", "application/json")

	_, err := writer.Write([]byte(body))
	if err != nil {
		t.Errorf("writing stub response: %v", err)
	}
}

// configure points the flag-backed config at the stub API for one test.
func configure(t *testing.T, apiURL, project string, strict bool) {
	t.Helper()

	t.Setenv("AMP_API_URL", apiURL)
	viper.Set("project", project)
	viper.Set("key", "test-api-key")

	validateStrict = strict

	t.Cleanup(func() {
		viper.Set("project", "")
		viper.Set("key", "")

		validateStrict = false
	})
}

// writeManifest writes contents to amp.yaml in a fresh directory and returns the
// directory, mimicking how the command is normally pointed at a project folder.
func writeManifest(t *testing.T, contents string) string {
	t.Helper()

	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "amp.yaml"), []byte(contents), 0o600)
	if err != nil {
		t.Fatalf("writing manifest: %v", err)
	}

	return dir
}

// captureOutput collects everything the command prints while function runs.
func captureOutput(t *testing.T, function func()) string {
	t.Helper()

	out, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatalf("creating output file: %v", err)
	}

	orig := os.Stdout
	os.Stdout = out

	defer func() { os.Stdout = orig }()

	function()

	contents, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatalf("reading captured output: %v", err)
	}

	return string(contents)
}

// The validate tests mutate process-wide state — the API URL env var, Viper config, and the
// command flag globals — so they cannot run in parallel.
func TestRunValidateExitCode(t *testing.T) { //nolint:paralleltest
	defaultStub := apiStub{
		destinationNames: []string{testDestination},
		providerNames:    []string{testProvider},
	}

	tests := []struct {
		name     string
		manifest string
		source   func(t *testing.T) string
		project  string
		strict   bool
		stub     apiStub
		want     int
	}{
		{
			name:     "valid manifest without a project",
			manifest: validManifest,
			stub:     defaultStub,
			want:     exitValidateSuccess,
		},
		{
			name:     "valid manifest without a project under strict mode",
			manifest: validManifest,
			strict:   true,
			stub:     defaultStub,
			want:     exitValidateFailure, // the unverifiable destination is a warning
		},
		{
			name:     "manifest with schema errors",
			manifest: invalidManifest,
			stub:     defaultStub,
			want:     exitValidateFailure,
		},
		{
			name:     "manifest is not valid YAML",
			manifest: "specVersion: 1.0.0\nintegrations: [",
			stub:     defaultStub,
			want:     exitValidateFailure,
		},
		{
			name: "no manifest in the directory",
			source: func(t *testing.T) string {
				t.Helper()

				return t.TempDir()
			},
			stub: defaultStub,
			want: exitValidateFailure,
		},
		{
			name: "source does not exist",
			source: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "missing")
			},
			stub: defaultStub,
			want: exitValidateFailure,
		},
		{
			name:     "project checks pass",
			manifest: validManifest,
			project:  testProject,
			stub:     defaultStub,
			want:     exitValidateSuccess,
		},
		{
			name:     "destination missing from the project",
			manifest: validManifest,
			project:  testProject,
			stub: apiStub{
				destinationNames: []string{"some-other-destination"},
				providerNames:    []string{testProvider},
			},
			want: exitValidateFailure,
		},
		{
			name:     "destinations cannot be fetched for the project",
			manifest: validManifest,
			project:  testProject,
			stub: apiStub{
				noDestinations: true,
				providerNames:  []string{testProvider},
			},
			want: exitValidateFailure,
		},
		{
			name:     "provider apps cannot be fetched for the project",
			manifest: validManifest,
			project:  testProject,
			stub: apiStub{
				noProviderApps:   true,
				destinationNames: []string{testDestination},
			},
			want: exitValidateFailure,
		},
		{
			name:     "catalog cannot be fetched",
			manifest: validManifest,
			project:  testProject,
			stub: apiStub{
				noCatalog:        true,
				destinationNames: []string{testDestination},
				providerNames:    []string{testProvider},
			},
			want: exitValidateSuccess, // falls back to the bundled catalog
		},
	}

	for _, test := range tests { //nolint:paralleltest // see the note above
		t.Run(test.name, func(t *testing.T) {
			configure(t, test.stub.start(t), test.project, test.strict)

			var source string

			if test.source != nil {
				source = test.source(t)
			} else {
				source = writeManifest(t, test.manifest)
			}

			var got int

			output := captureOutput(t, func() {
				got = runValidate(context.Background(), source)
			})

			if got != test.want {
				t.Errorf("runValidate() = %d, want %d\noutput:\n%s", got, test.want, output)
			}
		})
	}
}

// TestRunValidateReportsDegradedChecks covers the messages that tell the user which
// checks did not run: skipping them silently is what makes a passing exit code
// misleading.
func TestRunValidateReportsDegradedChecks(t *testing.T) { //nolint:paralleltest
	t.Run("no project configured", func(t *testing.T) { //nolint:paralleltest // see the note above
		configure(t, apiStub{}.start(t), "", false)

		source := writeManifest(t, validManifest)

		output := captureOutput(t, func() {
			runValidate(context.Background(), source)
		})

		if !strings.Contains(output, "No project configured") {
			t.Errorf("expected a note about the skipped project checks, got:\n%s", output)
		}
	})

	t.Run("catalog unavailable", func(t *testing.T) { //nolint:paralleltest // see the note above
		stub := apiStub{
			noCatalog:        true,
			destinationNames: []string{testDestination},
			providerNames:    []string{testProvider},
		}

		configure(t, stub.start(t), testProject, false)

		source := writeManifest(t, validManifest)

		output := captureOutput(t, func() {
			runValidate(context.Background(), source)
		})

		if !strings.Contains(output, "falling back") {
			t.Errorf("expected a warning about the stale catalog fallback, got:\n%s", output)
		}
	})

	t.Run("project checks cannot run", func(t *testing.T) { //nolint:paralleltest // see the note above
		stub := apiStub{
			noDestinations: true,
			providerNames:  []string{testProvider},
		}

		configure(t, stub.start(t), testProject, false)

		source := writeManifest(t, validManifest)

		output := captureOutput(t, func() {
			runValidate(context.Background(), source)
		})

		if !strings.Contains(output, "unable to list the destinations") {
			t.Errorf("expected the destination fetch failure to be reported, got:\n%s", output)
		}
	})
}
