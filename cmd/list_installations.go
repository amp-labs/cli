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

var listInstallationsCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "list:installations <integrationId>",
	Short: "List installations",
	Long:  "List installations",
	Args:  cobra.ExactArgs(1), //nolint:gomnd,mnd
	Run: func(cmd *cobra.Command, args []string) {
		integrationId := args[0]
		projectId := flags.GetProjectId()
		if projectId == "" {
			logger.Fatal("Must provide a project ID in the --project flag")
		}

		apiKey := flags.GetAPIKey()

		insts, err := request.NewAPIClient(projectId, &apiKey).
			ListInstallations(cmd.Context(), integrationId)
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to list installations", err)
			}
		}

		sort.Slice(insts, func(i, j int) bool {
			return insts[i].CreateTime.Before(insts[j].CreateTime)
		})

		for _, inst := range insts {
			logger.Info(inst.Id + " " + inst.CreatedBy + " (" + inst.HealthStatus + ")")
		}
	},
}

func init() {
	rootCmd.AddCommand(listInstallationsCmd)
}
