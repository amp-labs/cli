package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of an ampersand account",
	Long:  "Log out of an ampersand account.",
	Run: func(cmd *cobra.Command, args []string) {
		path := getJwtPath()
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return
			} else {
				log.Fatalln(err)
			}
		}

		if err := os.Remove(path); err != nil {
			log.Fatalln(err)
		}

		fmt.Println("logout successful")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
