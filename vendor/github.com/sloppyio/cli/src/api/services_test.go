package api

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestServicesList(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letchats/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"status":"success", "data":{"project":"letschat","services":[{"id": "frontend"},{"id":"backend"}]}}`)
	})

	services, _, _ := client.Services.List("letchats")

	want := []*Service{
		{
			ID: String("frontend"),
		},
		{
			ID: String("backend"),
		},
	}

	if !reflect.DeepEqual(services, want) {
		t.Errorf("List %+v, want %+v", services, want)
	}
}

func TestServicesGet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat/services/frontend", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"status":"success", "data":{"id":"frontend"}}`)
	})

	service, _, _ := client.Services.Get("letschat", "frontend")

	want := &Service{
		ID: String("frontend"),
	}

	if !reflect.DeepEqual(service, want) {
		t.Errorf("Get(%v) %+v, want %+v", "letschat/frontend", service, want)
	}
}

func TestServicesDelete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat/services/frontend", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		fmt.Fprint(w, `{"status":"success","message":"Service letschat/frontend successfully deleted."}`)
	})

	result, _, _ := client.Services.Delete("letschat", "frontend", false)

	want := &StatusResponse{
		Status:  "success",
		Message: "Service letschat/frontend successfully deleted.",
	}

	if !reflect.DeepEqual(result, want) {
		t.Errorf("Delete(%v) %+v, want %+v", "letschat/frontend", result, want)
	}
}

func TestServicesLogs(t *testing.T) {
	setup()
	defer teardown()

	testRegisterMockLogHandler(t, "/apps/letschat/services/frontend/logs")

	logCh, errCh := client.Services.GetLogs("letschat", "frontend", 5)

	testLogOutput(t, logCh, errCh)
}

func TestServicesLogs_invalidJSONBody(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat/services/frontend/logs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`aaaa`))
	})

	logCh, errCh := client.Services.GetLogs("letschat", "fronted", 0)
	for {
		select {
		case err := <-errCh:
			if err == nil {
				t.Errorf("Expected JSON parse error: %v", err)
			}
			return
		case log, ok := <-logCh:
			if ok {
				t.Errorf("Unexpected log entry: %v", log)
			}
		}
	}
}

func TestServicesURLParseErrors(t *testing.T) {
	setup()
	defer teardown()

	var urlTests = []struct {
		call func() error
	}{
		{
			call: func() error {
				_, _, err := client.Services.List("%")
				return err
			},
		},
		{
			call: func() error {
				_, _, err := client.Services.Get("%", "%")
				return err
			},
		},
		{
			call: func() error {
				_, _, err := client.Services.Delete("%", "%", true)
				return err
			},
		},
		{
			call: func() error {
				_, errs := client.Services.GetLogs("%", "%", 0)
				return <-errs
			},
		},
	}

	for _, tt := range urlTests {
		testURLParseError(t, tt.call())
	}
}

func TestServiceServerErrors(t *testing.T) {
	var serverErrorTests = []struct {
		uri    string
		method string
		call   func() (*http.Response, error)
		err    *ErrorResponse
	}{
		{
			uri:    "/apps/letschat/",
			method: "GET",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Services.List("letschat")
				return resp, err
			},
		},
		{
			uri:    "/apps/letschat/services/frontend",
			method: "GET",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Services.Get("letschat", "frontend")
				return resp, err
			},
		},
		{
			uri:    "/apps/letschat/services/frontend",
			method: "DELETE",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Services.Delete("letschat", "frontend", true)
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

var testServiceInput = []struct {
	input *Service
	want  string
}{
	{
		nil,
		"missing the required service",
	},
	{
		&Service{},
		"missing the required service.ID",
	},
	{
		&Service{
			ID: String("frontend"),
		},
		"missing the required service.Apps",
	},
}

func TestValidateService(t *testing.T) {
	for _, k := range testServiceInput {
		err := validateService(k.input)
		if err == nil {
			t.Errorf("Expected error to be returned")
		}
		if err.Error() != k.want {
			t.Errorf("Unexpected error to be returned: %v, want %v", err, k.want)
		}
	}
}
