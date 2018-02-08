package api

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"
)

// LogEntry represents a sloppy log entry.
type LogEntry struct {
	Project   *string    `json:"project,omitempty"`
	Service   *string    `json:"service,omitempty"`
	App       *string    `json:"app,omitempty"`
	CreatedAt *Timestamp `json:"createdAt,omitempty"`
	Log       *string    `json:"body,omitempty"`
}

// String prints a log entry
func (e *LogEntry) String() string {
	return fmt.Sprintf("%s %s %s %s %s",
		e.CreatedAt.Format("2006-01-02 15:04:05"), *e.Project, *e.Service, *e.App, *e.Log)
}

// Timestamp represents a sloppy timestamp.
type Timestamp struct {
	time.Time
}

// UnmarshalJSON decodes sloppy's date format.
func (u *Timestamp) UnmarshalJSON(data []byte) error {
	var aux int
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("timestamp not a number, got %s", string(data))
	}
	u.Time = time.Unix(int64(aux/1000), 0).UTC()
	return nil
}

func RetrieveLogs(c *Client, urlStr string, limit int, fromDate string, toDate string) (<-chan LogEntry, <-chan error) {
	logs := make(chan LogEntry)
	errors := make(chan error)

	go func() {
		defer close(logs)
		defer close(errors)

		req, err := c.NewRequest("GET", urlStr, nil)
		if err != nil {
			errors <- err
			return
		}

		// Add limit parameter
		if limit > 0 {
			values := req.URL.Query()
			values.Add("lines", strconv.Itoa(limit))
			req.URL.RawQuery = values.Encode()
		}
		if fromDate != "" {
			values := req.URL.Query()
			values.Add("from", fromDate)
			req.URL.RawQuery = values.Encode()
		}
		if toDate != "" {
			values := req.URL.Query()
			values.Add("to", toDate)
			req.URL.RawQuery = values.Encode()
		}
		resp, err := c.client.Do(req)
		if err != nil {
			errors <- err
			return
		}
		if err := checkResponse(resp); err != nil {
			errors <- err
			return
		}

		dec := json.NewDecoder(resp.Body)
		for {
			var log LogEntry
			err := dec.Decode(&log)
			if err != nil {
				if err == io.EOF {
					break
				}
				errors <- err
				return
			}
			logs <- log
		}
	}()

	return logs, errors
}
