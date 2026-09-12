package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const maxClientSecretBytes = 16 * 1024

var (
	errClientSecretEmpty     = errors.New("client secret cannot be empty")
	errClientSecretMultiline = errors.New("client secret must contain exactly one line")
	errClientSecretTooLong   = errors.New("client secret is too long")
)

type providerAppClient interface {
	CreateProviderApp(ctx context.Context, params *request.CreateProviderAppParams) (*request.ProviderApp, error)
}

type providerAppClientFactory func(projectId string, apiKey string) providerAppClient

type projectIdResolver func() string

func readClientSecret(reader io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxClientSecretBytes+1))
	if err != nil {
		return "", fmt.Errorf("read client secret: %w", err)
	}

	if len(data) > maxClientSecretBytes {
		return "", errClientSecretTooLong
	}

	secret := strings.TrimSuffix(string(data), "\n")
	secret = strings.TrimSuffix(secret, "\r")

	if strings.TrimSpace(secret) == "" {
		return "", errClientSecretEmpty
	}

	if strings.ContainsAny(secret, "\r\n") {
		return "", errClientSecretMultiline
	}

	return secret, nil
}

func promptClientSecret() (string, error) {
	prompt := promptui.Prompt{
		Label:  "Client secret",
		Mask:   '*',
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}

	secret, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("read client secret: %w", err)
	}

	if strings.TrimSpace(secret) == "" {
		return "", errClientSecretEmpty
	}

	return secret, nil
}

func providerAppCreatedMessage(providerApp *request.ProviderApp) string {
	scopes := "none"
	if len(providerApp.Scopes) > 0 {
		scopes = strings.Join(providerApp.Scopes, ", ")
	}

	return fmt.Sprintf("Provider App ID: %s, Provider: %s, Scopes: %s",
		providerApp.Id, providerApp.Provider, scopes)
}

func newCreateProviderAppCmd(
	newClient providerAppClientFactory,
	getProjectId projectIdResolver,
) *cobra.Command {
	var (
		clientId          string
		scopes            []string
		clientSecretStdin bool
	)

	cmd := &cobra.Command{
		Use:   "create:provider-app <provider>",
		Short: "Create a provider app",
		Long: "Create an OAuth provider app in an Ampersand project. " +
			"The client secret is requested in a masked prompt unless --client-secret-stdin is set.",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectId := getProjectId()
			apiKey := flags.GetAPIKey()

			if flags.GetDebugMode() {
				logger.Fatal("Debug logging is not available when submitting a client secret")
			}

			var (
				clientSecret string
				err          error
			)

			if clientSecretStdin {
				clientSecret, err = readClientSecret(cmd.InOrStdin())
			} else {
				clientSecret, err = promptClientSecret()
			}

			if err != nil {
				logger.FatalErr("Unable to read client secret", err)
			}

			client := newClient(projectId, apiKey)

			providerApp, err := client.CreateProviderApp(cmd.Context(), &request.CreateProviderAppParams{
				Provider:     args[0],
				ClientId:     clientId,
				ClientSecret: clientSecret,
				Scopes:       scopes,
			})
			if err != nil {
				if errors.Is(err, clerk.ErrNoSessions) {
					logger.Fatal("Authenticated session has expired, please log in using amp login")
				} else {
					logger.Fatal("Unable to create provider app. The API rejected the request")
				}
			}

			logger.Info(providerAppCreatedMessage(providerApp))
		},
	}

	cmd.Flags().StringVar(&clientId, "client-id", "", "OAuth client ID")
	cmd.Flags().StringArrayVar(&scopes, "scope", nil, "OAuth scope to register (repeat for multiple scopes)")
	cmd.Flags().BoolVar(&clientSecretStdin, "client-secret-stdin", false,
		"Read the OAuth client secret from stdin instead of prompting")

	err := cmd.MarkFlagRequired("client-id")
	if err != nil {
		logger.FatalErr("unable to require client-id flag", err)
	}

	return cmd
}

var createProviderAppCmd = newCreateProviderAppCmd( //nolint:gochecknoglobals
	func(projectId string, apiKey string) providerAppClient {
		return request.NewAPIClient(projectId, &apiKey)
	},
	flags.GetProjectOrFail,
)

func init() {
	rootCmd.AddCommand(createProviderAppCmd)
}
