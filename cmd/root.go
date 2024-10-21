package cmd

import (
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "amp",
	Short: "Ampersand CLI",
	Long:  "The Ampersand CLI allows you to interact with the Ampersand platform.",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		return
	}
}

func init() {
	// Disable the autcompletion command from being shown in `amp --help`
	// nolint: lll
	// See https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md#adapting-the-default-completion-command
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	err := flags.Init(rootCmd)
	if err != nil {
		logger.FatalErr("unable to initialize flags", err)
	}
}
