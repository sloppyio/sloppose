package api

import (
	"fmt"
	"net/http"
	"strconv"
)

// ServicesEndpoint handles communication with the service related
// methods of the sloppy API.
type ServicesEndpoint struct {
	client *Client
}

// Service represents a sloppy service.
type Service struct {
	ID   *string `json:"id,omitempty"`
	Apps []*App  `json:"apps,omitempty"`
}

func (s *Service) String() string {
	return Stringify(s)
}

// List returns services of a given project.
func (s *ServicesEndpoint) List(project string) ([]*Service, *http.Response, error) {
	u := fmt.Sprintf("apps/%s/", project)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	proj := new(Project)
	resp, err := s.client.Do(req, proj)
	if err != nil {
		return nil, resp, err
	}

	return proj.Services, resp, err
}

// Get fetches a sloppy service by project and id.
func (s *ServicesEndpoint) Get(project, id string) (*Service, *http.Response, error) {
	u := fmt.Sprintf("apps/%s/services/%s", project, id)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	service := new(Service)
	resp, err := s.client.Do(req, service)
	if err != nil {
		return nil, resp, err
	}

	return service, resp, err
}

// Delete deletes a sloppy service by project and id.
func (s *ServicesEndpoint) Delete(project, id string, force bool) (*StatusResponse, *http.Response, error) {
	u := fmt.Sprintf("apps/%s/services/%s", project, id)
	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// Add force parameter
	values := req.URL.Query()
	values.Add("force", strconv.FormatBool(force))
	req.URL.RawQuery = values.Encode()

	status := new(StatusResponse)
	resp, err := s.client.Do(req, status)
	if err != nil {
		return nil, resp, err
	}

	return status, resp, err
}

// GetLogs returns logs for all apps included in a specific project.
func (s *ServicesEndpoint) GetLogs(project, service string, limit int, fromDate string, toDate string) (<-chan LogEntry, <-chan error) {
	u := fmt.Sprintf("apps/%s/services/%s/logs", project, service)

	return RetrieveLogs(s.client, u, limit, fromDate, toDate)
}

// validateService checks whether service's attributes are missing.
func validateService(service *Service) error {
	if service == nil {
		return fmt.Errorf("missing the required service")
	}
	if service.ID == nil {
		return fmt.Errorf("missing the required service.ID")
	}
	if service.Apps == nil {
		return fmt.Errorf("missing the required service.Apps")
	}
	return nil
}

// ServicesGetter is an interface which provides the Get method.
type ServicesGetter interface {
	Get(project, id string) (*Service, *http.Response, error)
}

// ServicesDeleter is an interface which provides the Delete method.
type ServicesDeleter interface {
	Delete(project, id string, force bool) (*StatusResponse, *http.Response, error)
}

// ServicesLogger is an interface which provides the getLogs method.
type ServicesLogger interface {
	GetLogs(project, id string, limit int, fromDate string, toDate string) (<-chan LogEntry, <-chan error)
}
