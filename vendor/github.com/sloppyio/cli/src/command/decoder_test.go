package command

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/sloppyio/cli/src/api"
)

func TestDecode_invalidSyntaxCompactJSON(t *testing.T) {
	reader := strings.NewReader(`{"project":"apache","services":[{"id":"frontend","apps":[{"id":"apache","domain":{"type":"HTTP","uri":"sloppy-cli-testing.sloppy.zone"},"mem":124"image":"sloppy/apache-php","instances":1,"port_mappings":[{"container_port":80},{"container_port":443}]}]}]}`)

	d := newDecoder(reader, stringMap{})
	input := new(api.Project)

	if err := d.DecodeJSON(input); err == nil {
		t.Error("Expected json syntax error")
	} else if err.Error() != "got syntax error around line 1:146" {
		t.Errorf("Error = '%s', want 'got syntax error around line 1:146'", err.Error())
	}
}

func TestDecode_invalidSyntaxJSON(t *testing.T) {
	reader := strings.NewReader(`{
    "project": "apache",
    "services": [
        {
            "id": "frontend",
            "apps": [
                {
                    "id": "apache",
                    "domain": {
                        "uri": "sloppy-cli-testing.sloppy.zone"
                    },
                    "mem": 124
                    "image": "sloppy/apache-php",
                    "instances": 1,
                    "port_mappings": [
                        {
                            "container_port": 80
                        },
                        {
                            "container_port": 443
                        }
                    ]
                }
            ]
        }
    ]
}`)

	d := newDecoder(reader, stringMap{})
	input := new(api.Project)

	if err := d.DecodeJSON(input); err == nil {
		t.Error("Expected json syntax error")
	} else if err.Error() != "got syntax error around line 13:21" {
		t.Errorf("Error = '%s', want 'got syntax error around line 13:21'", err.Error())
	}
}

func TestDecode_invalidTypeCompactJSON(t *testing.T) {
	reader := strings.NewReader(`{"project":"apache","services":[{"id":"frontend","apps":[{"id":"apache","domain":{"uri":"sloppy-cli-testing.sloppy.zone"},"mem":"124","image":"sloppy/apache-php","instances":1,"port_mappings":[{"container_port":80},{"container_port":443}]}]}]}`)

	d := newDecoder(reader, stringMap{})
	input := new(api.Project)

	if err := d.DecodeJSON(input); err == nil {
		t.Error("Expected json type mismatch error")
	} else if err.Error() != "got type mismatch on line 1:133, expect number" {
		t.Errorf("Error = '%s', want 'got type mismatch on line 1:133, expect number'", err.Error())
	}
}

func TestDecode_invalidTypeJSON(t *testing.T) {
	reader := strings.NewReader(`{
    "project": "apache",
    "services": [
        {
            "id": "frontend",
            "apps": [
                {
                    "id": "apache",
                    "domain": {
                        "uri": "sloppy-cli-testing.sloppy.zone"
                    },
                    "mem": "124",
                    "image": "sloppy/apache-php",
                    "instances": 1,
                    "port_mappings": [
                        {
                            "container_port": 80
                        },
                        {
                            "container_port": 443
                        }
                    ]
                }
            ]
        }
    ]
}`)

	d := newDecoder(reader, stringMap{})
	input := new(api.Project)

	if err := d.DecodeJSON(input); err == nil {
		t.Error("Expected json type mismatch error")
	} else if err.Error() != "got type mismatch on line 12:32, expect number" {
		t.Errorf("Error = '%s', want 'got type mismatch on line 12:32, expect number'", err.Error())
	}
}

func TestDecode_unknownFieldsJSON(t *testing.T) {
	var unknownFieldsTests = []struct {
		input       string
		expectError bool
	}{
		0: {
			input:       `{"foo":"bar"}`,
			expectError: true,
		},
		1: {
			input:       `{"project":"bar"}`,
			expectError: false,
		},
		2: {
			input:       `{"project":"bar","services":[{"foo":"test"}]}`,
			expectError: true,
		},
		3: {
			input:       `{"project":"bar","services":[{"id":"test"}]}`,
			expectError: false,
		},
		4: {
			input:       `{"project":"bar","services":[{"apps":[{"foo":"test"}]}]}`,
			expectError: true,
		},
		5: {
			input:       `{"project":"bar","services":[{"apps":[{"id":"test"}]}]}`,
			expectError: false,
		},
		6: {
			input:       `{"project":"bar","services":[{"apps":[{"id":"test", "port_mappings":[{"foo": 8080}]}]}]}`,
			expectError: true,
		},
		7: {
			input:       `{"project":"bar","services":[{"apps":[{"id":"test", "port_mapping":[{"container_port": 8080}]}]}]}`,
			expectError: true,
		},
		8: {
			input:       `{"project":"bar","services":[{"apps":[{"id":"test", "port_mappings":[{"container_port": 8080}]}]}]}`,
			expectError: false,
		},
	}

	for i, tt := range unknownFieldsTests {
		r := strings.NewReader(tt.input)
		d := newDecoder(r, stringMap{})
		input := new(api.Project)
		err := d.DecodeJSON(input)
		if err != nil && !tt.expectError {
			t.Errorf("%d) Unexpected json unknown key error: %v", i, err)
		}
		if err == nil && tt.expectError {
			t.Errorf("%d) Expected json unknown key error", i)
		}
	}
}

