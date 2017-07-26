package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the sloppy.io client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setup sets up a test HTTP server along with a sloppy.Client that is
// configured to talk to that test server. Tests should register handlers
// on mux which provide mock responses for the API method being tested.
func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient()

	client.SetAccessToken("setupToken")

	testURL, _ := url.Parse(server.URL)
	client.baseURL = testURL
}

// teardown to close the test server
func teardown() {
	server.Close()
}

func TestNewClient(t *testing.T) {
	client := NewClient()

	if got, want := client.baseURL.String(), DefaultBaseURL+apiVersion+"/"; got != want {
		t.Errorf("BaseURL = %v, want %v", got, want)
	}

	if got, want := client.UserAgent, userAgent; got != want {
		t.Errorf("UserAgent = %v, want %v", got, want)
	}
}

func TestNewRequest(t *testing.T) {
	client := NewClient()
	client.SetAccessToken("testToken")

	// test input, output
	inURL, outURL := "bar/", DefaultBaseURL+apiVersion+"/bar/"
	var inBody, outBody = struct {
		Bar *string
	}{
		Bar: String("Baz"),
	}, `{"Bar":"Baz"}` + "\n"

	req, _ := client.NewRequest("GET", inURL, inBody)

	if got, want := req.URL.String(), outURL; got != want {
		t.Errorf("%v: URL = %v, want %v", inURL, got, want)
	}

	body, _ := ioutil.ReadAll(req.Body)
	if got, want := string(body), outBody; got != want {
		t.Errorf("%+v: Body = %v, want %v", inBody, got, want)
	}

	if got, want := req.Header.Get("User-Agent"), client.UserAgent; got != want {
		t.Errorf("UserAgent = %v, want %v", got, want)
	}

	if got, want := req.Header.Get("Accept"), defaultMIMEType; got != want {
		t.Errorf("Accept = %v, want %v", got, want)
	}

	if got, want := req.Header.Get("Authorization"), "Bearer testToken"; got != want {
		t.Errorf("Authorization = %v, want %v", got, want)
	}
}

func TestNewRequest_notAuthenticated(t *testing.T) {
	client := NewClient()

	if _, err := client.NewRequest("GET", "/", nil); err == nil {
		t.Errorf("Expect error to be returned")
	} else if err != ErrMissingAccessToken {
		t.Errorf("Error: %v, want %s", err, ErrMissingAccessToken)
	}
}

func TestNewRequest_invalidURL(t *testing.T) {
	client := NewClient()

	_, err := client.NewRequest("GET", "%", nil)
	testURLParseError(t, err)
}

func TestNewRequest_invalidJSON(t *testing.T) {
	client := NewClient()

	type T struct {
		A map[struct{}]interface{}
	}

	_, err := client.NewRequest("GET", "/", &T{})
	if err == nil {
		t.Error("Expected error to be returned")
	}
	if err, ok := err.(*json.UnsupportedTypeError); !ok {
		t.Errorf("Expected JSON marshal error, got %+v", err)
	}
}

type errorReadWriter struct{}

func (w *errorReadWriter) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("errorReadWriter always returns an error")
}
func (w *errorReadWriter) Write(d []byte) (int, error) {
	return w.Read(d)
}

func TestNewRequest_withErrorReadWriter(t *testing.T) {
	client := NewClient()
	if _, err := client.NewRequest("GET", "", &errorReadWriter{}); err == nil {
		t.Error("Expected error to be returned")
	}
}

func TestDo(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"status":"success", "data":{"Bar":"Baz"}}`)
	})

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

func TestDo_withWriter(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `abc`)
	})

	req, err := client.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	var buf bytes.Buffer
	client.Do(req, &buf)

	body, err := ioutil.ReadAll(&buf)
	if err != nil {
		t.Fatalf("Expected no error: %v", err)
	}

	if want := `abc`; want != string(body) {
		t.Errorf("Response body = %q, want %q", body, want)
	}
}

func TestDo_withErrorReadWriter(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{}")
	})

	req, err := client.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if _, err := client.Do(req, &errorReadWriter{}); err == nil {
		t.Error("Expected error to be returned")
	}
}

func TestDo_errorResponse(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"status":"error", "message":"not found","reason":"some reason"}`)
	})

	req, err := client.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	resp, err := client.Do(req, nil)
	testErrorResponse(t, err, newErrorResponse(resp, "not found", "some reason"))
}

func TestDo_invalidToken(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"status":"error","message":"No valid token found.", "reason": "Check https://admin.sloppy.io/account/profile for your token."}`)
	})

	req, err := client.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resp, err := client.Do(req, nil)
	testErrorResponse(t, err, newErrorResponse(resp, "No valid token found.", "Check https://admin.sloppy.io/account/profile for your token."))
}

func TestDo_invalidJSONBody(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{b}`)
	})

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
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{}`)
	})

	if _, err := client.Do(&http.Request{}, nil); err == nil {
		t.Error("Expected error to be returned.")
	}
}

func TestStatusResponse(t *testing.T) {
	status := StatusResponse{Status: "Success", Message: "m"}
	if status.String() == "" {
		t.Error("Expected non-empty StatusResponse.String()")
	}
}

func TestErrorResponse(t *testing.T) {
	err := ErrorResponse{
		StatusResponse: StatusResponse{
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

func testErrorResponse(t *testing.T, err error, want *ErrorResponse) {
	switch err := err.(type) {
	case *ErrorResponse:
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

func newStatusResponse(v interface{}) *StatusResponse {
	data, err := json.Marshal(v)
	if err != nil {
		return &StatusResponse{Status: "error"}
	}
	return &StatusResponse{
		Status: "success",
		Data:   data,
	}
}

func newErrorResponse(r *http.Response, message, reason string) *ErrorResponse {
	return &ErrorResponse{
		Response: r,
		StatusResponse: StatusResponse{
			Status:  "error",
			Message: message,
		},
		Reason: reason,
	}
}
