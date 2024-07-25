package cmd

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/buildkite/shellwords"
	"github.com/spf13/cobra"
)

func quote(s string) string {
	return "\"" + strings.Trim(shellwords.Quote(s), "\"") + "\""
}

var listProviderAppsCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:    "list:provider-apps",
	Short:  "List provider apps",
	Long:   "List provider apps",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		projectId := flags.GetProjectOrFail()
		apiKey := flags.GetAPIKey()

		client := request.NewAPIClient(projectId, &apiKey)

		apps, err := client.ListProviderApps(cmd.Context())
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to list connections", err)
			}
		}

		sort.Slice(apps, func(i, j int) bool {
			return apps[i].CreateTime.Before(apps[j].CreateTime)
		})

		for _, app := range apps {
			parts := []string{
				app.Id,
				app.CreateTime.Format(time.RFC3339),
				app.Provider,
			}

			if len(app.ClientId) > 0 {
				parts = append(parts, "clientId="+quote(app.ClientId))
			}

			if len(app.ClientSecret) > 0 {
				secret := strings.Map(func(_ rune) rune {
					return '*'
				}, app.ClientSecret)

				parts = append(parts, "clientSecret="+quote(secret))
			}

			if app.ExternalRef != "" {
				parts = append(parts, "externalRef="+quote(app.ExternalRef))
			}

			if len(app.Scopes) > 0 {
				for i, scope := range app.Scopes {
					app.Scopes[i] = quote(scope)
				}

				parts = append(parts, "scopes=["+strings.Join(app.Scopes, ", ")+"]")
			}

			logger.Info(strings.Join(parts, " "))
		}
	},
}

func init() {
	rootCmd.AddCommand(listProviderAppsCmd)
}
