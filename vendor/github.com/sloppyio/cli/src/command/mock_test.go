package command

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

type testClient struct {
	wantUsername string
	wantPassword string
	wantMessage  string
}

func NewErrorResponse(statusCode int, message, reason string) *api.ErrorResponse {
	return &api.ErrorResponse{
		Response: &http.Response{
			StatusCode: statusCode,
			Status:     http.StatusText(statusCode),
			Request:    &http.Request{},
		},
		StatusResponse: api.StatusResponse{
			Status:  "error",
			Message: message,
		},
		Reason: reason,
	}
}

// MockProjectsEndpoint
type mockProjectsEndpoint struct {
	wantMessage string
	project     string
	input       *api.Project
}

func (m *mockProjectsEndpoint) Create(input *api.Project) (*api.Project, *http.Response, error) {
	m.input = input
	if err := api.ValidateProject(input); err != nil {
		return nil, nil, err
	}
	return input, nil, nil
}

func (m *mockProjectsEndpoint) List() ([]api.Project, *http.Response, error) {
	if m.wantMessage != "" {
		return nil, nil, NewErrorResponse(http.StatusNotFound, m.wantMessage, "")
	}

	result := []api.Project{
		*mockProject,
	}

	// Just return the input
	return result, nil, nil
}

func (m *mockProjectsEndpoint) Get(project string) (*api.Project, *http.Response, error) {
	if project != "letschat" {
		return nil, nil, NewErrorResponse(http.StatusNotFound,
			fmt.Sprintf("Project with id \"%s\" could not be found", project), "")
	}

	result := mockProject

	// Just return the input
	return result, nil, nil
}

func (m *mockProjectsEndpoint) Update(project string, input *api.Project, force bool) (*api.Project, *http.Response, error) {
	if project != "letschat" {
		return nil, nil, NewErrorResponse(http.StatusNotFound, fmt.Sprintf("Project with id \"%s\" could not be found", project), "")
	}

	result := mockProject
	m.input = input

	// Just return the input
	return result, nil, nil
}

func (m *mockProjectsEndpoint) Delete(project string, force bool) (*api.StatusResponse, *http.Response, error) {
	if project != "letschat" {
		return nil, nil, NewErrorResponse(http.StatusNotFound,
			fmt.Sprintf("Project with id \"%s\" could not be found", project), "")
	}

	result := &api.StatusResponse{
		Status:  "success",
		Message: "Project letschat successfully deleted.",
	}

	// Just return the input
	return result, nil, nil
}

func (m *mockProjectsEndpoint) GetLogs(project string, limit int) (<-chan api.LogEntry, <-chan error) {
	errCh := make(chan error, 1)
	logCh := make(chan api.LogEntry)

	if project != "letschat" {
		errCh <- NewErrorResponse(http.StatusNotFound, fmt.Sprintf("Project with id \"%s\" could not be found", project), "")
		close(logCh)
		close(errCh)
		return logCh, errCh
	}

	go func() {
		logCh <- api.LogEntry{
			Project:   api.String("letschat"),
			Service:   api.String("frontend"),
			App:       api.String("node"),
			CreatedAt: &api.Timestamp{Time: time.Now()},
			Log:       api.String("1234"),
		}
		close(logCh)
		close(errCh)
	}()

	// Just return the input
	return logCh, errCh
}

// mockServicesEndpoint
type mockServicesEndpoint struct {
	wantMessage string
	input       *api.Service
}

func (m *mockServicesEndpoint) Get(project, service string) (*api.Service, *http.Response, error) {
	if project != "letschat" || service != "frontend" {
		return nil, nil, NewErrorResponse(http.StatusNotFound, fmt.Sprintf("Service with id \"%s\" could not be found", service), "")
	}

	result := mockProject.Services[0]

	// Just return the input
	return result, nil, nil
}

func (m *mockServicesEndpoint) Delete(project, service string, force bool) (*api.StatusResponse, *http.Response, error) {
	if project != "letschat" || service != "frontend" {
		return nil, nil, NewErrorResponse(http.StatusNotFound,
			fmt.Sprintf("Service with id \"%s\" could not be found", service), "")
	}

	result := &api.StatusResponse{
		Status:  "success",
		Message: "Service frontend successfully deleted.",
	}

	// Just return the input
	return result, nil, nil
}

func (m *mockServicesEndpoint) GetLogs(project, service string, limit int) (<-chan api.LogEntry, <-chan error) {
	errCh := make(chan error, 1)
	logCh := make(chan api.LogEntry)

	if project != "letschat" || service != "frontend" {
		errCh <- NewErrorResponse(http.StatusNotFound, fmt.Sprintf("Service with id \"%s\" could not be found", service), "")
		close(logCh)
		close(errCh)
		return logCh, errCh
	}

	go func() {
		logCh <- api.LogEntry{
			Project:   api.String("letschat"),
			Service:   api.String("frontend"),
			App:       api.String("node"),
			CreatedAt: &api.Timestamp{Time: time.Now()},
			Log:       api.String("1234"),
		}
		close(logCh)
		close(errCh)
	}()

	// Just return the input
	return logCh, errCh
}

