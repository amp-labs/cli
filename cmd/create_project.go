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

type projectClient interface {
	GetCurrentOrganization(ctx context.Context) (*request.Organization, error)
	CreateProject(ctx context.Context, params *request.CreateProjectParams) (*request.Project, error)
}

type projectClientFactory func(apiKey string) projectClient

func createProject(
	ctx context.Context,
	client projectClient,
	name string,
	appName string,
) (*request.Project, error) {
	if appName == "" {
		appName = name
	}

	// Project creation requires an org ID, but user info already identifies the signed-in builder's current org.
	organization, err := client.GetCurrentOrganization(ctx)
	if err != nil {
		return nil, err
	}

	return client.CreateProject(ctx, &request.CreateProjectParams{
		AppName: appName,
		Name:    name,
		OrgId:   organization.Id,
	})
}

func newCreateProjectCmd(newClient projectClientFactory) *cobra.Command {
	var appName string

	cmd := &cobra.Command{
		Use:   "create:project <name>",
		Short: "Create a project",
		Long: "Create a project in your Ampersand organization. " +
			"The application display name defaults to the project name.",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			apiKey := flags.GetAPIKey()
			client := newClient(apiKey)

			project, err := createProject(cmd.Context(), client, args[0], appName)
			if err != nil {
				if errors.Is(err, clerk.ErrNoSessions) {
					logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
				} else {
					logger.FatalErr("Unable to create project", err)
				}
			}

			logger.Infof("Project ID: %s, Project Name: %s, App Name: %s",
				project.Id, project.Name, project.AppName)
		},
	}

	cmd.Flags().StringVar(&appName, "app-name", "",
		"Application display name shown during connection flows (defaults to project name)")

	return cmd
}

var createProjectCmd = newCreateProjectCmd(func(apiKey string) projectClient { //nolint:gochecknoglobals
	return request.NewAPIClient("unknown", &apiKey)
})

func init() {
	rootCmd.AddCommand(createProjectCmd)
}
