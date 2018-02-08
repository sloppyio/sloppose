package api_test

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/sloppyio/cli/internal/test"
	"github.com/sloppyio/cli/pkg/api"
)

func TestServicesList(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler(
		[]byte(`{"status":"success", "data":{"project":"letschat","services":[{"id": "frontend"},{"id":"backend"}]}}`),
		"/apps/letchats/")
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	services, _, _ := client.Services.List("letchats")

	want := []*api.Service{
		{
			ID: api.String("frontend"),
		},
		{
			ID: api.String("backend"),
		},
	}

	if !reflect.DeepEqual(services, want) {
		t.Errorf("List %+v, want %+v", services, want)
	}
}

func TestServicesGet(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler(
		[]byte(`{"status":"success", "data":{"id":"frontend"}}`),
		"/apps/letschat/services/frontend")
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	service, _, _ := client.Services.Get("letschat", "frontend")

	want := &api.Service{
		ID: api.String("frontend"),
	}

	if !reflect.DeepEqual(service, want) {
		t.Errorf("Get(%v) %+v, want %+v", "letschat/frontend", service, want)
	}
}

func TestServicesDelete(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) { // /apps/letschat/services/frontend
		testMethod(t, r, "DELETE")
		fmt.Fprint(w, `{"status":"success","message":"Service letschat/frontend successfully deleted."}`)
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	result, _, _ := client.Services.Delete("letschat", "frontend", false)

	want := &api.StatusResponse{
		Status:  "success",
		Message: "Service letschat/frontend successfully deleted.",
	}

	if !reflect.DeepEqual(result, want) {
		t.Errorf("Delete(%v) %+v, want %+v", "letschat/frontend", result, want)
	}
}

func TestServicesURLParseErrors(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler([]byte{}, "/")
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

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
				_, err := client.Services.GetLogs("%", "%", 0, "%", "%")
				return <-err
			},
		},
	}

	for _, tt := range urlTests {
		testURLParseError(t, tt.call())
	}
}

func TestServiceServerErrors(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"status": "error", "message": "something happend"}`)
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("access")

	var serverErrorTests = []struct {
		uri    string
		method string
		call   func() (*http.Response, error)
		err    *api.ErrorResponse
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
		func(uri, method string, call func() (*http.Response, error), errR *api.ErrorResponse) {
			resp, err := call()
			errR.Response = resp
			testErrorResponse(t, err, errR)
		}(tt.uri, tt.method, tt.call, tt.err)
	}
}
