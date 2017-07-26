package api

import (
	"io"
	"net/http"
)

// RegistryCredentialsEndpoint handles communication with the registry credentails related methods of the sloppy API.
type RegistryCredentialsEndpoint struct {
	client *Client
}

// Check checks if docker credentials exist.
func (r *RegistryCredentialsEndpoint) Check() (*StatusResponse, *http.Response, error) {

	req, err := r.client.NewRequest("GET", "registrycredentials", nil)
	if err != nil {
		return nil, nil, err
	}

	status := new(StatusResponse)
	resp, err := r.client.Do(req, status)
	if resp.StatusCode == http.StatusNotFound {
		status = &(err.(*ErrorResponse)).StatusResponse
	}

	return status, resp, err
}

// Delete removes docker credentials.
func (r *RegistryCredentialsEndpoint) Delete() (*StatusResponse, *http.Response, error) {
	req, err := r.client.NewRequest("DELETE", "registrycredentials", nil)
	if err != nil {
		return nil, nil, err
	}

	status := new(StatusResponse)
	resp, err := r.client.Do(req, status)

	return status, resp, err
}

// Upload uploads docker credentials.
func (r *RegistryCredentialsEndpoint) Upload(reader io.Reader) (*StatusResponse, *http.Response, error) {

	req, err := r.client.NewRequest("PUT", "registrycredentials", reader)
	if err != nil {
		return nil, nil, err
	}

	status := new(StatusResponse)
	resp, err := r.client.Do(req, status)

	return status, resp, err
}

// RegistryCredentialsUploader is an interface which provides the Upload method.
type RegistryCredentialsUploader interface {
	Upload(reader io.Reader) (*StatusResponse, *http.Response, error)
}

// RegistryCredentialsCheckDeleter is an interface which combines delete and check.
type RegistryCredentialsCheckDeleter interface {
	Check() (*StatusResponse, *http.Response, error)
	Delete() (*StatusResponse, *http.Response, error)
}
