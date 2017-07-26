package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestAppsList(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letchats/services/frontend", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"status": "success", "data":{"service":"frontend","apps":[{"id": "apache"},{"id":"nginx"}]}}`)
	})

	apps, _, _ := client.Apps.List("letchats", "frontend")

	want := []*App{
		{
			ID: String("apache"),
		},
		{
			ID: String("nginx"),
		},
	}

	if !reflect.DeepEqual(apps, want) {
		t.Errorf("List %+v, want %+v", apps, want)
	}
}

func TestAppsGet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat/services/frontend/apps/apache", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"status":"success", "data":{"id":"apache"}}`)
	})

	app, _, _ := client.Apps.Get("letschat", "frontend", "apache")

	want := &App{
		ID: String("apache"),
	}

	if !reflect.DeepEqual(app, want) {
		t.Errorf("Get(%v) %+v, want %+v", "letschat/frontend/apache", app, want)
	}
}

func TestAppsUpdate(t *testing.T) {
	setup()
	defer teardown()

	want := &App{
		ID:        String("apache"),
		Memory:    Int(128),
		Image:     String("wordpress"),
		Instances: Int(1),
	}

	mux.HandleFunc("/apps/letschat/services/frontend/apps/apache", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")

		body, _ := ioutil.ReadAll(r.Body)
		if got, want := string(body), `{"mem":128}`+"\n"; got != want {
			t.Errorf("Update body: %v, want %s", got, want)
		}
		json.NewEncoder(w).Encode(newStatusResponse(want))
	})

	input := &App{
		Memory: Int(128),
	}

	got, _, _ := client.Apps.Update("letschat", "frontend", "apache", input)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Update(%v) = %+v, want %+v", input, got, want)
	}
}

func TestAppsDelete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat/services/frontend/apps/apache", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		fmt.Fprint(w, `{"status":"success","message":"App letschat/frontend/apache successfully deleted."}`)
	})

	result, _, _ := client.Apps.Delete("letschat", "frontend", "apache", false)

	want := &StatusResponse{
		Status:  "success",
		Message: "App letschat/frontend/apache successfully deleted.",
	}

	if !reflect.DeepEqual(result, want) {
		t.Errorf("Delete(%v) %+v, want %+v", "letschat/frontend/apache", result, want)
	}
}

func TestAppsRestart(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat/services/frontend/apps/apache/restart", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"status":"success","message":"Restarting app."}`)
	})

	result, _, _ := client.Apps.Restart("letschat", "frontend", "apache")

	want := &StatusResponse{
		Status:  "success",
		Message: "Restarting app.",
	}

	if !reflect.DeepEqual(result, want) {
		t.Errorf("Restart(%v) %+v, want %+v", "letschat/frontend/apache", result, want)
	}
}

func TestAppsScale(t *testing.T) {
	setup()
	defer teardown()

	want := &App{
		ID:        String("apache"),
		Memory:    Int(128),
		Image:     String("wordpress"),
		Instances: Int(2),
	}

	mux.HandleFunc("/apps/letschat/services/frontend/apps/apache", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")

		body, _ := ioutil.ReadAll(r.Body)
		if got, want := string(body), `{"instances":2}`+"\n"; got != want {
			t.Errorf("Scale body: %v, want %s", got, want)
		}
		json.NewEncoder(w).Encode(newStatusResponse(want))
	})

	got, _, _ := client.Apps.Scale("letschat", "frontend", "apache", 2)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Scale(%v) = %+v, want %+v", 2, got, want)
	}
}

func TestAppsRollback(t *testing.T) {
	setup()
	defer teardown()

	want := &App{
		ID:        String("apache"),
		Memory:    Int(128),
		Image:     String("wordpress"),
		Instances: Int(2),
		Version:   String("2015-06-15T16:50:14.947Z"),
	}

	mux.HandleFunc("/apps/letschat/services/frontend/apps/apache", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")

		body, _ := ioutil.ReadAll(r.Body)
		if got, want := string(body), `{"version":"2015-06-15T16:50:14.947Z"}`+"\n"; got != want {
			t.Errorf("Rollback body: %v, want %s", got, want)
		}
		json.NewEncoder(w).Encode(newStatusResponse(want))
	})

	got, _, _ := client.Apps.Rollback("letschat", "frontend", "apache", "2015-06-15T16:50:14.947Z")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Rollback(%v) = %+v, want %+v", 2, got, want)
	}
}

