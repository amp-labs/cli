package cmd

import (
	"fmt"

	"github.com/amp-labs/cli/vars"
	"github.com/spf13/cobra"
)

// Print out some basic info about the running binary. Useful for debugging.
var versionCommand = &cobra.Command{ //nolint:gochecknoglobals
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Ampersand CLI") //nolint:forbidigo

		if vars.Stage == "prod" {
			fmt.Println("version: " + vars.Version) //nolint:forbidigo
		} else {
			fmt.Println("version: " + vars.Version + " (" + vars.Stage + ")") //nolint:forbidigo
		}

		fmt.Println("build date: " + vars.BuildDate) //nolint:forbidigo
		fmt.Println("commit: " + vars.CommitID)      //nolint:forbidigo
		fmt.Println("branch: " + vars.Branch)        //nolint:forbidigo
	},
}

func init() {
	rootCmd.AddCommand(versionCommand)
}
