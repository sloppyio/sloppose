package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/sloppyio/cli/internal/test"
	"github.com/sloppyio/cli/pkg/api"
)

func TestProjectsList(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler(
		[]byte(`{"status": "success", "data":[{"project":"letschat"}]}`),
		"/apps/",
	)
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	projects, _, _ := client.Projects.List()

	want := []api.Project{
		{
			Name: api.String("letschat"),
		},
	}

	if !reflect.DeepEqual(projects, want) {
		t.Errorf("List %+v, want %+v", projects, want)
	}
}

func TestProjectsGet(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler(
		[]byte(`{"status": "success", "data":{"project":"letschat"}}`),
		"/apps/letschat",
	)
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	project, _, _ := client.Projects.Get("letschat")

	want := &api.Project{
		Name: api.String("letschat"),
	}

	if !reflect.DeepEqual(project, want) {
		t.Errorf("Get(%v) %+v, want %+v", "letschat", project, want)
	}
}

func TestProjectsCreate(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		project := new(api.Project)
		if err := json.NewDecoder(r.Body).Decode(project); err != nil {
			t.Errorf("Returned unexpected error: %v", err)
		}

		json.NewEncoder(w).Encode(newStatusResponse(project))
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	project := &api.Project{
		Name: api.String("letschat"),
		Services: []*api.Service{
			{
				ID: api.String("frontend"),
				Apps: []*api.App{
					{
						ID:        api.String("apache"),
						Memory:    api.Int(512),
						Instances: api.Int(2),
						Image:     api.String("wordpress"),
					},
				},
			},
		},
	}

	proj, _, _ := client.Projects.Create(project)
	if !reflect.DeepEqual(proj, project) {
		t.Errorf("Create(%v) = %+v, want %+v", project, proj, project)
	}
}

func TestProjectsCreate_validate(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler([]byte{}, "/")
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	_, _, err := client.Projects.Create(nil)
	if err == nil {
		t.Error("Expected error to be returned")
	}
}

func TestProjectsUpdate(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) { // /apps/letschat
		testMethod(t, r, "PUT")
		project := new(api.Project)
		if err := json.NewDecoder(r.Body).Decode(project); err != nil {
			t.Errorf("Returned unexpected error: %v", err)
		}

		if r.URL.Query().Get("force") != "" {
			t.Errorf("Expect force flag to be false, got %v", r.URL.Query().Get("force"))
		}

		json.NewEncoder(w).Encode(newStatusResponse(project))
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	project := &api.Project{
		Name: api.String("letschat"),
		Services: []*api.Service{
			{
				ID: api.String("frontend"),
				Apps: []*api.App{
					{
						ID:        api.String("apache"),
						Memory:    api.Int(512),
						Instances: api.Int(2),
						Image:     api.String("wordpress"),
					},
				},
			},
		},
	}

	proj, _, _ := client.Projects.Update("letschat", project, false)
	if !reflect.DeepEqual(proj, project) {
		t.Errorf("Update(%v) = %+v, want %+v", project, proj, project)
	}
}

func TestProjectsUpdate_forceFlag(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) { // /apps/letschat
		testMethod(t, r, "PUT")
		project := new(api.Project)
		if err := json.NewDecoder(r.Body).Decode(project); err != nil {
			t.Errorf("Returned unexpected error: %v", err)
		}

		if r.URL.Query().Get("force") != "true" {
			t.Errorf("Expect force flag to be true, got %v", r.URL.Query().Get("force"))
		}

		json.NewEncoder(w).Encode(newStatusResponse(project))
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	project := &api.Project{
		Name: api.String("letschat"),
		Services: []*api.Service{
			{
				ID: api.String("frontend"),
				Apps: []*api.App{
					{
						ID:        api.String("apache"),
						Memory:    api.Int(512),
						Instances: api.Int(2),
						Image:     api.String("wordpress"),
					},
				},
			},
		},
	}

	proj, _, _ := client.Projects.Update("letschat", project, true)
	if !reflect.DeepEqual(proj, project) {
		t.Errorf("Update(%v) = %+v, want %+v", project, proj, project)
	}
}

func TestProjectsDelete(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		if got, want := r.URL.Query().Get("force"), "true"; got != want {
			t.Errorf("URL Query(%s): %v, want %v", r.URL.Query().Encode(), got, want)
		}
		fmt.Fprint(w, `{"status":"success","message":"Project letschat successfully deleted."}`)
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	status, _, _ := client.Projects.Delete("letschat", true)

	want := &api.StatusResponse{
		Status:  "success",
		Message: "Project letschat successfully deleted.",
	}

	if !reflect.DeepEqual(status, want) {
		t.Errorf("Delete(%v) %+v, want %+v", "letschat", status, want)
	}
}

