// Package api provides a client for using the sloppy.io API.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func checkResponse(r *http.Response) error {
	if 200 <= r.StatusCode && r.StatusCode <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}

	if r.ContentLength != 0 {
		json.NewDecoder(r.Body).Decode(errorResponse)
	}
	return errorResponse
}

// StatusResponse represents common API response received by delete or
// restart requests.
type StatusResponse struct {
	Status  string          `json:"status,omitempty"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (s *StatusResponse) String() string {
	return fmt.Sprintf("%s: %s", s.Status, s.Message)
}

// ErrorResponse represents errors caused by an API request.
type ErrorResponse struct {
	Response *http.Response
	StatusResponse
	Reason string `json:"reason,omitempty"`
}

func (e *ErrorResponse) Error() string {
	if e.Response != nil {
		return fmt.Sprintf("\"%s %s\" %d %s; %s",
			e.Response.Request.Method,
			e.Response.Request.URL, e.Response.StatusCode,
			e.Message, e.Reason,
		)
	}
	return fmt.Sprintf("%s; %s", e.Message, e.Reason)
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}

// Int is a helper routine that allocates a new int value
// to store v and returns a pointer to it.
func Int(v int) *int {
	p := new(int)
	*p = v
	return p
}

// Float64 is a helper routine that allocates a new float64 value
// to store v and returns a pointer to it.
func Float64(v float64) *float64 {
	p := new(float64)
	*p = v
	return p
}

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}
