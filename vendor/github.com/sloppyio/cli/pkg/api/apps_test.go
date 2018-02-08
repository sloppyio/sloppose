package api_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/sloppyio/cli/internal/test"
	"github.com/sloppyio/cli/pkg/api"
)

func TestAppsList(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler(
		[]byte(`{"status": "success", "data":{"service":"frontend","apps":[{"id": "apache"},{"id":"nginx"}]}}`),
		"/apps/letchats/services/frontend",
	)
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	apps, _, _ := client.Apps.List("letchats", "frontend")

	want := []*api.App{
		{
			ID: api.String("apache"),
		},
		{
			ID: api.String("nginx"),
		},
	}

	if !reflect.DeepEqual(apps, want) {
		t.Errorf("List %+v, want %+v", apps, want)
	}
}

func TestAppsGet(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler(
		[]byte(`{"status":"success", "data":{"id":"apache"}}`),
		"/apps/letschat/services/frontend/apps/apache",
	)
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	app, _, _ := client.Apps.Get("letschat", "frontend", "apache")

	want := &api.App{
		ID: api.String("apache"),
	}

	if !reflect.DeepEqual(app, want) {
		t.Errorf("Get(%v) %+v, want %+v", "letschat/frontend/apache", app, want)
	}
}

func TestAppsUpdate(t *testing.T) {
	helper := test.NewHelper(t)

	want := &api.App{
		ID:        api.String("apache"),
		Memory:    api.Int(128),
		Image:     api.String("wordpress"),
		Instances: api.Int(1),
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")

		if r.URL.Path != "/apps/letschat/services/frontend/apps/apache" {
			t.Error("wrong path")
		}

		body, _ := ioutil.ReadAll(r.Body)
		if got, want := string(body), `{"mem":128}`+"\n"; got != want {
			t.Errorf("Update body: %v, want %s", got, want)
		}
		json.NewEncoder(w).Encode(newStatusResponse(want))
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	input := &api.App{
		Memory: api.Int(128),
	}

	got, _, _ := client.Apps.Update("letschat", "frontend", "apache", input)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Update(%v) = %+v, want %+v", input, got, want)
	}
}

func TestAppsDelete(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) { // /apps/letschat/services/frontend/apps/apache
		testMethod(t, r, "DELETE")
		fmt.Fprint(w, `{"status":"success","message":"App letschat/frontend/apache successfully deleted."}`)
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	result, _, _ := client.Apps.Delete("letschat", "frontend", "apache", false)

	want := &api.StatusResponse{
		Status:  "success",
		Message: "App letschat/frontend/apache successfully deleted.",
	}

	if !reflect.DeepEqual(result, want) {
		t.Errorf("Delete(%v) %+v, want %+v", "letschat/frontend/apache", result, want)
	}
}

