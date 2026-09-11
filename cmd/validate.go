package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	ampyaml "github.com/amp-labs/amp-yaml-validator"
	"github.com/amp-labs/amp-yaml-validator/catalog"
	"github.com/amp-labs/cli/files"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/amp-labs/cli/validate"
	"github.com/spf13/cobra"
)

// Exit codes for the validate command. Zero means the manifest is clean; every
// failure — an invalid manifest, a manifest that can't be found or parsed, or a
// check that couldn't be run — is exitValidateFailure, so CI and agents only have
// to distinguish zero from non-zero.
const (
	exitValidateSuccess = 0
	exitValidateFailure = 1
)

var validateStrict bool

var validateCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "validate [ampYamlSourcePath]",
	Short: "Validate an amp.yaml manifest",
	Long: "Validate an amp.yaml manifest without deploying it.\n\n" +
		"You can provide a path to the folder that contains amp.yaml or a path to the file " +
		"itself; if omitted the current directory is used.\n\n" +
		"When a project is configured (via --project), destinations and provider apps " +
		"referenced by the manifest are checked against your Ampersand project, and the " +
		"command fails if those checks can't be run. Without a project only schema and " +
		"best-practice checks run.\n\n" +
		"Exits 0 when the manifest is clean and 1 otherwise, so it can gate CI.",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		source := "."
		if len(args) > 0 {
			source = args[0]
		}

		if code := runValidate(cmd.Context(), source); code != exitValidateSuccess {
			os.Exit(code)
		}
	},
}

// runValidate validates the manifest found at source and returns the process exit
// code. It reports failures itself rather than calling logger.Fatal so that the
// exit-code decisions are testable.
func runValidate(ctx context.Context, source string) int {
	manifestPath, err := files.FindManifestFile(source)
	if err != nil {
		if errors.Is(err, files.ErrBadManifest) {
			logger.Info(err.Error())
		} else {
			printValidateError("Unable to locate manifest", err)
		}

		return exitValidateFailure
	}

	logger.Infof("Validating: %s", manifestPath)

	opts, err := buildValidateOptions(ctx)
	if err != nil {
		printValidateError("Unable to validate manifest", err)

		return exitValidateFailure
	}

	result, err := ampyaml.ValidateFile(ctx, manifestPath, opts...)
	if err != nil {
		printValidateError("Unable to validate manifest", err)

		return exitValidateFailure
	}

	printValidationResult(result)

	if !result.Valid {
		return exitValidateFailure
	}

	return exitValidateSuccess
}

// printValidateError reports a failure the way logger.FatalErr would, minus the
// exit, leaving the exit code to the caller.
func printValidateError(msg string, err error) {
	logger.Infof("✗ %s\nerror: %v", msg, err)
	logger.PrintDebugTip()
}

// buildValidateOptions assembles the validator options from the command flags, the
// provider catalog, and — when a project is configured — the API-backed checkers.
func buildValidateOptions(ctx context.Context) ([]ampyaml.Option, error) {
	opts := []ampyaml.Option{catalogOption(ctx)}

	if validateStrict {
		opts = append(opts, ampyaml.WithStrictMode(true))
	}

	checkerOpts, err := apiCheckerOptions(ctx)
	if err != nil {
		return nil, err
	}

	return append(opts, checkerOpts...), nil
}

// catalogOption backs provider/module/capability validation with the live ("dynamic")
// provider catalog fetched from the API, which changes several times a day. The
// catalog endpoint is public, so this runs regardless of whether a project is
// configured. If the fetch fails (e.g. offline), it degrades to the catalog embedded
// in the connectors library and says so, since that catalog goes stale between
// releases.
func catalogOption(ctx context.Context) ampyaml.Option {
	catProvider, err := validate.NewCatalogProvider(ctx)
	if err != nil {
		logger.Info("⚠ Unable to fetch the live provider catalog; falling back to the catalog " +
			"bundled with this CLI. Provider and module checks may be out of date.")
		logger.Debugf("Provider catalog fetch failed: %v", err)

		return ampyaml.WithCatalogProvider(catalog.NewDefaultCatalogProvider())
	}

	return ampyaml.WithCatalogProvider(catProvider)
}

// apiCheckerOptions wires the destination and provider-app checkers to the Ampersand
// API. Without a project there is nothing to check them against, so those checks are
// left out and the rest of validation runs offline. With a project the user is asking
// for the manifest to be checked against that project, so a failure to fetch either
// list is reported as an error: skipping the check would let validation report
// success without ever having run it.
func apiCheckerOptions(ctx context.Context) ([]ampyaml.Option, error) {
	projectID := flags.GetProject()
	if projectID == "" {
		logger.Info("ℹ No project configured; skipping destination and provider-app checks. " +
			"Pass --project to check those against your Ampersand project.")

		return nil, nil
	}

	apiKey := flags.GetAPIKey()
	client := request.NewAPIClient(projectID, &apiKey)

	destChecker, err := validate.NewDestinationChecker(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("unable to list the destinations in project %q: %w", projectID, err)
	}

	appChecker, err := validate.NewProviderAppChecker(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("unable to list the provider apps in project %q: %w", projectID, err)
	}

	return []ampyaml.Option{
		ampyaml.WithDestinationChecker(destChecker),
		ampyaml.WithProviderAppChecker(appChecker),
	}, nil
}

func printValidationResult(result *ampyaml.ValidationResult) {
	if result.Valid && len(result.Warnings) == 0 {
		logger.Info("✓ Validation passed with no issues!")

		return
	}

	if len(result.Errors) > 0 {
		logger.Infof("\n✗ Errors (%d):", len(result.Errors))

		for i, issue := range result.Errors {
			printValidationIssue(i+1, issue)
		}
	}

	if len(result.Warnings) > 0 {
		logger.Infof("\n⚠ Warnings (%d):", len(result.Warnings))

		for i, issue := range result.Warnings {
			printValidationIssue(i+1, issue)
		}
	}

	logger.Info("")

	switch {
	case result.Valid:
		logger.Infof("✓ Validation passed with %d warning(s)", len(result.Warnings))
	case len(result.Errors) == 0:
		logger.Infof("✗ Validation failed: %d warning(s) treated as errors by --strict",
			len(result.Warnings))
	default:
		logger.Infof("✗ Validation failed with %d error(s) and %d warning(s)",
			len(result.Errors), len(result.Warnings))
	}
}

func printValidationIssue(num int, issue ampyaml.ValidationIssue) {
	logger.Infof("\n  %d. [%s] %s", num, issue.Rule, issue.Message)

	if issue.Path != "" {
		logger.Infof("     Path: %s", issue.Path)
	}

	if issue.Line > 0 {
		if issue.Column > 0 {
			logger.Infof("     Location: line %d, column %d", issue.Line, issue.Column)
		} else {
			logger.Infof("     Location: line %d", issue.Line)
		}
	}

	if issue.Suggestion != "" {
		logger.Infof("     Suggestion: %s", issue.Suggestion)
	}
}

func init() {
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "Treat warnings as errors")

	rootCmd.AddCommand(validateCmd)
}
