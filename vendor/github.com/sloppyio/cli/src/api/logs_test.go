package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestLogEntryString(t *testing.T) {
	input := &LogEntry{
		Project:   String("letschat"),
		Service:   String("frontend"),
		App:       String("apache"),
		CreatedAt: &Timestamp{time.Date(2015, 12, 07, 20, 10, 0, 0, time.UTC)},
		Log:       String("WARN: 123"),
	}
	want := "2015-12-07 20:10:00 letschat frontend apache WARN: 123"

	if input.String() != want {
		t.Errorf("String(%v) = %s, want %s", input, input.String(), want)
	}
}

func TestTimestampUnmarshal(t *testing.T) {
	input := []byte(`1449519000000`)
	want := time.Date(2015, 12, 07, 20, 10, 0, 0, time.UTC)

	timestamp := Timestamp{}
	if err := timestamp.UnmarshalJSON(input); err != nil {
		t.Errorf("Unexpected error: %v\n", err)
	}
	if timestamp.String() != want.String() {
		t.Errorf("UnmarshalJSON: %v, want %v", timestamp.String(), want.String())
	}
}

func TestTimestampUnmarshal_invalidTimestamp(t *testing.T) {
	input := []byte(`abcd`)

	timestamp := Timestamp{}
	if err := timestamp.UnmarshalJSON(input); err == nil {
		t.Error("Expected error to returned")
	}
}

func testRegisterMockLogHandler(t *testing.T, urlStr string) {
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusOK)
		done := make(chan struct{})
		go func(w http.ResponseWriter) {
			for i := 0; i < 5; i++ {
				fmt.Fprintf(w, `{"service": "frontend-%d"}`+"\n", i)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				// Don't write everything at once.
				time.Sleep(100 * time.Microsecond)
			}
			close(done)
		}(w)
		<-done
	})

}

func testLogOutput(t *testing.T, logs <-chan LogEntry, errs <-chan error) {
	for i := 0; i < 5; i++ {
		want := fmt.Sprintf("frontend-%d", i)
		select {
		case log, ok := <-logs:
			if ok && *log.Service != want {
				t.Errorf("Log.Service = %v, want %v", *log.Service, want)
			}
		case err := <-errs:
			t.Errorf("Unexpected error: %v", err)
		}
	}
}
