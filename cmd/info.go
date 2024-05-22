package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/amp-labs/cli/vars"
	//nolint:gosec
	"github.com/spf13/cobra"
)

var myInfoCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "my-info",
	Short: "Get info about the current user",
	Long:  "Get info about the current user",
	Run: func(cmd *cobra.Command, args []string) {
		rootURL, ok := os.LookupEnv("AMP_API_URL")
		if !ok {
			rootURL = vars.ApiURL
		}

		client := &request.APIClient{
			Root:          fmt.Sprintf("%s/%s", rootURL, request.API_VERSION),
			RequestClient: request.NewRequestClient(),
		}

		info, err := client.GetMyInfo(cmd.Context())
		if err != nil {
			logger.FatalErr("Failed to get user info", err)
		}

		js, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			logger.FatalErr("Failed to marshal user info", err)
		}

		fmt.Println(string(js))
	},
}

func init() {
	rootCmd.AddCommand(myInfoCmd)
}