func TestDecodeYAML(t *testing.T) {
	reader := strings.NewReader(testYAMLInput)
	d := newDecoder(reader, stringMap{})
	input := new(api.Project)

	if err := d.DecodeYAML(input); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	want := &api.Project{
		Name: api.String("wordpress"),
		Services: []*api.Service{
			{
				ID: api.String("frontend"),
				Apps: []*api.App{
					{
						ID:    api.String("apache"),
						Image: api.String("wordpress:4.2"),
						SSL:   api.Bool(true),
						Domain: &api.Domain{
							URI: api.String("superblog.volks.cloud"),
						},
						Memory: api.Int(512),
						EnvVars: map[string]string{
							"WORDPRESS_DB_HOST":     "mysql.backend.wordpress",
							"WORDPRESS_DB_USER":     "wordpress",
							"WORDPRESS_DB_PASSWORD": "wordpress",
						},
						Instances: api.Int(1),
						PortMappings: []*api.PortMap{
							{
								Port: api.Int(80),
							},
						},
						Dependencies: []string{
							"../../backend/mysql",
						},
					},
				},
			},
			{
				ID: api.String("backend"),
				Apps: []*api.App{
					{
						ID:     api.String("mysql"),
						Image:  api.String("mysql"),
						Memory: api.Int(512),
						EnvVars: map[string]string{
							"MYSQL_ROOT_PASSWORD": "supersicher",
							"MYSQL_USER":          "wordpress",
							"MYSQL_PASSWORD":      "wordpress",
							"MYSQL_DATABASE":      "wordpress",
						},
						Instances: api.Int(1),
						Command:   api.String("mysqld"),
						PortMappings: []*api.PortMap{
							{
								Port: api.Int(3306),
							},
						},
						Volumes: []*api.Volume{
							{
								Path: api.String("/var/lib/mysql"),
								Size: api.String("8GB"),
							},
						},
						HealthChecks: []*api.HealthCheck{
							{
								Timeout:              api.Int(10),
								Interval:             api.Int(10),
								MaxConsectiveFailure: api.Int(3),
								Path:                 api.String("/"),
								Type:                 api.String("HTTP"),
								GracePeriod:          api.Int(3),
							},
						},
						Logging: &api.Logging{
							Driver: api.String("syslog"),
							Options: map[string]string{
								"syslog-address": "tcp://192.168.0.42:123",
							},
						},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(input, want) {
		t.Errorf("\nApp:\t%+v\nWant:\t%+v\n", input, want)
	}
}

func TestDecodeYAML_minimalYAML(t *testing.T) {
	reader := strings.NewReader(testMinimalYAMLInput)
	d := newDecoder(reader, stringMap{})
	input := new(api.Project)

	if err := d.DecodeYAML(input); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	want := &api.Project{
		Name: api.String("wordpress"),
		Services: []*api.Service{
			{
				ID: api.String("frontend"),
				Apps: []*api.App{
					{
						ID:    api.String("apache"),
						Image: api.String("wordpress:4.2"),
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(input, want) {
		t.Errorf("App = %+v, want %+v", input, want)
	}
}

func TestDecodeYAML_errors(t *testing.T) {
	var errorsTests = []struct {
		in  string
		out string
	}{
		{
			in:  "unknown: 123",
			out: "'unknown' key is not supported",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tunknown: 1",
			out: "'app1.unknown' key is not supported",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\thealthchecks:\n\t\t\t\t- path1: 1",
			out: "'app1.healthchecks[1].path1' key is not supported",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tvolumes:\n\t\t\t\t- path1: 1",
			out: "'app1.volumes[1].path1' key is not supported",
		},

		{
			in:  "project: test\n",
			out: "invalid version specified",
		},
		{
			in:  "version: \"v5\"\n",
			out: "invalid version specified",
		},
		{
			in:  "version: 2\n",
			out: "invalid version specified",
		},

		{
			in:  "version: v1\nservices:\n\ttest: a\n",
			out: "'services' expects a 'service.id'",
		},
		{
			in:  "version: v1\nservices:\n\ttrue:",
			out: "'service.id' needs to be a string",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\ttest: test\n",
			out: "'service1' expects an 'app.id'",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\t1:\n",
			out: "'app.id' needs to be a string",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1: test\n",
			out: "'service1' expects an 'app.id'",
		},

		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\t1: 1",
			out: "'app1' keys need to be strings",
		},

		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\timage: 1",
			out: "'app1.image' needs to be a string",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tdomain: 1",
			out: "'app1.domain' needs to be a string",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tcmd: 1",
			out: "'app1.cmd' needs to be a string",
		},

		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tinstances: a",
			out: "'app1.instances' needs to be an integer",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tmem: a",
			out: "'app1.mem' needs to be an integer",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tport: a",
			out: "'app1.port' needs to be an integer",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tdependencies: a",
			out: "'app1.dependencies' needs to be an array",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tdependencies:\n\t\t\t\t- 1",
			out: "'app1.dependencies[1]' needs to be a string",
		},

		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tvolumes: a",
			out: "'app1.volumes' needs to be an array",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tvolumes:\n\t\t\t\t- a",
			out: "'app1.volumes[1]' expects a volume object",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tvolumes:\n\t\t\t\t- 1: 1",
			out: "'app1.volumes[1]' keys need to be strings",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tvolumes:\n\t\t\t\t- path: 1",
			out: "'app1.volumes[1].path' needs to be a string",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tvolumes:\n\t\t\t\t- size: 1",
			out: "'app1.volumes[1].size' needs to be a string",
		},

		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\thealthchecks: a",
			out: "'app1.healthchecks' needs to be an array",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\thealthchecks:\n\t\t\t\t- a",
			out: "'app1.healthchecks[1]' expects a healthcheck object",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\thealthchecks:\n\t\t\t\t- 1: 1",
			out: "'app1.healthchecks[1]' keys need to be strings",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\thealthchecks:\n\t\t\t\t- path: 1",
			out: "'app1.healthchecks[1].path' needs to be a string",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\thealthchecks:\n\t\t\t\t- timeout: a",
			out: "'app1.healthchecks[1].timeout' needs to be an integer",
		},

		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tenv: a",
			out: "'app1.env' needs to be either an object or an array",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tenv:\n\t\t\t\t- 1: 1",
			out: "'app1.env[1]' key need to be a string",
		},
		{
			in:  "version: v1\nservices:\n\tservice1:\n\t\tapp1:\n\t\t\tenv:\n\t\t\t\ttest: 1",
			out: "'app1.env[1].test' value need to be a string",
		},
	}

	for i, tt := range errorsTests {
		reader := strings.NewReader(strings.Replace(tt.in, "\t", "  ", -1))
		d := newDecoder(reader, stringMap{})
		input := new(api.Project)

		err := d.DecodeYAML(input)
		if err == nil {
			t.Fatalf("%d) Expected error to be returned", i)
		}

		if err.Error() != tt.out {
			t.Log(tt.in)
			t.Errorf("%d) Error = %v, want %s", i, err, tt.out)
		}
	}
}

func TestSet(t *testing.T) {
	type testType struct {
		Test1 *string
		Test2 *int
		Test3 *bool
		Test4 *float64
		Test5 *string `json:"test6"`
	}

	var out testType
	want := &testType{
		Test1: api.String("abcd"),
		Test2: api.Int(42),
		Test3: api.Bool(true),
		Test4: api.Float64(5.42),
		Test5: api.String("test6"),
	}

	m := yaml.MapSlice{
		{
			Key:   "test1",
			Value: "abcd",
		},
		{
			Key:   "test2",
			Value: 42,
		},
		{
			Key:   "test3",
			Value: true,
		},
		{
			Key:   "test4",
			Value: 5.42,
		},
		{
			Key:   "test6",
			Value: "test6",
		},
	}

	if err := set(m, &out); err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(out, want) {
		t.Errorf("set(%v) = %v, want %v", m, out, want)
	}
}

func TestReplaceReader(t *testing.T) {
	r := strings.NewReader(`$abc $abc123`)
	pattern := map[string]string{
		"abc":    "1",
		"abc123": "2",
	}
	rr, err := replaceReader(r, pattern)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	b, _ := ioutil.ReadAll(rr)
	if string(b) != `1 2` {
		t.Errorf("replaceReader(%v) = %s, want 1 2", pattern, string(b))
	}
}

func TestReplaceVariables_missingVariable(t *testing.T) {
	r := strings.NewReader(`$abc $abc123 \\$escaped`)
	pattern := map[string]string{
		"abc": "1",
	}
	_, err := replaceReader(r, pattern)
	if err == nil {
		t.Errorf("Expect error to be returned: pattern: %v input: %s", pattern, "$abc $abc123 \\$escaped")
	}
}

func TestReplaceVariables_escapedVariable(t *testing.T) {
	r := strings.NewReader(`$abc \\$abc123`)
	pattern := map[string]string{
		"abc": "1",
	}
	rr, err := replaceReader(r, pattern)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	b, _ := ioutil.ReadAll(rr)
	if string(b) != `1 $abc123` {
		t.Errorf("replaceReader(%v) = %s, want 1 $abc123", pattern, string(b))
	}
}

func TestFindUnknownFields(t *testing.T) {
	var unknownFieldTests = []struct {
		fieldMap    map[string]interface{}
		structType  reflect.Type
		expectError bool
	}{
		0: {
			fieldMap: map[string]interface{}{
				"Foo": 1,
			},
			structType:  reflect.TypeOf(struct{ Foo int }{}),
			expectError: false,
		},
		1: {
			fieldMap: map[string]interface{}{
				"Bar": 1,
			},
			structType: reflect.TypeOf(struct {
				Foo int `json:"Bar"`
			}{}),
			expectError: false,
		},
		2: {
			fieldMap: map[string]interface{}{
				"Bar": 1,
			},
			structType:  reflect.TypeOf(struct{ Foo int }{}),
			expectError: true,
		},
		3: {
			fieldMap: map[string]interface{}{
				"Foo": []interface{}{"Bar", "Bar"},
			},
			structType:  reflect.TypeOf(struct{ Foo []string }{}),
			expectError: false,
		},
		4: {
			fieldMap: map[string]interface{}{
				"Foo": []interface{}{
					map[string]interface{}{
						"Bar": "test",
					},
					map[string]interface{}{
						"Bar": "test1",
					},
				},
			},
			structType: reflect.TypeOf(struct {
				Foo []struct {
					Bar string
				}
			}{}),
			expectError: false,
		},
		5: {
			fieldMap: map[string]interface{}{
				"Foo": []interface{}{
					map[string]interface{}{
						"Bar": "test",
					},
					map[string]interface{}{
						"Foo": "test1",
					},
				},
			},
			structType: reflect.TypeOf(struct {
				Foo []struct {
					Bar string
				}
			}{}),
			expectError: true,
		},
		6: {
			fieldMap: map[string]interface{}{
				"Foo": map[string]interface{}{
					"Bar": "test",
				},
			},
			structType: reflect.TypeOf(struct {
				Foo struct {
					Bar string
				}
			}{}),
			expectError: false,
		},
		7: {
			fieldMap: map[string]interface{}{
				"Foo": map[string]interface{}{
					"Foo": "test1",
				},
			},
			structType: reflect.TypeOf(struct {
				Foo struct {
					Bar string
				}
			}{}),
			expectError: true,
		},
	}

	for i, tt := range unknownFieldTests {
		err := findUnknownFields(tt.fieldMap, tt.structType)
		if err != nil && !tt.expectError {
			t.Errorf("%d) Unexpected error = %v; fieldMap: %v struct type: %v", i, err, tt.fieldMap, tt.structType)
		}
		if err == nil && tt.expectError {
			t.Errorf("%d) Expect error to be returned: fieldMap: %v struct type: %v", i, tt.fieldMap, tt.structType)
		}
	}
}

var testYAMLInput = `version: "v1"
project: "wordpress"
services:
  frontend:
    apache:
      image: "wordpress:4.2"
      ssl: true
      instances: 1
      mem: 512
      domain: "superblog.volks.cloud"
      port: 80
      env:
        - WORDPRESS_DB_HOST: "mysql.backend.wordpress"
        - WORDPRESS_DB_USER: "wordpress"
        - WORDPRESS_DB_PASSWORD: "wordpress"
      dependencies:
        - "../../backend/mysql"
  backend:
    mysql:
      image: "mysql"
      instances: 1
      mem: 512
      cmd: "mysqld"
      port: 3306
      volumes:
        -
          path: "/var/lib/mysql"
          size: "8GB"
      env:
        MYSQL_ROOT_PASSWORD: "supersicher"
        MYSQL_USER: "wordpress"
        MYSQL_PASSWORD: "wordpress"
        MYSQL_DATABASE: "wordpress"
      healthchecks:
        -
          path: "/"
          type: "HTTP"
          timeout_seconds: 10
          interval_seconds: 10
          max_consecutive_failures: 3
          grace_period_seconds: 3
      logging:
        driver: syslog
        options:
          syslog-address: "tcp://192.168.0.42:123"
`

var testMinimalYAMLInput = `version: "v1"
project: "wordpress"
services:
  frontend:
    apache:
      image: "wordpress:4.2"
`
