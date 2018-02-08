package api_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/sloppyio/cli/internal/test"
	"github.com/sloppyio/cli/pkg/api"
)

func TestNewRequest(t *testing.T) {
	client := api.NewClient()
	client.SetAccessToken("testToken")

	// test input, output
	inURL, outURL := "bar/", fmt.Sprintf("%sbar/", client.GetBaseURL())
	var inBody, outBody = struct {
		Bar string
	}{
		Bar: "Baz",
	}, `{"Bar":"Baz"}` + "\n"

	req, _ := client.NewRequest("GET", inURL, inBody)

	if got, want := req.URL.String(), outURL; got != want {
		t.Errorf("in %s:\ngot:\t%s\nwant:\t%v\n", inURL, got, want)
	}

	body, _ := ioutil.ReadAll(req.Body)
	if got, want := string(body), outBody; got != want {
		t.Errorf("%+v: Body = %v, want %v", inBody, got, want)
	}

	if got, want := req.Header.Get("User-Agent"), client.GetHeader("User-Agent")[0]; got != want {
		t.Errorf("UserAgent = %v, want %v", got, want)
	}

	if got, want := req.Header.Get("Accept"), client.GetHeader("Accept")[0]; got != want {
		t.Errorf("Accept = %v, want %v", got, want)
	}

	if got, want := req.Header.Get("Authorization"), "Bearer testToken"; got != want {
		t.Errorf("Authorization = %v, want %v", got, want)
	}
}

func TestNewRequest_notAuthenticated(t *testing.T) {
	client := api.NewClient()
	_, err := client.NewRequest("GET", "/", nil)
	if err == nil {
		t.Error("Expect error to be returned")
	} else if err != nil && err != api.ErrMissingAccessToken {
		t.Errorf("Error: %v, want %v", err, api.ErrMissingAccessToken)
	}
}

func TestNewRequest_invalidURL(t *testing.T) {
	client := api.NewClient()
	client.SetAccessToken("someToken")

	_, err := client.NewRequest("GET", "%", nil)
	testURLParseError(t, err)
}

type errorReadWriter struct{}

func (w *errorReadWriter) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("errorReadWriter always returns an error")
}
func (w *errorReadWriter) Write(d []byte) (int, error) {
	return w.Read(d)
}

func TestNewRequest_withErrorReadWriter(t *testing.T) {
	client := api.NewClient()
	if _, err := client.NewRequest("GET", "", &errorReadWriter{}); err == nil {
		t.Error("Expected error to be returned")
	}
}

func TestDo(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler([]byte(`{"status":"success", "data":{"Bar":"Baz"}}`), "/")
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	type foo struct {
		Bar string
	}

	req, err := client.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	body := new(foo)
	client.Do(req, body)

	want := &foo{"Baz"}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("Response body = %v, want %v", body, want)
	}
}

func TestDo_errorResponse(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"status":"error", "message":"not found","reason":"some reason"}`)
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	req, err := client.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	resp, err := client.Do(req, nil)
	testErrorResponse(t, err, newErrorResponse(resp, "not found", "some reason"))
}

func TestDo_invalidToken(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"status":"error","message":"No valid token found.", "reason": "Check https://admin.sloppy.io/account/profile for your token."}`)
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	req, err := client.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resp, err := client.Do(req, nil)
	testErrorResponse(t, err, newErrorResponse(resp, "No valid token found.", "Check https://admin.sloppy.io/account/profile for your token."))
}

func TestDo_invalidJSONBody(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler([]byte(`{b}`), "/")
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	type T struct {
		A map[int]interface{}
	}

	req, err := client.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, err = client.Do(req, &T{}); err == nil {
		t.Error("Expected error to be returned.")
	}
}

func TestDo_noRequest(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler([]byte(`{}`), "/")
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	if _, err := client.Do(&http.Request{}, nil); err == nil {
		t.Error("Expected error to be returned.")
	}
}

func TestStatusResponse(t *testing.T) {
	status := api.StatusResponse{Status: "Success", Message: "m"}
	if status.String() == "" {
		t.Error("Expected non-empty StatusResponse.String()")
	}
}

func TestErrorResponse(t *testing.T) {
	err := api.ErrorResponse{
		StatusResponse: api.StatusResponse{
			Status: "error", Message: "m",
		},
		Reason: "r",
	}
	if err.Error() != "m; r" {
		t.Errorf("Error() = %s, want %s", err.Error(), "m; r")
	}

	err.Response = &http.Response{Request: &http.Request{Method: "GET"}}
	if err.Error() != "\"GET <nil>\" 0 m; r" {
		t.Errorf("Error() = %s, want %s", err.Error(), "\"GET <nil>\" 0 m; r")
	}
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if r.Method != want {
		t.Errorf("Method = %v, want %v", r.Method, want)
	}
}

func testErrorResponse(t *testing.T, err error, want *api.ErrorResponse) {
	switch err := err.(type) {
	case *api.ErrorResponse:
		if want == nil {
			return
		}
		if !reflect.DeepEqual(err, want) {
			t.Errorf("errorResponse = %+v, want %+v", err, want)
		}
	case nil:
		t.Error("Expected error to be returned")
	default:
		t.Errorf("Expected ErrorResponse error, got %+v", err)
	}
	return
}

func testURLParseError(t *testing.T, err error) {
	switch err := err.(type) {
	case *url.Error:
		if err.Op != "parse" {
			t.Errorf("Expected URL parse error, got %+v", err)
		}
		return
	case nil:
		t.Error("Expected error to be returned")
	default:
		t.Errorf("Expected URL parse error, got %+v", err)
	}
	return
}

func newStatusResponse(v interface{}) *api.StatusResponse {
	data, err := json.Marshal(v)
	if err != nil {
		return &api.StatusResponse{Status: "error"}
	}
	return &api.StatusResponse{
		Status: "success",
		Data:   data,
	}
}

func newErrorResponse(r *http.Response, message, reason string) *api.ErrorResponse {
	return &api.ErrorResponse{
		Response: r,
		StatusResponse: api.StatusResponse{
			Status:  "error",
			Message: message,
		},
		Reason: reason,
	}
}
