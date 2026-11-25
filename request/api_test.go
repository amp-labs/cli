package request

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	testAPIKey    = "test-api-key"
	testOrgID     = "org-id"
	testProjectID = "project-id"
)

func newTestAPIClient(server *httptest.Server) *APIClient {
	apiKey := testAPIKey

	return &APIClient{
		Root:   server.URL + "/v1",
		APIKey: &apiKey,
		Client: &Client{Client: server.Client()},
	}
}

func TestGetCurrentOrganization(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/v1/my-info" {
			t.Errorf("request = %s %s, want GET /v1/my-info", req.Method, req.URL.Path)
		}

		if req.Header.Get("X-Api-Key") != testAPIKey {
			t.Errorf("X-Api-Key = %q, want test-api-key", req.Header.Get("X-Api-Key"))
		}

		writer.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(writer).Encode(map[string]any{
			"orgRole": map[string]any{
				"org": map[string]string{
					"id":    testOrgID,
					"label": "Test Organization",
				},
			},
		})
		if err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	organization, err := newTestAPIClient(server).GetCurrentOrganization(t.Context())
	if err != nil {
		t.Fatalf("get current organization: %v", err)
	}

	if organization.Id != testOrgID || organization.Label != "Test Organization" {
		t.Fatalf("organization = %#v", organization)
	}
}

func TestGetCurrentOrganizationRequiresOrganization(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(writer).Encode(map[string]any{})
		if err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	_, err := newTestAPIClient(server).GetCurrentOrganization(t.Context())
	if !errors.Is(err, ErrNoCurrentOrganization) {
		t.Fatalf("error = %v, want ErrNoCurrentOrganization", err)
	}
}

func TestCreateProject(t *testing.T) {
	t.Parallel()

	wantParams := CreateProjectParams{
		AppName: "Skills Onboarding CLI",
		Name:    "skills-onboarding-cli-dev",
		OrgId:   testOrgID,
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost || req.URL.Path != "/v1/projects" {
			t.Errorf("request = %s %s, want POST /v1/projects", req.Method, req.URL.Path)
		}

		if req.Header.Get("X-Api-Key") != testAPIKey {
			t.Errorf("X-Api-Key = %q, want test-api-key", req.Header.Get("X-Api-Key"))
		}

		var params CreateProjectParams

		err := json.NewDecoder(req.Body).Decode(&params)
		if err != nil {
			t.Errorf("decode request: %v", err)
		}

		if params != wantParams {
			t.Errorf("request body = %#v, want %#v", params, wantParams)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)

		err = json.NewEncoder(writer).Encode(Project{
			Id:         testProjectID,
			AppName:    params.AppName,
			Name:       params.Name,
			CreateTime: time.Date(2026, time.August, 18, 12, 0, 0, 0, time.UTC),
			OrgId:      params.OrgId,
		})
		if err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	project, err := newTestAPIClient(server).CreateProject(t.Context(), &wantParams)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	if project.Id != testProjectID || project.Name != wantParams.Name || project.AppName != wantParams.AppName {
		t.Fatalf("project = %#v", project)
	}
}
