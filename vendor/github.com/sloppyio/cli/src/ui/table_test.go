package ui

import (
	"bytes"
	"testing"

	"github.com/sloppyio/cli/src/api"
)

func TestListApp_completeApp(t *testing.T) {
	app := &api.App{
		ID:      api.String("letschat"),
		Version: api.String("2016-01-25T10:59:53.409Z"),
		Memory:  api.Int(512),
		Domain: &api.Domain{
			URI: api.String("test.sloppy.zone"),
		},
		Command: api.String("serve"),
		PortMappings: []*api.PortMap{
			&api.PortMap{
				Port: api.Int(80),
			},
		},
		Image:        api.String("sdelements/lets-chat"),
		Instances:    api.Int(2),
		Status:       []string{"running", "running"},
		Dependencies: []string{"/work-letschat/backend/mysql"},
		Versions: []string{
			"2016-01-25T10:59:53.409Z",
			"2016-01-25T10:55:53.409Z",
		},
		EnvVars: map[string]string{
			"LETSCHAT_DB_PASSWORD": "test",
			"LETSCHAT_DB_USER":     "user",
		},
		Volumes: []*api.Volume{
			&api.Volume{
				Path: api.String("/var/data"),
				Size: api.String("8GB"),
			},
			&api.Volume{
				Path: api.String("/var/db"),
				Size: api.String("16GB"),
			},
		},
	}
	var buf bytes.Buffer

	listApp(&buf, app)

	if buf.String() != testOutput {
		t.Errorf("Output(%v) =\n%v, want\n%v", app, buf.String(), testOutput)
	}
}

func TestListApp_minimalApp(t *testing.T) {
	app := &api.App{
		ID:        api.String("letschat"),
		Version:   api.String("2016-01-25T10:59:53.409Z"),
		Memory:    api.Int(512),
		Image:     api.String("sdelements/lets-chat"),
		Instances: api.Int(1),
		Status:    []string{"running"},
		Versions: []string{
			"2016-01-25T10:59:53.409Z",
		},
	}
	var buf bytes.Buffer

	listApp(&buf, app)

	if buf.String() != testMinimalOutput {
		t.Errorf("Output(%v) =\n%v, want\n%v", app, buf.String(), testMinimalOutput)
	}
}

var testOutput = `Application: 	 letschat
Version: 	 2016-01-25T10:59:53.409Z
Memory:		 2 x 512 MiB
Instances:	 2 / 2
Domain:		 test.sloppy.zone
Image:		 sdelements/lets-chat
Command:	 serve
Volumes:	 '/var/data' 8GB
		 '/var/db' 16GB
Ports:		 80
Dependencies:	 /work-letschat/backend/mysql
Environments:	 LETSCHAT_DB_PASSWORD="test"
		 LETSCHAT_DB_USER="user"
Versions:	 2016-01-25T10:59:53.409Z
		 2016-01-25T10:55:53.409Z` + "\n"

var testMinimalOutput = `Application: 	 letschat
Version: 	 2016-01-25T10:59:53.409Z
Memory:		 512 MiB
Instances:	 1 / 1
Domain:		 -
Image:		 sdelements/lets-chat
Command:	 -
Volumes:	 -
Ports:		 -
Dependencies:	 -
Environments:	 -
Versions:	 2016-01-25T10:59:53.409Z` + "\n"