func TestProjectsLogs_notFound(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPNotFoundHandler(
		[]byte(`{"status":"error","message":"something happend"}`),
		"/apps/letschat/logs")
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	logs, errors := client.Projects.GetLogs("letschat", 5, "", "")
	select {
	case err := <-errors:
		testErrorResponse(t, err, nil)
	case log := <-logs:
		t.Errorf("Unexpected log entry: %v", log)
	}
}

func TestProjectsLogs_invalidJSONBody(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler([]byte(`aaaa`), "/apps/letschat/logs")
	server := helper.NewAPIServer(handler)
	defer server.Close()

	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("test")

	logs, errors := client.Projects.GetLogs("letschat", 0, "", "")
	select {
	case err := <-errors:
		if !strings.HasPrefix(err.Error(), "invalid character") {
			t.Errorf("Expected JSON parse error: %v", err)
		}
	case log := <-logs:
		t.Errorf("Unexpected log entry: %v", log)
	}
}

var testProject = &api.Project{
	Name: api.String("letschat"),
	Services: []*api.Service{
		{
			ID: api.String("frontend"),
			Apps: []*api.App{
				{
					ID:        api.String("apache"),
					Memory:    api.Int(512),
					Instances: api.Int(2),
					Image:     api.String("wordpress"),
				},
			},
		},
	},
}

func TestProjectURLParseErrors(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler([]byte{}, "/")
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("someToken")

	var urlTests = []struct {
		call func() error
	}{
		{
			call: func() error {
				_, _, err := client.Projects.Get("%")
				return err
			},
		},
		{
			call: func() error {
				_, _, err := client.Projects.Update("%", testProject, false)
				return err
			},
		},
		{
			call: func() error {
				_, _, err := client.Projects.Delete("%", true)
				return err
			},
		},
		{
			call: func() error {
				_, err := client.Projects.GetLogs("%", 0, "%", "%")
				return <-err
			},
		},
	}

	for _, tt := range urlTests {
		testURLParseError(t, tt.call())
	}
}

func TestProjectsServerErrors(t *testing.T) {
	helper := test.NewHelper(t)
	type wantMethod struct {
		m      sync.Mutex
		method string
	}
	want := wantMethod{}
	handler := func(w http.ResponseWriter, r *http.Request) {
		want.m.Lock()
		testMethod(t, r, want.method)
		want.m.Unlock()
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"status": "error", "message": "something happend"}`)
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("someToken")

	var serverErrorTests = []struct {
		uri    string
		method string
		err    *api.ErrorResponse
		call   func() (*http.Response, error)
	}{
		{
			uri:    "/apps/",
			method: "GET",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Projects.List()
				return resp, err
			},
		},
		{
			uri:    "/apps/letschat",
			method: "GET",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Projects.Get("letschat")
				return resp, err
			},
		},
		{
			uri:    "/apps/",
			method: "POST",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Projects.Create(testProject)
				return resp, err
			},
		},
		{
			uri:    "/apps/letschat",
			method: "PUT",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Projects.Update("letschat", testProject, false)
				return resp, err
			},
		},
		{
			uri:    "/apps/letschat",
			method: "DELETE",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Projects.Delete("letschat", true)
				return resp, err
			},
		},
	}

	for _, tt := range serverErrorTests {
		func(uri, method string, call func() (*http.Response, error), errR *api.ErrorResponse) {
			want.m.Lock()
			want.method = method
			want.m.Unlock()
			resp, err := call()
			errR.Response = resp
			testErrorResponse(t, err, errR)
		}(tt.uri, tt.method, tt.call, tt.err)
	}
}

func TestValidateProject(t *testing.T) {
	t.SkipNow()
	var testProjectInput = []struct {
		input *api.Project
		want  string
	}{
		{
			nil,
			"missing the required project",
		},
		{
			&api.Project{},
			"missing the required project.Name",
		},
		{
			&api.Project{
				Name: api.String("Letschat"),
			},
			"missing the required project.Services",
		},
		{
			&api.Project{
				Name: api.String("Letschat"),
				Services: []*api.Service{
					{
						ID: api.String("frontend"),
					},
				},
			},
			"missing the required service.Apps",
		},
		{
			&api.Project{
				Name: api.String("Letschat"),
				Services: []*api.Service{
					{
						ID: api.String("frontend"),
						Apps: []*api.App{
							{
								ID: api.String("frontend"),
							},
						},
					},
				},
			},
			"missing the required app.Image",
		},
	}

	for _, k := range testProjectInput {
		err := api.ValidateProject(k.input)
		if err == nil {
			t.Errorf("Expected error to be returned")
		}
		if err.Error() != k.want {
			t.Errorf("Unexpected error to be returned: %v, want %v", err, k.want)
		}
	}
}