// mockAppsEndpoint
type mockAppsEndpoint struct {
	wantMessage string
	input       *api.App
}

func (m *mockAppsEndpoint) Get(project, service, app string) (*api.App, *http.Response, error) {
	if project != "letschat" || service != "frontend" || app != "node" {
		return nil, nil, NewErrorResponse(http.StatusNotFound, fmt.Sprintf("App with id \"%s\" could not be found", app), "")
	}

	result := mockProject.Services[0].Apps[0]

	// Just return the input
	return result, nil, nil
}

func (m *mockAppsEndpoint) Restart(project, service, app string) (*api.StatusResponse, *http.Response, error) {
	if project != "letschat" || service != "frontend" || app != "node" {
		return nil, nil, NewErrorResponse(http.StatusNotFound, fmt.Sprintf("App with id \"%s\" could not be found", app), "")
	}

	result := &api.StatusResponse{
		Status:  "success",
		Message: "Restarting app.",
	}

	return result, nil, nil
}

func (m *mockAppsEndpoint) Rollback(project, service, app, version string) (*api.App, *http.Response, error) {
	if project != "letschat" || service != "frontend" || app != "node" {
		return nil, nil, NewErrorResponse(http.StatusNotFound, fmt.Sprintf("App with id \"%s\" could not be found", app), "")
	}
	result := mockProject.Services[0].Apps[0]
	return result, nil, nil
}

func (m *mockAppsEndpoint) Scale(project, service, app string, n int) (*api.App, *http.Response, error) {
	if project != "letschat" || service != "frontend" || app != "node" {
		return nil, nil, NewErrorResponse(http.StatusNotFound, fmt.Sprintf("App with id \"%s\" could not be found", app), "")
	}
	result := mockProject.Services[0].Apps[0]
	return result, nil, nil
}

func (m *mockAppsEndpoint) GetLogs(project, service, app string, limit int) (<-chan api.LogEntry, <-chan error) {
	errCh := make(chan error, 1)
	logCh := make(chan api.LogEntry)

	if project != "letschat" || service != "frontend" || app != "node" {
		errCh <- NewErrorResponse(http.StatusNotFound, fmt.Sprintf("App with id \"%s\" could not be found", app), "")
		return logCh, errCh
	}

	// Just return the input
	return logCh, errCh
}

func (m *mockAppsEndpoint) Update(project, service, app string, input *api.App) (*api.App, *http.Response, error) {
	if project != "letschat" || service != "frontend" || app != "node" {
		return nil, nil, NewErrorResponse(http.StatusNotFound, fmt.Sprintf("App with id \"%s\" could not be found", app), "")
	}

	result := mockProject.Services[0].Apps[0]
	m.input = input

	// Just return the input
	return result, nil, nil
}

func (m *mockAppsEndpoint) Delete(project, service, app string, force bool) (*api.StatusResponse, *http.Response, error) {
	if project != "letschat" || service != "frontend" || app != "node" {
		return nil, nil, NewErrorResponse(http.StatusNotFound,
			fmt.Sprintf("App with id \"%s\" could not be found", app), "")
	}

	result := &api.StatusResponse{
		Status:  "success",
		Message: "App node successfully deleted.",
	}

	// Just return the input
	return result, nil, nil
}

