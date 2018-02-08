package api_test

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/sloppyio/cli/internal/test"
	"github.com/sloppyio/cli/pkg/api"
)

func TestLogEntryString(t *testing.T) {
	input := &api.LogEntry{
		Project:   api.String("letschat"),
		Service:   api.String("frontend"),
		App:       api.String("apache"),
		CreatedAt: &api.Timestamp{time.Date(2015, 12, 07, 20, 10, 0, 0, time.UTC)},
		Log:       api.String("WARN: 123"),
	}
	want := "2015-12-07 20:10:00 letschat frontend apache WARN: 123"

	if input.String() != want {
		t.Errorf("String(%v) = %s, want %s", input, input.String(), want)
	}
}

func TestTimestampUnmarshal(t *testing.T) {
	input := []byte(`1449519000000`)
	want := time.Date(2015, 12, 07, 20, 10, 0, 0, time.UTC)

	timestamp := api.Timestamp{}
	if err := timestamp.UnmarshalJSON(input); err != nil {
		t.Errorf("Unexpected error: %v\n", err)
	}
	if timestamp.String() != want.String() {
		t.Errorf("UnmarshalJSON: %v, want %v", timestamp.String(), want.String())
	}
}

func TestTimestampUnmarshal_invalidTimestamp(t *testing.T) {
	input := []byte(`abcd`)

	timestamp := api.Timestamp{}
	if err := timestamp.UnmarshalJSON(input); err == nil {
		t.Error("Expected error to returned")
	}
}

func TestRetrieveLogs(t *testing.T) {
	helper := test.NewHelper(t)
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Transfer-Encoding", "chunked")
		for i := 0; i < 10; i++ {
			w.Write([]byte(fmt.Sprintf(`{"body": "chunk-%d"}`, i+1)))
			time.Sleep(time.Second / 10)
		}
	}
	server := helper.NewAPIServer(handler)
	defer server.Close()
	client := helper.NewClient(server.Listener.Addr())
	client.SetAccessToken("testToken")

	logs, errors := api.RetrieveLogs(client, "/", 0, "", "")
	entries := make([]api.LogEntry, 0)
	for {
		select {
		case err := <-errors:
			if err != nil && err != io.EOF {
				t.Error(err)
			}
			return
		case entry, ok := <-logs:
			if !ok {
				if len(entries) != 10 {
					t.Errorf("Expected 10 entries, got: %d", len(entries))
				}
				return
			}
			entries = append(entries, entry)
		}
	}
}
