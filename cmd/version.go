package cmd

import (
	"fmt"

	"github.com/amp-labs/cli/vars"
	"github.com/spf13/cobra"
)

// Print out some basic info about the running binary. Useful for debugging.
var versionCommand = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Ampersand CLI")

		if vars.Stage == "prod" {
			fmt.Println("version: " + vars.Version)
		} else {
			fmt.Println("version: " + vars.Version + " (" + vars.Stage + ")")
		}

		fmt.Println("build date: " + vars.BuildDate)
		fmt.Println("commit: " + vars.CommitID)
		fmt.Println("branch: " + vars.Branch)
	},
}

func init() {
	rootCmd.AddCommand(versionCommand)
}
