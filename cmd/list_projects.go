package cmd

import (
	"errors"
	"fmt"
	"sort"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/spf13/cobra"
)

var listProjectsCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "list:projects",
	Short: "List projects",
	Long:  "List projects",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := flags.GetAPIKey()

		client := request.NewAPIClient("unknown", &apiKey)

		projects, err := client.ListProjects(cmd.Context())
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to list installations", err)
			}
		}

		sort.Slice(projects, func(i, j int) bool {
			return projects[i].Name < projects[j].Name
		})

		for _, proj := range projects {
			logger.Info(fmt.Sprintf("Project ID: %s, Project Name: %s", proj.Id, proj.Name))
		}
	},
}

func init() {
	rootCmd.AddCommand(listProjectsCmd)
}
