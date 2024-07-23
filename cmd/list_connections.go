package cmd

import (
	"errors"
	"sort"
	"time"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/spf13/cobra"
)

var listConnectionsCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:    "list:connections",
	Short:  "List connections",
	Long:   "List connections",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		projectId := flags.GetProjectOrFail()
		apiKey := flags.GetAPIKey()

		client := request.NewAPIClient(projectId, &apiKey)

		conns, err := client.ListConnections(cmd.Context())
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to list connections", err)
			}
		}

		sort.Slice(conns, func(i, j int) bool {
			return conns[i].CreateTime.Before(conns[j].CreateTime)
		})

		for _, conn := range conns {
			logger.Info(conn.Id + " " + conn.CreateTime.Format(time.RFC3339) + " (" + conn.Status + ")")
		}
	},
}

func init() {
	rootCmd.AddCommand(listConnectionsCmd)
}
