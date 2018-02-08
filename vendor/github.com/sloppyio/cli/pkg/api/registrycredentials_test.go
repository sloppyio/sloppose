package api_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/sloppyio/cli/internal/test"
	"github.com/sloppyio/cli/pkg/api"
)

func TestRegistryCredentialsCheck(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"status":"success", "message":"Docker credentials exist"}`)
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()

	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("test")

	status, _, _ := client.RegistryCredentials.Check()

	want := &api.StatusResponse{
		Status:  "success",
		Message: "Docker credentials exist",
	}

	if !reflect.DeepEqual(status, want) {
		t.Errorf("Check() %v, want %v", status, want)
	}
}

func TestRegistryCredentialsCheck_notFound(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler(
		[]byte(`{"status":"error", "message":"No credentials found"}`),
		"/registrycredentials",
	)
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	status, _, _ := client.RegistryCredentials.Check()

	want := &api.StatusResponse{
		Status:  "error",
		Message: "No credentials found",
	}

	if !reflect.DeepEqual(status, want) {
		t.Errorf("Check() = %v, want %v", status, want)
	}
}

func TestRegistryCredentialsDelete(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		fmt.Fprint(w, `{"status":"success", "message":"Docker credentials removed"}`)
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	status, _, _ := client.RegistryCredentials.Delete()

	want := &api.StatusResponse{
		Status:  "success",
		Message: "Docker credentials removed",
	}

	if !reflect.DeepEqual(status, want) {
		t.Errorf("Delete() = %v, want %v", status, want)
	}
}

func TestRegistryCredentialsUpload(t *testing.T) {
	input := `{"auths":{"https://index.docker.io/v1/":{"auth":"abc","email":"dev@example.com"}}}`
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		defer r.Body.Close()

		if string(data) != input {
			t.Errorf("Wrong body = %s, want %s", string(data), input)
		}

		fmt.Fprint(w, `{"status":"success", "message":"Uploaded docker credentials"}`)
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	reader := strings.NewReader(input)
	status, _, _ := client.RegistryCredentials.Upload(reader)

	want := &api.StatusResponse{
		Status:  "success",
		Message: "Uploaded docker credentials",
	}

	if !reflect.DeepEqual(status, want) {
		t.Errorf("Upload(%s) = %v, want %v", "test", status, want)
	}
}
