package cmd

import (
	"errors"
	"sort"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/spf13/cobra"
)

var listDestinationsCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "list:destinations",
	Short: "List destinations",
	Long:  "List destinations",
	Run: func(cmd *cobra.Command, args []string) {
		projectId := flags.GetProjectOrFail()
		apiKey := flags.GetAPIKey()
		client := request.NewAPIClient(projectId, &apiKey)

		destinations, err := client.ListDestinations(cmd.Context())
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to list destinations", err)
			}
		}

		sort.Slice(destinations, func(i, j int) bool {
			return destinations[i].Name < destinations[j].Name
		})

		for _, dest := range destinations {
			output := dest.Name
			if dest.Metadata.URL != "" {
				output += " (" + dest.Metadata.URL + ")"
			}
			logger.Info(output)
		}
	},
}

func init() {
	rootCmd.AddCommand(listDestinationsCmd)
}
