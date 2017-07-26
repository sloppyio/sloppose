// Package api provides a client for using the sloppy.io API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	libVersion      = "1.0.0"
	userAgent       = "go-sloppy/" + libVersion
	apiVersion      = "v1"
	defaultMIMEType = "application/json"

	// DefaultBaseURL is for default api requests.
	DefaultBaseURL = "https://api.sloppy.io/"
)

// Errors introduced by sloppy.io Client.
var (
	ErrMissingAccessToken = errors.New("missing access token")
)

// A Client handles communication with the sloppy.io API.
type Client struct {
	// HTTP client which is actually used to communicate with the API.
	client *http.Client

	// BaseURL for accessing the Sloppy.io API. BaseURL should always
	// have a trailing slash.
	baseURL *url.URL

	// User-Agent specified the HTTP header "user-agent".
	UserAgent string

	// AccessToken specified the AccessToken used for authentication.
	// Empty string means unauthorized.
	accessToken string

	// Timer is used to handle timeouts more granular.
	Timer *time.Timer

	// API endpoints
	Projects            *ProjectsEndpoint
	Services            *ServicesEndpoint
	Apps                *AppsEndpoint
	RegistryCredentials *RegistryCredentialsEndpoint
}

// NewClient returns a new Client to handle API requests.
func NewClient() *Client {
	client := &Client{
		client:    http.DefaultClient,
		UserAgent: userAgent,
	}

	client.SetBaseURL(DefaultBaseURL)

	client.Projects = &ProjectsEndpoint{client: client}
	client.Services = &ServicesEndpoint{client: client}
	client.Apps = &AppsEndpoint{client: client}
	client.RegistryCredentials = &RegistryCredentialsEndpoint{client: client}

	return client
}

// SetBaseURL sets client's baseURL and appends version path.
func (c *Client) SetBaseURL(urlStr string) error {
	if urlStr == "" {
		urlStr = DefaultBaseURL
	}

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	if apiVersion != "" {
		baseURL.Path += apiVersion + "/"
	}
	c.baseURL = baseURL

	return nil
}

// NewRequest returns a new Request given a method, URL, and a value pointed
// to by body. If a relative URL is provided in urlStr, it is resolved
// relative to the Client's BaseURL. Relative URLs should never have a
// preceding slash. If body is specified, body is JSON encoded and included
// as request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	// Timeout
	ctx, cancel := context.WithCancel(context.TODO())
	c.Timer = time.AfterFunc(2*time.Minute, func() {
		cancel()
	})

	relURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	u := c.baseURL.ResolveReference(relURL)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		if r, ok := body.(io.Reader); ok {
			_, err = io.Copy(buf, r)
		} else {
			err = json.NewEncoder(buf).Encode(body)
		}

		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Content-Type", defaultMIMEType)
	req.Header.Set("Accept", defaultMIMEType)

	if c.accessToken == "" {
		return nil, ErrMissingAccessToken
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	return req, err
}

// Do sends an API request and returns an API response. The API response is
// decoded and stored in the value pointed to by v, or returned as an error
// if an API error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err = checkResponse(resp); err != nil {
		return resp, err
	}

	if v != nil {
		switch t := v.(type) {
		case io.Writer:
			_, err = io.Copy(t, resp.Body)
		case *StatusResponse:
			err = json.NewDecoder(resp.Body).Decode(v)
		default:
			status := new(StatusResponse)
			if err := json.NewDecoder(resp.Body).Decode(status); err != nil {
				return resp, err
			}
			err = json.Unmarshal(status.Data, v)
		}
	}
	return resp, err
}

// SetAccessToken sets Client's access token.
func (c *Client) SetAccessToken(t string) {
	c.accessToken = t
}

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