func (m *mockAppsEndpoint) GetMetrics(project, service, app string) (api.Metrics, *http.Response, error) {
	if project != "letschat" || service != "frontend" || app != "node" {
		return nil, nil, NewErrorResponse(http.StatusNotFound, fmt.Sprintf("App with id \"%s\" could not be found", app), "")
	}

	result := api.Metrics{
		"container_memory_usage_bytes": api.Series{
			"USERNAME-letschat_frontend_node.59f7edf4-82ff-11e5-8ac1-56847afe9799": api.DataPoints{0: &api.Point{
				X: api.Timestamp{Time: time.Date(2015, 11, 4, 14, 17, 19, 0, time.UTC)},
				Y: api.Float64(134217728),
			},
			},
		},
		"container_volume_usage_percentage": api.Series{
			"USERNAME-letschat_frontend_node./var/www": api.DataPoints{0: &api.Point{
				X: api.Timestamp{Time: time.Date(2015, 11, 4, 14, 17, 19, 0, time.UTC)},
				Y: api.Float64(12.2),
			},
			},
			"USERNAME-letschat_frontend_node./var/test": api.DataPoints{0: &api.Point{
				X: api.Timestamp{Time: time.Date(2015, 11, 4, 14, 17, 19, 0, time.UTC)},
				Y: api.Float64(12.9),
			},
			},
		},
		"container_network_receive_bytes_per_second": api.Series{
			"USERNAME-letschat_frontend_node.59f7edf4-82ff-11e5-8ac1-56847afe9799-eth0": api.DataPoints{0: &api.Point{
				X: api.Timestamp{Time: time.Date(2015, 11, 4, 14, 17, 19, 0, time.UTC)},
				Y: api.Float64(5767168),
			},
			},
			"USERNAME-letschat_frontend_node.59f7edf4-82ff-11e5-8ac1-56847afe9799-ethwe": api.DataPoints{0: &api.Point{
				X: api.Timestamp{Time: time.Date(2015, 11, 4, 14, 17, 19, 0, time.UTC)},
				Y: api.Float64(97.792),
			},
			},
		},
		"container_network_transmit_bytes_per_second": api.Series{
			"USERNAME-letschat_frontend_node.59f7edf4-82ff-11e5-8ac1-56847afe9799-eth0": api.DataPoints{0: &api.Point{
				X: api.Timestamp{Time: time.Date(2015, 11, 4, 14, 17, 19, 0, time.UTC)},
				Y: api.Float64(152567808),
			},
			},
			"USERNAME-letschat_frontend_node.59f7edf4-82ff-11e5-8ac1-56847afe9799-ethwe": api.DataPoints{0: &api.Point{
				X: api.Timestamp{Time: time.Date(2015, 11, 4, 14, 17, 19, 0, time.UTC)},
				Y: api.Float64(108032),
			},
			},
		},
	}

	// Just return the input
	return result, nil, nil
}

var mockProject = &api.Project{
	Name: api.String("letschat"),
	Services: []*api.Service{
		{
			ID: api.String("frontend"),
			Apps: []*api.App{
				{
					ID: api.String("node"),
					Domain: &api.Domain{
						URI: api.String("letschat.sloppy.zone"),
					},
					Version:   api.String("2015-12-21T10:56:33.081Z"),
					Memory:    api.Int(1024),
					Image:     api.String("mikemichel/lets-chat"),
					Instances: api.Int(1),
					PortMappings: []*api.PortMap{
						{Port: api.Int(5000)},
					},
					Volumes: []*api.Volume{
						&api.Volume{
							Path: api.String("/var/www"),
							Size: api.String("8GB"),
						},
						&api.Volume{
							Path: api.String("/var/test"),
							Size: api.String("8GB"),
						},
					},
					EnvVars: map[string]string{
						"LCB_DATABASE_URI": "mongodb://...",
					},
				},
			},
		},
		{
			ID: api.String("backend"),
			Apps: []*api.App{
				{
					ID:        api.String("mongodb"),
					Memory:    api.Int(512),
					Image:     api.String("mikemichel/lets-chat"),
					Instances: api.Int(1),
					PortMappings: []*api.PortMap{
						{Port: api.Int(27017)},
					},
				},
			},
		},
	},
}

// mockAppsEndpoint
type mockRegistryCredentialsEndpoint struct {
	wantMessage string
}

func (m *mockRegistryCredentialsEndpoint) Upload(r io.Reader) (*api.StatusResponse, *http.Response, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	if string(data) != `{"auth":"success"}` {
		return nil, nil, NewErrorResponse(http.StatusBadRequest, "Unable to upload docker credentials", "")
	}

	return &api.StatusResponse{
		Status:  "success",
		Message: "Uploaded docker credentials",
	}, nil, nil
}

func (m *mockRegistryCredentialsEndpoint) Delete() (*api.StatusResponse, *http.Response, error) {

	return &api.StatusResponse{
		Status:  "success",
		Message: "Docker credentials removed",
	}, nil, nil
}

func (m *mockRegistryCredentialsEndpoint) Check() (*api.StatusResponse, *http.Response, error) {

	if m.wantMessage == api.ErrMissingAccessToken.Error() {
		return nil, nil, api.ErrMissingAccessToken
	} else if m.wantMessage != "" {
		return nil, nil, NewErrorResponse(http.StatusNotFound, m.wantMessage, "")
	}

	return &api.StatusResponse{
		Status:  "success",
		Message: "Docker credentials exist",
	}, nil, nil
}

func testCodeAndOutput(t *testing.T, ui *ui.MockUI, code, want int, message string) {
	out := ui.OutputWriter.String()
	err := ui.ErrorWriter.String()

	if want != code {
		t.Errorf("ExitCode = %d, want %d", code, want)
		t.Errorf("Output = %s", out)
		t.Errorf("Error = %s", err)
	}

	if code != 0 {
		if !strings.Contains(err, message) {
			t.Errorf("Output = %s", out)
			t.Errorf("Error = %s", err)
		}
	} else {
		if !strings.Contains(out, message) {
			t.Errorf("Output = %s", out)
			t.Errorf("Error = %s", err)
		}
	}
}