func TestAppsRestart(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) { // /apps/letschat/services/frontend/apps/apache/restart
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"status":"success","message":"Restarting app."}`)
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	result, _, _ := client.Apps.Restart("letschat", "frontend", "apache")

	want := &api.StatusResponse{
		Status:  "success",
		Message: "Restarting app.",
	}

	if !reflect.DeepEqual(result, want) {
		t.Errorf("Restart(%v) %+v, want %+v", "letschat/frontend/apache", result, want)
	}
}

func TestAppsScale(t *testing.T) {
	helper := test.NewHelper(t)
	want := &api.App{
		ID:        api.String("apache"),
		Memory:    api.Int(128),
		Image:     api.String("wordpress"),
		Instances: api.Int(2),
	}
	handler := func(w http.ResponseWriter, r *http.Request) { // /apps/letschat/services/frontend/apps/apache
		testMethod(t, r, "PATCH")

		body, _ := ioutil.ReadAll(r.Body)
		if got, want := string(body), `{"instances":2}`+"\n"; got != want {
			t.Errorf("Scale body: %v, want %s", got, want)
		}

		json.NewEncoder(w).Encode(newStatusResponse(want))
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	got, _, _ := client.Apps.Scale("letschat", "frontend", "apache", 2)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Scale(%v) = %+v, want %+v", 2, got, want)
	}
}

func TestAppsRollback(t *testing.T) {
	helper := test.NewHelper(t)
	want := &api.App{
		ID:        api.String("apache"),
		Memory:    api.Int(128),
		Image:     api.String("wordpress"),
		Instances: api.Int(2),
		Version:   api.String("2015-06-15T16:50:14.947Z"),
	}
	handler := func(w http.ResponseWriter, r *http.Request) { // /apps/letschat/services/frontend/apps/apache
		testMethod(t, r, "PATCH")

		body, _ := ioutil.ReadAll(r.Body)
		if got, want := string(body), `{"version":"2015-06-15T16:50:14.947Z"}`+"\n"; got != want {
			t.Errorf("Rollback body: %v, want %s", got, want)
		}
		json.NewEncoder(w).Encode(newStatusResponse(want))
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	got, _, _ := client.Apps.Rollback("letschat", "frontend", "apache", "2015-06-15T16:50:14.947Z")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Rollback(%v) = %+v, want %+v", 2, got, want)
	}
}

func TestAppsLogs_notFound(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPNotFoundHandler(
		[]byte(`{"status":"error", "message":"not found","reason":"some reason"}`),
		"/apps/letschat/services/frontend/apps/apache/logs")
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	_, errors := client.Apps.GetLogs("letschat", "frontend", "apache", 5, "", "")
	select {
	case err := <-errors:
		testErrorResponse(t, err, nil)
	}
}

func TestAppsLogs_invalidJSONBody(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler([]byte(`aaaa`), "/apps/letschat/services/frontend/apps/apache/logs")
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	logs, errors := client.Apps.GetLogs("letschat", "fronted", "apache", 0, "%", "%")
	var err error
	select {
	case err = <-errors:
		// pass
		return
	case log := <-logs:
		t.Errorf("Unexpected log entry: %v", log)
		return
	}
	t.Errorf("Expected JSON parse error. Got: %v", err)
}

func TestAppsGetMetrics(t *testing.T) {
	helper := test.NewHelper(t)
	handler := helper.NewHTTPTestHandler(
		[]byte(`{"status":"success","data":[{"metric":"container_memory_usage_bytes","values":[{"name":"USERNAME-letschat_backend_mongodb.59f7edf4-82ff-11e5-8ac1-56847afe9799","data":[{"x":1446646639934,"y":31244288}]}]}]}`),
		"/apps/letschat/services/frontend/apps/apache/stats")
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	metrics, _, _ := client.Apps.GetMetrics("letschat", "frontend", "apache")

	want := api.Metrics{
		"container_memory_usage_bytes": api.Series{
			"USERNAME-letschat_backend_mongodb.59f7edf4-82ff-11e5-8ac1-56847afe9799": api.DataPoints{
				&api.DataPoint{
					X: api.Timestamp{time.Date(2015, 11, 4, 14, 17, 19, 0, time.UTC)},
					Y: api.Float64(31244288),
				},
			},
		},
	}

	if !reflect.DeepEqual(metrics, want) {
		t.Errorf("GetMetrics(%v) %+v, want %+v", "letschat/frontend/apache", metrics, want)
	}
}

func TestAppsURLParseErrors(t *testing.T) {
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
				_, _, err := client.Apps.List("%", "%")
				return err
			},
		},
		{
			call: func() error {
				_, _, err := client.Apps.Get("%", "%", "%")
				return err
			},
		},
		{
			call: func() error {
				_, _, err := client.Apps.Update("%", "%", "%", nil)
				return err
			},
		},
		{
			call: func() error {
				_, _, err := client.Apps.Delete("%", "%", "%", true)
				return err
			},
		},
		{
			call: func() error {
				_, _, err := client.Apps.Restart("%", "%", "%")
				return err
			},
		},
		{
			call: func() error {
				_, _, err := client.Apps.Rollback("%", "%", "%", "v1")
				return err
			},
		},
		{
			call: func() error {
				_, _, err := client.Apps.Scale("%", "%", "%", 1)
				return err
			},
		},
		{
			call: func() error {
				_, err := client.Apps.GetLogs("%", "%", "%", 0, "%", "%")
				return <-err
			},
		},
		{
			call: func() error {
				_, _, err := client.Apps.GetMetrics("%", "%", "%")
				return err
			},
		},
	}

	for _, tt := range urlTests {
		testURLParseError(t, tt.call())
	}
}
