package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/amp-labs/cli/clerk"
	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command.
var logoutCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "logout",
	Short: "Log out of an ampersand account",
	Long:  "Log out of an ampersand account.",
	Run: func(cmd *cobra.Command, args []string) {
		path := clerk.GetJwtPath()
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("You're already logged out") //nolint:forbidigo

				return
			} else {
				log.Fatalln(err)
			}
		}

		if err := os.Remove(path); err != nil {
			log.Fatalln(err)
		}

		fmt.Println("logout successful") //nolint:forbidigo
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
