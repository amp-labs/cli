package cmd

import (
	"fmt"
	"log"

	"github.com/amp-labs/cli/clerk"
	"github.com/spf13/cobra"
)

// tokenCmd represents the generate-request-token command.
var tokenCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "generate-request-token",
	Short: "Generate a request token",
	Long: "Generate a JWT token to be used for HTTP requests, and prints it." +
		" This command is useful for testing purposes.",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		jwt, err := clerk.FetchJwt(cmd.Context())
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(jwt) //nolint:forbidigo
	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)
}
