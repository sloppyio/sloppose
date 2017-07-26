package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestProjectsList(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"status": "success", "data":[{"project":"letschat"}]}`)
	})

	projects, _, _ := client.Projects.List()

	want := []Project{
		{
			Name: String("letschat"),
		},
	}

	if !reflect.DeepEqual(projects, want) {
		t.Errorf("List %+v, want %+v", projects, want)
	}
}

func TestProjectsGet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"status": "success", "data":{"project":"letschat"}}`)
	})

	project, _, _ := client.Projects.Get("letschat")

	want := &Project{
		Name: String("letschat"),
	}

	if !reflect.DeepEqual(project, want) {
		t.Errorf("Get(%v) %+v, want %+v", "letschat", project, want)
	}
}

func TestProjectsCreate(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		project := new(Project)
		if err := json.NewDecoder(r.Body).Decode(project); err != nil {
			t.Errorf("Returned unexpected error: %v", err)
		}

		json.NewEncoder(w).Encode(newStatusResponse(project))
	})

	project := &Project{
		Name: String("letschat"),
		Services: []*Service{
			{
				ID: String("frontend"),
				Apps: []*App{
					{
						ID:        String("apache"),
						Memory:    Int(512),
						Instances: Int(2),
						Image:     String("wordpress"),
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
	setup()
	defer teardown()

	_, _, err := client.Projects.Create(nil)
	if err == nil {
		t.Error("Expected error to be returned")
	}
}

func TestProjectsUpdate(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		project := new(Project)
		if err := json.NewDecoder(r.Body).Decode(project); err != nil {
			t.Errorf("Returned unexpected error: %v", err)
		}

		if r.URL.Query().Get("force") != "" {
			t.Errorf("Expect force flag to be false, got %v", r.URL.Query().Get("force"))
		}

		json.NewEncoder(w).Encode(newStatusResponse(project))
	})

	project := &Project{
		Name: String("letschat"),
		Services: []*Service{
			{
				ID: String("frontend"),
				Apps: []*App{
					{
						ID:        String("apache"),
						Memory:    Int(512),
						Instances: Int(2),
						Image:     String("wordpress"),
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
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		project := new(Project)
		if err := json.NewDecoder(r.Body).Decode(project); err != nil {
			t.Errorf("Returned unexpected error: %v", err)
		}

		if r.URL.Query().Get("force") != "true" {
			t.Errorf("Expect force flag to be true, got %v", r.URL.Query().Get("force"))
		}

		json.NewEncoder(w).Encode(newStatusResponse(project))
	})

	project := &Project{
		Name: String("letschat"),
		Services: []*Service{
			{
				ID: String("frontend"),
				Apps: []*App{
					{
						ID:        String("apache"),
						Memory:    Int(512),
						Instances: Int(2),
						Image:     String("wordpress"),
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

func TestProjectsUpdate_validate(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.Projects.Update("letschat", nil, false)
	if err == nil {
		t.Error("Expected error to be returned")
	}
}

func TestProjectsDelete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		if got, want := r.URL.Query().Get("force"), "true"; got != want {
			t.Errorf("URL Query(%s): %v, want %v", r.URL.Query().Encode(), got, want)
		}
		fmt.Fprint(w, `{"status":"success","message":"Project letschat successfully deleted."}`)
	})

	status, _, _ := client.Projects.Delete("letschat", true)

	want := &StatusResponse{
		Status:  "success",
		Message: "Project letschat successfully deleted.",
	}

	if !reflect.DeepEqual(status, want) {
		t.Errorf("Delete(%v) %+v, want %+v", "letschat", status, want)
	}
}

func TestProjectsLogs(t *testing.T) {
	setup()
	defer teardown()

	testRegisterMockLogHandler(t, "/apps/letschat/logs")

	logs, errs := client.Projects.GetLogs("letschat", 5)
	testLogOutput(t, logs, errs)
}

func TestProjectsLogs_notFound(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat/logs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"error","message":"something happend"}`))
	})

	logs, errs := client.Projects.GetLogs("letschat", 5)
	select {
	case log := <-logs:
		t.Errorf("Unexpected log entry: %v", log)
	case err := <-errs:
		testErrorResponse(t, err, nil)
	}

}

func TestProjectsLogs_invalidJSONBody(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat/logs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`aaaa`))
	})

	logs, errs := client.Projects.GetLogs("letschat", 0)
	for {
		select {
		case err := <-errs:
			if err == nil {
				t.Errorf("Expected JSON parse error: %v", err)
			}
			return
		case log, ok := <-logs:
			if ok {
				t.Errorf("Unexpected log entry: %v", log)
			}
		}
	}
}

var testProject = &Project{
	Name: String("letschat"),
	Services: []*Service{
		{
			ID: String("frontend"),
			Apps: []*App{
				{
					ID:        String("apache"),
					Memory:    Int(512),
					Instances: Int(2),
					Image:     String("wordpress"),
				},
			},
		},
	},
}

func TestProjectURLParseErrors(t *testing.T) {
	setup()
	defer teardown()

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
				_, errs := client.Projects.GetLogs("%", 0)
				return <-errs
			},
		},
	}

	for _, tt := range urlTests {
		testURLParseError(t, tt.call())
	}
}

func TestProjectsServerErrors(t *testing.T) {
	var serverErrorTests = []struct {
		uri    string
		method string
		err    *ErrorResponse
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
		func(uri, method string, call func() (*http.Response, error), errR *ErrorResponse) {
			setup()
			defer teardown()

			mux.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, method)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"status": "error", "message": "something happend"}`)
			})

			resp, err := call()
			errR.Response = resp
			testErrorResponse(t, err, errR)
		}(tt.uri, tt.method, tt.call, tt.err)
	}
}

func TestValidateProject(t *testing.T) {
	var testProjectInput = []struct {
		input *Project
		want  string
	}{
		{
			nil,
			"missing the required project",
		},
		{
			&Project{},
			"missing the required project.Name",
		},
		{
			&Project{
				Name: String("Letschat"),
			},
			"missing the required project.Services",
		},
		{
			&Project{
				Name: String("Letschat"),
				Services: []*Service{
					{
						ID: String("frontend"),
					},
				},
			},
			"missing the required service.Apps",
		},
		{
			&Project{
				Name: String("Letschat"),
				Services: []*Service{
					{
						ID: String("frontend"),
						Apps: []*App{
							{
								ID: String("frontend"),
							},
						},
					},
				},
			},
			"missing the required app.Image",
		},
	}

	for _, k := range testProjectInput {
		err := ValidateProject(k.input)
		if err == nil {
			t.Errorf("Expected error to be returned")
		}
		if err.Error() != k.want {
			t.Errorf("Unexpected error to be returned: %v, want %v", err, k.want)
		}
	}
}
