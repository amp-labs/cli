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

var listIntegrationsCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "list:integrations",
	Short: "List integrations",
	Long:  "List integrations",
	Run: func(cmd *cobra.Command, args []string) {
		projectId := flags.GetProjectId()
		if projectId == "" {
			logger.Fatal("Must provide a project ID in the --project flag")
		}

		apiKey := flags.GetAPIKey()

		ints, err := request.NewAPIClient(projectId, &apiKey).ListIntegrations(cmd.Context())
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to list installations", err)
			}
		}

		sort.Slice(ints, func(i, j int) bool {
			return ints[i].CreateTime.Before(ints[j].CreateTime)
		})

		for _, inst := range ints {
			logger.Info(inst.Id + " " + inst.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(listIntegrationsCmd)
}
