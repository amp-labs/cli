package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra" //nolint:gosec
	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/amp-labs/cli/utils"
	"github.com/amp-labs/cli/vars"
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
			Root:   fmt.Sprintf("%s/%s", rootURL, request.ApiVersion),
			Client: request.NewRequestClient(),
		}

		info, err := client.GetMyInfo(cmd.Context())
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Failed to get user info", err)
			}
		}

		format := flags.GetOutputFormat()
		if err := utils.WriteStruct(os.Stdout, format, info); err != nil {
			logger.FatalErr("Unable to write user info", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(myInfoCmd)
}