func TestAppsLogs(t *testing.T) {
	setup()
	defer teardown()

	testRegisterMockLogHandler(t, "/apps/letschat/services/frontend/apps/apache/logs")

	logCh, errCh := client.Apps.GetLogs("letschat", "frontend", "apache", 5)

	testLogOutput(t, logCh, errCh)
}

func TestAppsLogs_notFound(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat/services/frontend/apps/apache/logs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNotFound)

		w.Write([]byte(`{"status":"error","message":"Could not found"}`))
	})

	logCh, errCh := client.Apps.GetLogs("letschat", "frontend", "apache", 5)

	select {
	case log := <-logCh:
		t.Errorf("Unexpected log entry: %v", log)
	case err := <-errCh:
		testErrorResponse(t, err, nil)
	}
}

func TestAppsLogs_invalidJSONBody(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat/services/frontend/apps/apache/logs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`aaaa`))
	})

	logCh, errCh := client.Apps.GetLogs("letschat", "fronted", "apache", 0)
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

func TestAppsGetMetrics(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apps/letschat/services/frontend/apps/apache/stats", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
{"status":"success","data":[{"metric":"container_memory_usage_bytes","values":[{"name":"USERNAME-letschat_backend_mongodb.59f7edf4-82ff-11e5-8ac1-56847afe9799","data":[{"x":1446646639934,"y":31244288}]}]}]}`)
	})

	metrics, _, _ := client.Apps.GetMetrics("letschat", "frontend", "apache")

	want := Metrics{
		"container_memory_usage_bytes": Series{
			"USERNAME-letschat_backend_mongodb.59f7edf4-82ff-11e5-8ac1-56847afe9799": DataPoints{
				&Point{
					X: Timestamp{time.Date(2015, 11, 4, 14, 17, 19, 0, time.UTC)},
					Y: Float64(31244288),
				},
			},
		},
	}

	if !reflect.DeepEqual(metrics, want) {
		t.Errorf("GetMetrics(%v) %+v, want %+v", "letschat/frontend/apache", metrics, want)
	}
}

func TestAppsURLParseErrors(t *testing.T) {
	setup()
	defer teardown()

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
				_, errs := client.Apps.GetLogs("%", "%", "%", 0)
				return <-errs
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

func TestAppsServerErrors(t *testing.T) {
	var serverErrorTests = []struct {
		uri    string
		method string
		err    *ErrorResponse
		call   func() (*http.Response, error)
	}{
		{
			uri:    "/apps/letschat/services/frontend",
			method: "GET",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Apps.List("letschat", "frontend")
				return resp, err
			},
		},
		{
			uri:    "/apps/letschat/services/frontend/apps/apache",
			method: "GET",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Apps.Get("letschat", "frontend", "apache")
				return resp, err
			},
		},
		{
			uri:    "/apps/letschat/services/frontend/apps/apache",
			method: "PATCH",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Apps.Update("letschat", "frontend", "apache", nil)
				return resp, err
			},
		},
		{
			uri:    "/apps/letschat/services/frontend/apps/apache",
			method: "DELETE",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Apps.Delete("letschat", "frontend", "apache", true)
				return resp, err
			},
		},
		{
			uri:    "/apps/letschat/services/frontend/apps/apache/restart",
			method: "POST",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Apps.Restart("letschat", "frontend", "apache")
				return resp, err
			},
		},
		{
			uri:    "/apps/letschat/services/frontend/apps/apache",
			method: "PATCH",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Apps.Scale("letschat", "frontend", "apache", 1)
				return resp, err
			},
		},
		{
			uri:    "/apps/letschat/services/frontend/apps/apache/stats",
			method: "GET",
			err:    newErrorResponse(nil, "something happend", ""),
			call: func() (*http.Response, error) {
				_, resp, err := client.Apps.GetMetrics("letschat", "frontend", "apache")
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

var testAppInput = []struct {
	input *App
	want  string
}{
	{
		nil,
		"missing the required app",
	},
	{
		&App{},
		"missing the required app.ID",
	},
	{
		&App{
			ID: String("frontend"),
		},
		"missing the required app.Image",
	},
}

func TestValidateApp(t *testing.T) {
	for _, k := range testAppInput {
		err := validateApp(k.input)
		if err == nil {
			t.Errorf("Expected error to be returned")
		}
		if err.Error() != k.want {
			t.Errorf("Unexpected error to be returned: %v, want %v", err, k.want)
		}
	}
}
