package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestRegistryCredentialsCheck(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/registrycredentials", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")

		fmt.Fprintf(w, `{"status":"success", "message":"Docker credentials exist"}`)
	})

	status, _, _ := client.RegistryCredentials.Check()

	want := &StatusResponse{
		Status:  "success",
		Message: "Docker credentials exist",
	}

	if !reflect.DeepEqual(status, want) {
		t.Errorf("Check() %v, want %v", status, want)
	}
}

func TestRegistryCredentialsCheck_notFound(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/registrycredentials", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"status":"error", "message":"No credentials found"}`)
	})

	status, _, _ := client.RegistryCredentials.Check()

	want := &StatusResponse{
		Status:  "error",
		Message: "No credentials found",
	}

	if !reflect.DeepEqual(status, want) {
		t.Errorf("Check() = %v, want %v", status, want)
	}
}

func TestRegistryCredentialsCheck_serverError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/registrycredentials", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"status":"error", "message":"something happend"}`)
	})

	_, resp, err := client.RegistryCredentials.Check()

	want := newErrorResponse(resp, "something happend", "")

	if !reflect.DeepEqual(err, want) {
		t.Errorf("Check() = %v, want %v", err, want)
	}
}

func TestRegistryCredentialsDelete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/registrycredentials", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		fmt.Fprintf(w, `{"status":"success", "message":"Docker credentials removed"}`)
	})

	status, _, _ := client.RegistryCredentials.Delete()

	want := &StatusResponse{
		Status:  "success",
		Message: "Docker credentials removed",
	}

	if !reflect.DeepEqual(status, want) {
		t.Errorf("Delete() = %v, want %v", status, want)
	}
}

func TestRegistryCredentialsUpload(t *testing.T) {
	setup()
	defer teardown()

	input := `{"auths":{"https://index.docker.io/v1/":{"auth":"abc","email":"dev@example.com"}}}`

	mux.HandleFunc("/registrycredentials", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		defer r.Body.Close()

		if string(data) != input {
			t.Errorf("Wrong body = %s, want %s", string(data), input)
		}

		fmt.Fprintf(w, `{"status":"success", "message":"Uploaded docker credentials"}`)
	})

	reader := strings.NewReader(input)
	status, _, _ := client.RegistryCredentials.Upload(reader)

	want := &StatusResponse{
		Status:  "success",
		Message: "Uploaded docker credentials",
	}

	if !reflect.DeepEqual(status, want) {
		t.Errorf("Upload(%s) = %v, want %v", "test", status, want)
	}
}

func TestRegistryCredentialsUpload_serverErrors(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/registrycredentials", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"status":"error", "message":"Uploaded docker credentials failed"}`)
	})

	_, resp, err := client.RegistryCredentials.Upload(strings.NewReader(`{}`))

	want := newErrorResponse(resp, "Uploaded docker credentials failed", "")

	if !reflect.DeepEqual(err, want) {
		t.Errorf("Upload(%s) = %v, want %v", "test", err, want)
	}
}
