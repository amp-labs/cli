package cmd

import (
	"context"
	"errors"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/spf13/cobra"
)

type oauthClient interface {
	GenerateOAuthAuthorizationURL(
		ctx context.Context,
		params *request.OAuthAuthorizationURLParams,
	) (string, error)
}

type oauthClientFactory func(projectId string, apiKey string) oauthClient

type connectProjectResolver func() string

type browserDetector func() bool

type browserOpener func(url string)

func newConnectProviderCmd(
	newClient oauthClientFactory,
	getProjectId connectProjectResolver,
	hasBrowser browserDetector,
	openURL browserOpener,
) *cobra.Command {
	var (
		groupRef      string
		consumerRef   string
		providerAppId string
	)

	cmd := &cobra.Command{
		Use:   "connect:provider <provider>",
		Short: "Connect an account to a provider",
		Long: "Generate an OAuth authorization URL. Open it in a browser when available; " +
			"otherwise, print it.",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectId := getProjectId()
			client := newClient(projectId, flags.GetAPIKey())

			url, err := client.GenerateOAuthAuthorizationURL(
				cmd.Context(),
				&request.OAuthAuthorizationURLParams{
					ProjectIdOrName: projectId,
					Provider:        args[0],
					GroupRef:        groupRef,
					ConsumerRef:     consumerRef,
					ProviderAppId:   providerAppId,
				},
			)
			if err != nil {
				if errors.Is(err, clerk.ErrNoSessions) {
					logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
				} else {
					logger.FatalErr("Unable to generate OAuth authorization URL", err)
				}
			}

			if hasBrowser() {
				openURL(url)
				logger.Info("Opened the authorization URL in your browser.")

				return
			}

			logger.Infof("Authorization URL: %s", url)
		},
	}

	cmd.Flags().StringVar(&groupRef, "group-ref", "", "Identifier for the organization or workspace")
	cmd.Flags().StringVar(&consumerRef, "consumer-ref", "", "Identifier for the user authorizing the connection")
	cmd.Flags().StringVar(&providerAppId, "provider-app", "", "Provider app ID (uses the project default if omitted)")

	for _, flag := range []string{"group-ref", "consumer-ref"} {
		err := cmd.MarkFlagRequired(flag)
		if err != nil {
			logger.FatalErr("unable to require "+flag+" flag", err)
		}
	}

	return cmd
}

var connectProviderCmd = newConnectProviderCmd( //nolint:gochecknoglobals
	func(projectId string, apiKey string) oauthClient {
		return request.NewAPIClient(projectId, &apiKey)
	},
	flags.GetProjectOrFail,
	canOpenBrowser,
	openBrowser,
)

func init() {
	rootCmd.AddCommand(connectProviderCmd)
}
