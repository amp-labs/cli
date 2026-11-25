package cmd

import (
	"context"
	"testing"

	"github.com/amp-labs/cli/request"
)

type fakeProjectClient struct {
	organization *request.Organization
	createdWith  *request.CreateProjectParams
}

func (f *fakeProjectClient) GetCurrentOrganization(_ context.Context) (*request.Organization, error) {
	return f.organization, nil
}

func (f *fakeProjectClient) CreateProject(
	_ context.Context,
	params *request.CreateProjectParams,
) (*request.Project, error) {
	f.createdWith = params

	return &request.Project{
		Id:      "project-id",
		Name:    params.Name,
		AppName: params.AppName,
		OrgId:   params.OrgId,
	}, nil
}

func TestCreateProjectCommandUsesCurrentOrganizationAndDefaultAppName(t *testing.T) {
	t.Parallel()

	client := &fakeProjectClient{
		organization: &request.Organization{Id: "org-id"},
	}
	cmd := newCreateProjectCmd(func(_ string) projectClient {
		return client
	})
	cmd.SetArgs([]string{"skills-onboarding-cli-dev"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute create project command: %v", err)
	}

	want := request.CreateProjectParams{
		AppName: "skills-onboarding-cli-dev",
		Name:    "skills-onboarding-cli-dev",
		OrgId:   "org-id",
	}
	if client.createdWith == nil || *client.createdWith != want {
		t.Fatalf("create project params = %#v, want %#v", client.createdWith, want)
	}
}

func TestCreateProjectCommandUsesAppNameFlag(t *testing.T) {
	t.Parallel()

	client := &fakeProjectClient{
		organization: &request.Organization{Id: "org-id"},
	}
	cmd := newCreateProjectCmd(func(_ string) projectClient {
		return client
	})
	cmd.SetArgs([]string{"skills-onboarding-cli-dev", "--app-name", "Skills Onboarding CLI"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute create project command: %v", err)
	}

	if client.createdWith == nil || client.createdWith.AppName != "Skills Onboarding CLI" {
		t.Fatalf("app name = %q, want %q", client.createdWith.AppName, "Skills Onboarding CLI")
	}
}
