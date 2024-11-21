package cmd

import (
	"fmt"
	"os"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/logger"
	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command.
var logoutCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "logout",
	Short: "Log out of an ampersand account",
	Long:  "Log out of an ampersand account.",
	Run: func(cmd *cobra.Command, args []string) {
		DoLogout(true)
	},
}

func DoLogout(showLogs bool) {
	path := clerk.GetJwtPath()

	_, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			if showLogs {
				fmt.Println("You're already logged out") //nolint:forbidigo
			}

			return
		} else {
			logger.Fatal(err.Error())
		}
	}

	if err := os.Remove(path); err != nil {
		logger.Fatal(err.Error())
	}

	if showLogs {
		fmt.Println("Successfully logged out!") //nolint:forbidigo
	}
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
