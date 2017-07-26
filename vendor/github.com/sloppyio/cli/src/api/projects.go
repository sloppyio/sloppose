package api

import (
	"fmt"
	"net/http"
	"strconv"
)

// ProjectsEndpoint handles communication with the project related
// methods of the sloppy API.
type ProjectsEndpoint struct {
	client *Client
}

// Project represents a sloppy project.
type Project struct {
	Name     *string    `json:"project,omitempty"`
	Services []*Service `json:"services,omitempty"`
}

func (p *Project) String() string {
	return Stringify(p)
}

// List returns user's projects.
func (p *ProjectsEndpoint) List() ([]Project, *http.Response, error) {
	req, err := p.client.NewRequest("GET", "apps/", nil)
	if err != nil {
		// Probably skippable
		return nil, nil, err
	}

	var projects []Project
	resp, err := p.client.Do(req, &projects)
	if err != nil {
		return nil, resp, err
	}

	return projects, resp, err
}

// Get fetches a sloppy project by name.
func (p *ProjectsEndpoint) Get(name string) (*Project, *http.Response, error) {
	u := fmt.Sprintf("apps/%s", name)
	req, err := p.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	project := new(Project)
	resp, err := p.client.Do(req, project)
	if err != nil {
		return nil, resp, err
	}

	return project, resp, err
}

// Create creates a new sloppy project.
func (p *ProjectsEndpoint) Create(input *Project) (*Project, *http.Response, error) {
	if err := ValidateProject(input); err != nil {
		return nil, nil, err
	}

	req, err := p.client.NewRequest("POST", "apps/", input)
	if err != nil {
		// Probably skippable
		return nil, nil, err
	}

	project := new(Project)
	resp, err := p.client.Do(req, project)
	if err != nil {
		return nil, resp, err
	}

	return project, resp, err
}

// Update changes a sloppy project.
func (p *ProjectsEndpoint) Update(name string, input *Project, force bool) (*Project, *http.Response, error) {
	if err := ValidateProject(input); err != nil {
		return nil, nil, err
	}

	u := fmt.Sprintf("apps/%s", name)
	req, err := p.client.NewRequest("PUT", u, input)
	if err != nil {
		return nil, nil, err
	}

	// Add force parameter.
	if force {
		values := req.URL.Query()
		values.Add("force", strconv.FormatBool(force))
		req.URL.RawQuery = values.Encode()
	}

	project := new(Project)
	resp, err := p.client.Do(req, project)
	if err != nil {
		return nil, resp, err
	}

	return project, resp, err
}

// Delete deletes a sloppy project.
func (p *ProjectsEndpoint) Delete(name string, force bool) (*StatusResponse, *http.Response, error) {
	u := fmt.Sprintf("apps/%s", name)
	req, err := p.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// Add force parameter
	values := req.URL.Query()
	values.Add("force", strconv.FormatBool(force))
	req.URL.RawQuery = values.Encode()

	status := new(StatusResponse)
	resp, err := p.client.Do(req, status)
	if err != nil {
		return nil, resp, err
	}

	return status, resp, err
}

// GetLogs returns logs for all apps included in a specific project.
func (p *ProjectsEndpoint) GetLogs(name string, limit int) (<-chan LogEntry, <-chan error) {

	u := fmt.Sprintf("apps/%s/logs", name)

	return retrieveLogs(p.client, u, limit)
}

// ValidateProject checks whether project's attributes are missing.
func ValidateProject(project *Project) error {
	if project == nil {
		return fmt.Errorf("missing the required project")
	}
	if project.Name == nil {
		return fmt.Errorf("missing the required project.Name")
	}
	if project.Services == nil {
		return fmt.Errorf("missing the required project.Services")
	}

	for i := range project.Services {
		if err := validateService(project.Services[i]); err != nil {
			return err
		}

		for j := range project.Services[i].Apps {
			if err := validateApp(project.Services[i].Apps[j]); err != nil {
				return err
			}
		}
	}

	return nil
}

// ProjectsGetter is an interface which provides the Get method.
type ProjectsGetter interface {
	Get(name string) (*Project, *http.Response, error)
}

// ProjectsLister is an interface which provides the List method.
type ProjectsLister interface {
	List() ([]Project, *http.Response, error)
}

// ProjectsGetLister is an interface which provides the Get and List method.
type ProjectsGetLister interface {
	ProjectsGetter
	ProjectsLister
}

// ProjectsUpdater is an interface which provides the update method.
type ProjectsUpdater interface {
	Update(name string, input *Project, force bool) (*Project, *http.Response, error)
}

// ProjectsDeleter is an interface which provides the delete method.
type ProjectsDeleter interface {
	Delete(name string, force bool) (*StatusResponse, *http.Response, error)
}

// ProjectsLogger is an interface which provides the getLogs method.
type ProjectsLogger interface {
	GetLogs(project string, limit int) (<-chan LogEntry, <-chan error)
}

// ProjectsCreater is an interface which provides the Create method.
type ProjectsCreater interface {
	Create(input *Project) (*Project, *http.Response, error)
}
