package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/spf13/cobra"
)

var errInvalidInstallationConfig = errors.New("config file must contain a JSON object")

type installationCreator interface {
	CreateInstallation(
		ctx context.Context,
		integrationId string,
		params *request.CreateInstallationParams,
	) (*request.Installation, error)
}

type installationCreatorFactory func(projectId string, apiKey string) installationCreator

type installationProjectResolver func() string

func readInstallationConfig(path string) (json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var content map[string]json.RawMessage

	err = json.Unmarshal(data, &content)
	if err != nil || content == nil {
		return nil, errInvalidInstallationConfig
	}

	return data, nil
}

func newCreateInstallationCmd(
	newClient installationCreatorFactory,
	getProjectId installationProjectResolver,
) *cobra.Command {
	var (
		groupRef     string
		connectionId string
		configPath   string
	)

	cmd := &cobra.Command{
		Use:   "create:installation <integrationId>",
		Short: "Create an installation",
		Long:  "Create an installation from a JSON file containing the installation config content.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			content, err := readInstallationConfig(configPath)
			if err != nil {
				logger.FatalErr("Unable to read installation config", err)
			}

			projectId := getProjectId()
			client := newClient(projectId, flags.GetAPIKey())

			installation, err := client.CreateInstallation(
				cmd.Context(),
				args[0],
				&request.CreateInstallationParams{
					GroupRef:     groupRef,
					ConnectionId: connectionId,
					Config: request.CreateInstallationConfig{
						Content: content,
					},
				},
			)
			if err != nil {
				if errors.Is(err, clerk.ErrNoSessions) {
					logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
				} else {
					logger.FatalErr("Unable to create installation", err)
				}
			}

			logger.Infof("Installation ID: %s, Group Ref: %s", installation.Id, groupRef)
		},
	}

	cmd.Flags().StringVar(&groupRef, "group-ref", "", "Identifier for the group that owns the installation")
	cmd.Flags().StringVar(&connectionId, "connection-id", "", "Connection ID to use")
	cmd.Flags().StringVar(&configPath, "config", "", "Path to a JSON file containing config content")

	for _, flag := range []string{"group-ref", "connection-id", "config"} {
		err := cmd.MarkFlagRequired(flag)
		if err != nil {
			logger.FatalErr("unable to require "+flag+" flag", err)
		}
	}

	return cmd
}

var createInstallationCmd = newCreateInstallationCmd( //nolint:gochecknoglobals
	func(projectId string, apiKey string) installationCreator {
		return request.NewAPIClient(projectId, &apiKey)
	},
	flags.GetProjectOrFail,
)

func init() {
	rootCmd.AddCommand(createInstallationCmd)
}
