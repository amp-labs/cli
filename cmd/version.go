package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/amp-labs/cli/utils"
	"github.com/spf13/cobra"
)

// Print out some basic info about the running binary. Useful for debugging.
var versionCommand = &cobra.Command{ //nolint:gochecknoglobals
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Ampersand CLI") //nolint:forbidigo

		versionInfo := utils.GetVersionInformation()
		if versionInfo.Version == "" {
			versionInfo.Version = "unknown"
		}

		if versionInfo.Stage == utils.Prod {
			fmt.Println("version: " + versionInfo.Version) //nolint:forbidigo
		} else {
			fmt.Println("version: " + versionInfo.Version + " (" + string(versionInfo.Stage) + ")") //nolint:forbidigo
		}

		unixTime, err := strconv.ParseInt(versionInfo.BuildDate, 10, 64)
		if unixTime > 0 && err == nil {
			fmt.Println("build date: " + time.Unix(unixTime, 0).Format(time.RFC3339)) //nolint:forbidigo
		} else {
			fmt.Println("build date: " + versionInfo.BuildDate) //nolint:forbidigo
		}

		fmt.Println("commit: " + versionInfo.CommitID) //nolint:forbidigo
		fmt.Println("branch: " + versionInfo.Branch)   //nolint:forbidigo
	},
}

func init() {
	rootCmd.AddCommand(versionCommand)
}
