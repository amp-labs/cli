package cmd

import (
	"errors"
	"os"
	"sort"
	"time"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/amp-labs/cli/utils"
	"github.com/spf13/cobra"
)

type operationSummary struct {
	Id          string  `json:"id"`
	Action      string  `json:"action"`
	Status      string  `json:"status"`
	ReadType    string  `json:"readType,omitempty"`
	StartedAt   string  `json:"startedAt"`
	CompletedAt *string `json:"completedAt"`
}

func summarizeOperations(operations []*request.Operation) []operationSummary {
	summaries := make([]operationSummary, 0, len(operations))

	for _, operation := range operations {
		var completedAt *string

		if operation.UpdateTime != nil {
			formatted := operation.UpdateTime.Format(time.RFC3339)
			completedAt = &formatted
		}

		summaries = append(summaries, operationSummary{
			Id:          operation.Id,
			Action:      operation.ActionType,
			Status:      operation.Status,
			ReadType:    operation.ReadType,
			StartedAt:   operation.CreateTime.Format(time.RFC3339),
			CompletedAt: completedAt,
		})
	}

	return summaries
}

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

		format, err := cmd.Flags().GetString("format")
		if err != nil {
			logger.FatalErr("Unable to read output format", err)
		}

		err = utils.WriteStruct(os.Stdout, utils.Format(format), summarizeOperations(operations))
		if err != nil {
			logger.FatalErr("Unable to write operations", err)
		}
	},
}

func init() {
	err := flags.InitAndBindFormatFlag(listOperationsCmd)
	if err != nil {
		logger.FatalErr("unable to initialize flags", err)
	}

	rootCmd.AddCommand(listOperationsCmd)
}
