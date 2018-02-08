package api

import (
	"fmt"
	"net/http"
	"strconv"
)

// AppsEndpoint handles communication with the app related
// methods of the sloppy API.
type AppsEndpoint struct {
	client *Client
}

// App represents a sloppy app.
type App struct {
	ID           *string           `json:"id,omitempty"`
	Status       []string          `json:"status,omitempty"`
	Domain       *Domain           `json:"domain,omitempty"`
	SSL          *bool             `json:"ssl,omitempty"`
	Memory       *int              `json:"mem,omitempty"`
	Instances    *int              `json:"instances,omitempty"`
	Image        *string           `json:"image,omitempty"`
	Version      *string           `json:"version,omitempty"`
	Versions     []string          `json:"versions,omitempty"`
	Command      *string           `json:"cmd,omitempty"`
	PortMappings []*PortMap        `json:"port_mappings,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	EnvVars      map[string]string `json:"env,omitempty"`
	Volumes      []*Volume         `json:"volumes,omitempty"`
	HealthChecks []*HealthCheck    `json:"health_checks,omitempty"`
	Logging      *Logging          `json:"logging,omitempty"`
}

// Returns the count how often the given status was found
func (a *App) StatusCount(s string) (n int) {
	for _, status := range a.Status {
		if status == s {
			n++
		}
	}
	return n
}

func (a *App) String() string {
	return Stringify(a)
}

// Domain represents sloppy domain.
type Domain struct {
	URI *string `json:"uri,omitempty"`
}

// PortMap represents a sloppy port map.
type PortMap struct {
	Port *int `json:"container_port,omitempty"`
}

// HealthCheck represents a sloppy health check.
type HealthCheck struct {
	Timeout              *int    `json:"timeout_seconds,omitempty"`
	Interval             *int    `json:"interval_seconds,omitempty"`
	MaxConsectiveFailure *int    `json:"max_consecutive_failures,omitempty"`
	Path                 *string `json:"path,omitempty"`
	Type                 *string `json:"type,omitempty"`
	GracePeriod          *int    `json:"grace_period_seconds,omitempty"`
}

// Volume represents a sloppy app volume.
type Volume struct {
	Path *string `json:"container_path,omitempty"`
	Size *string `json:"size,omitempty"`
}

// Logging represents a sloppy app logging configuration.
type Logging struct {
	Driver  *string           `json:"driver,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

// List returns apps of a given project and service.
func (a *AppsEndpoint) List(project, service string) ([]*App, *http.Response, error) {
	u := fmt.Sprintf("apps/%s/services/%s", project, service)
	req, err := a.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	serv := new(Service)
	resp, err := a.client.Do(req, serv)
	if err != nil {
		return nil, resp, err
	}

	return serv.Apps, resp, err
}

// Get fetches a sloppy app by id and project, service.
func (a *AppsEndpoint) Get(project, service, id string) (*App, *http.Response, error) {
	u := fmt.Sprintf("apps/%s/services/%s/apps/%s", project, service, id)
	req, err := a.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	app := new(App)
	resp, err := a.client.Do(req, app)
	if err != nil {
		return nil, resp, err
	}

	return app, resp, err
}

// Update changes a sloppy app.
func (a *AppsEndpoint) Update(project, service, id string, input *App) (*App, *http.Response, error) {
	u := fmt.Sprintf("apps/%s/services/%s/apps/%s", project, service, id)
	req, err := a.client.NewRequest("PATCH", u, input)
	if err != nil {
		return nil, nil, err
	}

	app := new(App)
	resp, err := a.client.Do(req, app)
	if err != nil {
		return nil, resp, err
	}

	return app, resp, err
}

// Delete deletes a sloppy app.
func (a *AppsEndpoint) Delete(project, service, app string, force bool) (*StatusResponse, *http.Response, error) {
	u := fmt.Sprintf("apps/%s/services/%s/apps/%s", project, service, app)
	req, err := a.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// Add force parameter.
	if force {
		values := req.URL.Query()
		values.Add("force", strconv.FormatBool(force))
		req.URL.RawQuery = values.Encode()
	}
	status := new(StatusResponse)
	resp, err := a.client.Do(req, status)
	if err != nil {
		return nil, resp, err
	}

	return status, resp, err
}

// Restart sends a restart request for a sloppy app.
func (a *AppsEndpoint) Restart(project, service, app string) (*StatusResponse, *http.Response, error) {
	u := fmt.Sprintf("apps/%s/services/%s/apps/%s/restart", project, service, app)
	req, err := a.client.NewRequest("POST", u, nil)
	if err != nil {
		return nil, nil, err
	}

	status := new(StatusResponse)
	resp, err := a.client.Do(req, status)
	if err != nil {
		return nil, resp, err
	}

	return status, resp, err
}

// Scale changes the running instances of a sloppy api.
func (a *AppsEndpoint) Scale(project, service, app string, n int) (*App, *http.Response, error) {
	input := &App{
		Instances: Int(n),
	}
	return a.Update(project, service, app, input)
}

// Rollback reverts a sloppy app to a previous version.
func (a *AppsEndpoint) Rollback(project, service, app, version string) (*App, *http.Response, error) {
	input := &App{
		Version: String(version),
	}
	return a.Update(project, service, app, input)
}

// GetLogs returns logs for specific app.
func (a *AppsEndpoint) GetLogs(project, service, app string, limit int, fromDate string, toDate string) (<-chan LogEntry, <-chan error) {
	u := fmt.Sprintf("apps/%s/services/%s/apps/%s/logs", project, service, app)

	return RetrieveLogs(a.client, u, limit, fromDate, toDate)
}

// GetMetrics fetches a sloppy app's stats by name and project, service.
func (a *AppsEndpoint) GetMetrics(project, service, app string) (Metrics, *http.Response, error) {
	u := fmt.Sprintf("apps/%s/services/%s/apps/%s/stats", project, service, app)
	req, err := a.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	metrics := make(Metrics, 5) // FIXME why 5?
	resp, err := a.client.Do(req, &metrics)
	if err != nil {
		return nil, resp, err
	}

	return metrics, resp, err
}

// validateApp checks whether app's attributes are missing.
func validateApp(app *App) error {
	if app == nil {
		return fmt.Errorf("missing the required app")
	}
	if app.ID == nil {
		return fmt.Errorf("missing the required app.ID")
	}
	if app.Image == nil {
		return fmt.Errorf("missing the required app.Image")
	}
	return nil
}

// AppsUpdater is an interface which provides the Update method.
type AppsUpdater interface {
	Update(project, service, id string, input *App) (*App, *http.Response, error)
}

// AppsDeleter is an interface which provides the Delete method.
type AppsDeleter interface {
	Delete(project, service, id string, force bool) (*StatusResponse, *http.Response, error)
}

// AppsLogger is an interface which provides the GetLogs method.
type AppsLogger interface {
	GetLogs(project, service, id string, limit int, fromDate string, toDate string) (<-chan LogEntry, <-chan error)
}

// AppsRestarter is an interface which provides the Restart method.
type AppsRestarter interface {
	Restart(project, service, id string) (*StatusResponse, *http.Response, error)
}

// AppsRollbacker is an interface which provides the Rollback method.
type AppsRollbacker interface {
	Rollback(project, service, id, version string) (*App, *http.Response, error)
}

// AppsScaler is an interface which provides the Scale method.
type AppsScaler interface {
	Scale(project, service, id string, n int) (*App, *http.Response, error)
}

// AppsGetter is an interface which provides the Get method.
type AppsGetter interface {
	Get(project, service, id string) (*App, *http.Response, error)
}

// AppsGetMetricer is an interface which provides the getMetrics method.
type AppsGetMetricer interface {
	GetMetrics(project, service, id string) (Metrics, *http.Response, error)
}
