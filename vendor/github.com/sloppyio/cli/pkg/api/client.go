package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const defaultBaseURL = "https://api.sloppy.io/v1/"
const defaultTimeOut = 2 * time.Minute

var (
	ErrMissingAccessToken = errors.New(`Missing "SLOPPY_APITOKEN", please login by exporting your token https://admin.sloppy.io/account/tokens`)
)

// A Client handles communication with the sloppy.io API.
type Client struct {
	m sync.RWMutex

	// HTTP client which is actually used to communicate with the API.
	client *http.Client

	// BaseURL for accessing the Sloppy.io API.
	// BaseURL should always have a trailing slash.
	baseURL *url.URL

	// header contains all related fields for an api request
	header http.Header

	// Timer is used to handle timeouts more granular.
	//Timer *time.Timer

	// API endpoints
	Projects            *ProjectsEndpoint
	Services            *ServicesEndpoint
	Apps                *AppsEndpoint
	RegistryCredentials *RegistryCredentialsEndpoint
}

// NewClient returns a new Client to handle API requests.
func NewClient() *Client {
	const mimeType = "application/json"
	client := &Client{
		client: &http.Client{
			Timeout: defaultTimeOut,
		},
		header: http.Header{
			"Accept":       {mimeType},
			"Content-Type": {mimeType},
			"User-Agent":   {"sloppy-cli/dev"},
		},
	}

	client.SetBaseURL(defaultBaseURL)

	client.Projects = &ProjectsEndpoint{client: client}
	client.Services = &ServicesEndpoint{client: client}
	client.Apps = &AppsEndpoint{client: client}
	client.RegistryCredentials = &RegistryCredentialsEndpoint{client: client}

	return client
}

func (c *Client) GetBaseURL() string {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.baseURL.String()
}

// SetBaseURL sets client's baseURL and appends version path.
func (c *Client) SetBaseURL(u string) error {
	c.m.Lock()
	defer c.m.Unlock()
	baseURL, err := url.Parse(u)
	if err != nil {
		return err
	}
	c.baseURL = baseURL
	return nil
}

func (c *Client) GetHeader(h string) []string {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.header[h]
}

func (c *Client) SetUserAgent(ua string) {
	c.m.Lock()
	c.header.Set("User-Agent", ua)
	c.m.Unlock()
}

// NewRequest returns a new Request given a method, URL, and a value pointed
// to by body. If a relative URL is provided in urlStr, it is resolved
// relative to the Client's BaseURL. Relative URLs should never have a
// preceding slash. If body is specified, body is JSON encoded and included
// as request body.
func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	c.m.RLock()
	defer c.m.RUnlock()
	if c.header.Get("Authorization") == "" {
		return nil, ErrMissingAccessToken
	}
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

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
	req.Header = c.header
	return req, nil
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
		switch v.(type) {
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

// SetAccessToken sets Client's access token header.
func (c *Client) SetAccessToken(t string) {
	c.m.Lock()
	c.header.Set("Authorization", "Bearer "+t)
	c.m.Unlock()
}
