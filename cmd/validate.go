package cmd

import (
	"context"
	"errors"
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

var (
	validateStrict       bool
	validateSkipProvider bool
	validateSkipAsync    bool
)

var validateCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "validate [ampYamlSourcePath]",
	Short: "Validate an amp.yaml manifest",
	Long: "Validate an amp.yaml manifest without deploying it.\n\n" +
		"You can provide a path to the folder that contains amp.yaml or a path to the file " +
		"itself; if omitted the current directory is used.\n\n" +
		"When a project is configured (via --project), destinations and provider apps " +
		"referenced by the manifest are checked against your Ampersand project. Without a " +
		"project only schema and best-practice checks run.",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		source := "."
		if len(args) > 0 {
			source = args[0]
		}

		manifestPath, err := files.FindManifestFile(source)
		if err != nil {
			if errors.Is(err, files.ErrBadManifest) {
				logger.Fatal(err.Error())
			}

			logger.FatalErr("Unable to locate manifest", err)
		}

		result, err := ampyaml.ValidateFile(cmd.Context(), manifestPath, buildValidateOptions(cmd.Context())...)
		if err != nil {
			logger.FatalErr("Unable to validate manifest", err)
		}

		printValidationResult(manifestPath, result)

		if !result.Valid {
			os.Exit(1)
		}
	},
}

// buildValidateOptions assembles the validator options from the command flags and,
// when a project is configured, the API-backed checkers.
func buildValidateOptions(ctx context.Context) []ampyaml.Option {
	var opts []ampyaml.Option

	if validateStrict {
		opts = append(opts, ampyaml.WithStrictMode(true))
	}

	if validateSkipProvider {
		opts = append(opts, ampyaml.WithSkipProviderValidation())
	}

	if validateSkipAsync {
		opts = append(opts, ampyaml.WithSkipAsyncValidation())
	}

	opts = append(opts, catalogOption(ctx))
	opts = append(opts, apiCheckerOptions(ctx)...)

	return opts
}

// catalogOption backs provider/module/capability validation with the live ("dynamic")
// provider catalog fetched from the API, which changes several times a day. The
// catalog endpoint is public, so this runs regardless of whether a project is
// configured. If the fetch fails (e.g. offline), it degrades gracefully to the
// catalog embedded in the connectors library.
func catalogOption(ctx context.Context) ampyaml.Option {
	catProvider, err := validate.NewCatalogProvider(ctx)
	if err != nil {
		logger.Debugf("Unable to fetch the live provider catalog, "+
			"falling back to the embedded catalog: %v", err)

		return ampyaml.WithCatalogProvider(catalog.NewDefaultCatalogProvider())
	}

	return ampyaml.WithCatalogProvider(catProvider)
}

// apiCheckerOptions wires the destination and provider-app checkers to the Ampersand
// API. If no project is configured these checks are skipped, and the validator falls
// back to emitting reminder warnings instead of hard errors. Failures fetching either
// list are logged (debug) and treated as "checker unavailable" rather than fatal so
// that offline/schema validation still succeeds.
func apiCheckerOptions(ctx context.Context) []ampyaml.Option {
	projectID := flags.GetProject()
	if projectID == "" {
		logger.Debugf("No project configured; skipping destination and provider-app checks. " +
			"Pass --project to validate these against your Ampersand project.")

		return nil
	}

	apiKey := flags.GetAPIKey()
	client := request.NewAPIClient(projectID, &apiKey)

	var opts []ampyaml.Option

	destChecker, err := validate.NewDestinationChecker(ctx, client)
	if err != nil {
		logger.Debugf("Unable to fetch destinations for validation, skipping destination checks: %v", err)
	} else {
		opts = append(opts, ampyaml.WithDestinationChecker(destChecker))
	}

	appChecker, err := validate.NewProviderAppChecker(ctx, client)
	if err != nil {
		logger.Debugf("Unable to fetch provider apps for validation, skipping provider-app checks: %v", err)
	} else {
		opts = append(opts, ampyaml.WithProviderAppChecker(appChecker))
	}

	return opts
}

func printValidationResult(manifestPath string, result *ampyaml.ValidationResult) {
	logger.Infof("Validating: %s", manifestPath)

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

	if result.Valid {
		logger.Infof("✓ Validation passed with %d warning(s)", len(result.Warnings))
	} else {
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
	validateCmd.Flags().BoolVar(&validateSkipProvider, "skip-provider", false,
		"Skip provider-specific validation")
	validateCmd.Flags().BoolVar(&validateSkipAsync, "skip-async", false,
		"Skip async error-prevention validation")

	rootCmd.AddCommand(validateCmd)
}
