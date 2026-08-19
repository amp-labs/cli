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

var listOperationsCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "list:operations <integrationId> <installationId>",
	Short: "List operations for an installation",
	Long:  "List recent operations for an installation.",
	Args:  cobra.ExactArgs(2), //nolint:mnd
	Run: func(cmd *cobra.Command, args []string) {
		projectId := flags.GetProjectOrFail()
		apiKey := flags.GetAPIKey()
		client := request.NewAPIClient(projectId, &apiKey)

		operations, err := client.ListOperations(cmd.Context(), args[0], args[1])
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to list operations", err)
			}
		}

		sort.Slice(operations, func(i, j int) bool {
			return operations[i].CreateTime.After(operations[j].CreateTime)
		})

		for _, operation := range operations {
			completed := "-"
			if operation.UpdateTime != nil {
				completed = operation.UpdateTime.Format(time.RFC3339)
			}

			logger.Infof(
				"Operation ID: %s, Action: %s, Resource: %s, Read Type: %s, Status: %s, Started: %s, Completed: %s",
				operation.Id,
				operation.ActionType,
				operation.Resource,
				operation.ReadType,
				operation.Status,
				operation.CreateTime.Format(time.RFC3339),
				completed,
			)
		}
	},
}

func init() {
	rootCmd.AddCommand(listOperationsCmd)
}
