package cmd

import (
	"fmt"
	"os"

	"github.com/amp-labs/cli/vars"
	"github.com/spf13/cobra"
)

var stageCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:    "stage",
	Short:  "Print the stage this binary has been compiled for",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprint(os.Stdout, vars.Stage+"\n")
	},
}

func init() {
	rootCmd.AddCommand(stageCmd)
}
