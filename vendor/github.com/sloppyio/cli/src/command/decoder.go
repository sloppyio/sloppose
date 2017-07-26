package command

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/sloppyio/cli/src/api"
)

type decoder struct {
	reader      io.Reader
	err         error
	errorReader io.Reader
}

func newDecoder(r io.Reader, vars stringMap) *decoder {
	var buf bytes.Buffer
	r, err := replaceReader(r, vars)
	return &decoder{
		reader:      io.TeeReader(r, &buf),
		err:         err,
		errorReader: &buf,
	}
}

func (d *decoder) DecodeJSON(p *api.Project) error {
	if d.err != nil {
		return d.err
	}

	// Detect unknown json keys
	var aux json.RawMessage
	var keys map[string]interface{}
	err := json.NewDecoder(d.reader).Decode(&aux)
	if err == nil {
		err = json.Unmarshal(aux, p)
		if err := json.Unmarshal(aux, &keys); err == nil {
			if err := findUnknownFields(keys, reflect.TypeOf(*p)); err != nil {
				return err
			}
		}
	}

	offset := 0
	message := ""
	switch err := err.(type) {
	case *json.SyntaxError:
		message = "got syntax error around line %d:%d"
		offset = int(err.Offset)
	case *json.UnmarshalTypeError:
		want := "object"
		switch err.Type.Kind() {
		case reflect.Slice, reflect.Array:
			want = "array"
		case reflect.Int, reflect.Float32, reflect.Float64:
			want = "number"
		case reflect.String:
			want = "string"
		}
		message = "got type mismatch on line %d:%d, expect " + want
		offset = int(err.Offset)
	default:
		return err
	}

	line := 1
	n := 0 // Read bytes
	scanner := bufio.NewScanner(d.errorReader)
	for scanner.Scan() {
		next := len(scanner.Bytes()) + 1
		if n+next > int(offset) {
			break
		}
		n += next
		line++
	}
	column := offset - n
	return fmt.Errorf(message, line, column)
}

func (d *decoder) DecodeYAML(p *api.Project) error {
	if d.err != nil {
		return d.err
	}

	data, err := ioutil.ReadAll(d.reader)
	if err != nil {
		return err
	}

	aux := &project{p}

	return yaml.Unmarshal(data, aux)
}

type project struct {
	*api.Project
}

func (p *project) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux yaml.MapSlice

	if err := unmarshal(&aux); err != nil {
		return err
	}

	var version string

	for _, root := range aux {
		key := root.Key.(string)

		switch key {
		case "version":
			v, ok := root.Value.(string)
			if !ok || v != "v1" {
				return fmt.Errorf("invalid version specified")
			}
			version = v
		case "project":
			if name, ok := root.Value.(string); ok {
				p.Name = api.String(name)
			}
		case "services":
			services, ok := root.Value.(yaml.MapSlice)
			if !ok {
				return &yamlError{key: "services", message: "expects a 'service.id'"}
			}
			p.Services = make([]*api.Service, 0, cap(services))

			for _, firstLevel := range services {
				id, ok := firstLevel.Key.(string)
				if !ok {
					return &yamlError{namespace: "service", key: "id", message: "needs to be a string"}
				}

				service := new(api.Service)
				service.ID = api.String(id)

				// App
				apps, ok := firstLevel.Value.(yaml.MapSlice)
				if !ok {
					return &yamlError{key: "services", message: "expects a 'service.id'"}
				}

				for _, secondLevel := range apps {
					id, ok := secondLevel.Key.(string)
					if !ok {
						return &yamlError{namespace: "app", key: "id", message: "needs to be a string"}
					}
					app := new(api.App)
					app.ID = api.String(id)

					settings, ok := secondLevel.Value.(yaml.MapSlice)
					if !ok {
						return &yamlError{key: *service.ID, message: "expects an 'app.id'"}
					}

					for _, thirdLevel := range settings {
						parameter, ok := thirdLevel.Key.(string)
						if !ok {
							return &yamlError{key: id, message: "keys need to be strings"}
						}

						switch parameter {
						case "image":
							if image, ok := thirdLevel.Value.(string); ok {
								app.Image = api.String(image)
							} else {
								return &yamlError{namespace: id, key: parameter, message: "needs to be a string"}
							}
						case "domain":
							if domain, ok := thirdLevel.Value.(string); ok {
								app.Domain = &api.Domain{URI: api.String(domain)}
							} else {
								return &yamlError{namespace: id, key: parameter, message: "needs to be a string"}
							}
						case "ssl":
							if ssl, ok := thirdLevel.Value.(bool); ok {
								app.SSL = api.Bool(ssl)
							} else {
								return &yamlError{namespace: id, key: parameter, message: "needs to be boolean"}
							}
						case "cmd":
							if cmd, ok := thirdLevel.Value.(string); ok {
								app.Command = api.String(cmd)
							} else {
								return &yamlError{namespace: id, key: parameter, message: "needs to be a string"}
							}
						case "instances":
							if instances, ok := thirdLevel.Value.(int); ok {
								app.Instances = api.Int(instances)
							} else {
								return &yamlError{namespace: id, key: parameter, message: "needs to be an integer"}
							}
						case "mem":
							if memory, ok := thirdLevel.Value.(int); ok {
								app.Memory = api.Int(memory)
							} else {
								return &yamlError{namespace: id, key: parameter, message: "needs to be an integer"}
							}
						case "ports":
							fallthrough
						case "port":
							if port, ok := thirdLevel.Value.(int); ok {
								app.PortMappings = []*api.PortMap{{Port: api.Int(port)}}
							} else {
								return &yamlError{namespace: id, key: parameter, message: "needs to be an integer"}
							}
						case "depends_on":
							fallthrough
						case "dependencies":
							if dependencies, ok := thirdLevel.Value.([]interface{}); ok {
								for i, dependency := range dependencies {
									dep, ok := dependency.(string)
									if !ok {
										return &yamlError{namespace: id, key: parameter, index: i + 1, message: "needs to be a string"}
									}
									app.Dependencies = append(app.Dependencies, dep)
								}
							} else {
								return &yamlError{namespace: id, key: parameter, message: "needs to be an array"}
							}
						case "volumes":
							if volumes, ok := thirdLevel.Value.([]interface{}); ok {
								for i, v := range volumes {
									volume, ok := v.(yaml.MapSlice)
									if !ok {
										return &yamlError{namespace: id, key: parameter, index: i + 1, message: "expects a volume object"}
									}
									vol := new(api.Volume)
									if err := set(volume, vol); err != nil {
										return &yamlError{namespace: id, key: parameter, index: i + 1, subkey: err.key, message: err.message}
									}
									app.Volumes = append(app.Volumes, vol)
								}
							} else {
								return &yamlError{namespace: id, key: parameter, message: "needs to be an array"}
							}
						case "healthchecks":
							if healthChecks, ok := thirdLevel.Value.([]interface{}); ok {
								for i, healthCheck := range healthChecks {
									hc, ok := healthCheck.(yaml.MapSlice)
									if !ok {
										return &yamlError{namespace: id, key: parameter, index: i + 1, message: "expects a healthcheck object"}
									}
									h := new(api.HealthCheck)
									if err := set(hc, h); err != nil {
										return &yamlError{namespace: id, key: parameter, index: i + 1, subkey: err.key, message: err.message}
									}
									app.HealthChecks = append(app.HealthChecks, h)
								}
							} else {
								return &yamlError{namespace: id, key: parameter, message: "needs to be an array"}
							}
						case "env":
							addEnv := func(m yaml.MapSlice) error {
								for i, env := range m {
									e, ok := env.Key.(string)
									if !ok {
										return &yamlError{namespace: id, key: parameter, index: i + 1, message: "key need to be a string"}
									}
									v, ok := env.Value.(string)
									if !ok {
										return &yamlError{namespace: id, key: parameter, index: i + 1, subkey: e, message: "value need to be a string"}
									}
									app.EnvVars[e] = v
								}
								return nil
							}

							switch envs := thirdLevel.Value.(type) {
							case []interface{}:
								app.EnvVars = make(map[string]string, len(envs))
								for _, v := range envs {
									if envs, ok := v.(yaml.MapSlice); ok {
										if err := addEnv(envs); err != nil {
											return err
										}
									}
								}
							case yaml.MapSlice:
								app.EnvVars = make(map[string]string, len(envs))
								if err := addEnv(envs); err != nil {
									return err
								}
							default:
								return &yamlError{namespace: id, key: parameter, message: "needs to be either an object or an array"}
							}
						case "logging":
							const keyDriver = "driver"
							m, ok := thirdLevel.Value.(yaml.MapSlice)
							if !ok {
								return &yamlError{namespace: id, key: parameter, message: "expected to be a logging object"}
							}

							logging := &api.Logging{
								Options: make(map[string]string),
							}

							for i, mapItem := range m {
								key, ok := mapItem.Key.(string)
								if !ok {
									return &yamlError{namespace: id, key: parameter, index: i, message: "key needs to be a string"}
								}
								if key == keyDriver {
									value, ok := mapItem.Value.(string)
									if !ok {
										return &yamlError{namespace: id, key: parameter, index: i, message: "value needs to be a string"}
									}
									logging.Driver = api.String(value)
								} else {
									mapSlice, ok := mapItem.Value.(yaml.MapSlice)
									if !ok {
										return &yamlError{namespace: id, key: parameter, index: i, message: "expected to be a object"}
									}
									for idx, item := range mapSlice {
										k, ok := item.Key.(string)
										if !ok {
											return &yamlError{namespace: id, key: parameter, index: idx, message: "key needs to be a string"}
										}
										v, ok := item.Value.(string)
										if !ok {
											return &yamlError{namespace: id, key: parameter, index: idx, message: "value needs to be a string"}
										}
										logging.Options[k] = v
									}
								}
							}
							app.Logging = logging
						default:
							return &yamlError{namespace: id, key: parameter, message: "key is not supported"}
						}
					}
					service.Apps = append(service.Apps, app)
				}
				p.Services = append(p.Services, service)
			}
		default:
			return &yamlError{key: key, message: "key is not supported"}
		}
	}

	if version == "" {
		return fmt.Errorf("invalid version specified")
	}

	return nil
}

// Set iterates over the given yaml.MapSlice and stores the result in the value pointed to by v.
//
// To match incoming key:value-pair to a struct it prefers an exact match for either the struct
// field name or its json tag but also accepts a case-insensitive match.
// It only sets exported fields ot the struct.
//
// If a key cannot be matched, it will return a yamlError. If a key:value-pair
// (string:interface{}) is not appropriate for a given target type, it will return
// a yamlError as well.
//
// set is implemented in order to reduce code duplication of particular keys/structs,
// e.g. Volumes, HealthChecks.
//
// If there is a yaml decoder which is implemented as the stdlib's json decoder, it will probably be obsolete.
func set(m yaml.MapSlice, v interface{}) *yamlError {
	matchedKey := make(map[string]bool, len(m))
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()

	for _, p := range m {
		key, ok := p.Key.(string)
		if !ok {
			return &yamlError{message: "keys need to be strings"}
		}
		matchedKey[key] = false

		for i := 0; i < rt.NumField(); i++ {
			fieldType := rt.Field(i)
			fieldValue := rv.Field(i)

			// Skip unexported fields of struct.
			if fieldType.PkgPath != "" && !fieldType.Anonymous {
				continue
			}
			tag := fieldType.Tag.Get("json")
			if tag == "-" {
				continue
			}
			if tag == "" {
				tag = fieldType.Name
			}
			if j := strings.Index(tag, ","); j != -1 {
				tag = tag[:j]
			}
			if tag != key && !strings.EqualFold(fieldType.Name, key) {
				continue
			}
			matchedKey[key] = true

			var kind reflect.Kind
			if fieldType.Type.Kind() == reflect.Ptr {
				kind = fieldType.Type.Elem().Kind()
			} else {
				return &yamlError{key: fieldType.Name, message: "needs to be a pointer"}
			}

			switch kind {
			case reflect.String:
				v, ok := p.Value.(string)
				if !ok {
					return &yamlError{message: "needs to be a string", key: key}
				}
				fieldValue.Set(reflect.ValueOf(api.String(v)))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				v, ok := p.Value.(int)
				if !ok {
					return &yamlError{message: "needs to be an integer", key: key}
				}
				fieldValue.Set(reflect.ValueOf(api.Int(v)))
			case reflect.Float64:
				v, ok := p.Value.(float64)
				if !ok {
					return &yamlError{message: "needs to be a float", key: key}
				}
				fieldValue.Set(reflect.ValueOf(api.Float64(v)))
			case reflect.Bool:
				v, ok := p.Value.(bool)
				if !ok {
					return &yamlError{message: "needs to be a boolean", key: key}
				}
				fieldValue.Set(reflect.ValueOf(api.Bool(v)))
			default:
				return &yamlError{message: "does not support type", key: key}
			}
		}
	}

	// Check for unknown keys which could not be matched
	for key, matched := range matchedKey {
		if !matched {
			return &yamlError{key: key, message: "key is not supported"}
		}
	}

	return nil
}

type yamlError struct {
	namespace string
	key       string
	subkey    string
	index     int
	message   string
}

func (e *yamlError) Error() string {
	var attribute string
	if e.key != "" {
		attribute = e.key
	}
	if e.namespace != "" {
		attribute = fmt.Sprintf("%s.%s", e.namespace, attribute)
	}
	if e.index != 0 {
		attribute = fmt.Sprintf("%s[%d]", attribute, e.index)
	}
	if e.subkey != "" {
		attribute = fmt.Sprintf("%s.%s", attribute, e.subkey)
	}
	if attribute != "" {
		return fmt.Sprintf("'%s' %s", attribute, e.message)
	}
	return fmt.Sprintf("%s%s", attribute, e.message)
}

var validVar = regexp.MustCompile(`(?:\\)*\$[[:alnum:]_-]+`)

// replaceReader returns an io.Reader, replacing matches of all variables with
// the replacement pattern, or returns error if any.
func replaceReader(r io.Reader, pattern map[string]string) (io.Reader, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	data = validVar.ReplaceAllFunc(data, func(b []byte) []byte {
		i := bytes.Index(b, []byte{'$'})
		// Ignore escaped variables
		if c := bytes.Count(b[:i], []byte{'\\'}); c > 0 && c%2 == 0 {
			return b[i:]
		}
		if v, ok := pattern[string(b[i+1:])]; ok {
			return []byte(v)
		}

		err = fmt.Errorf("missing variable '%s'. ", string(b[i+1:]))
		return b
	})

	return bytes.NewReader(data), err
}

// findUnknownFields returns an error if fields map does not match the given type.
// As soon as https://github.com/golang/go/issues/15314 gets accepted, we can remove
// the following function.
func findUnknownFields(fields map[string]interface{}, t reflect.Type) error {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// Skip unexported fields of struct.
		if f.PkgPath != "" && !f.Anonymous {
			continue
		}

		tag := f.Tag.Get("json")
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = f.Name
		}
		if j := strings.Index(tag, ","); j != -1 {
			tag = tag[:j]
		}

		ft := f.Type
		if f.Type.Kind() == reflect.Ptr {
			ft = f.Type.Elem()
		}

		switch ft.Kind() {
		case reflect.Struct:
			if v, ok := fields[tag].(map[string]interface{}); ok {
				if err := findUnknownFields(v, ft); err != nil {
					return err
				}
			}
		case reflect.Slice:
			if ft.Elem().Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Elem().Kind() != reflect.Struct {
				delete(fields, tag)
				continue
			}
			slice, ok := fields[tag].([]interface{})
			if !ok {
				delete(fields, tag)
				continue
			}

			for _, elem := range slice {
				v, ok := elem.(map[string]interface{})
				if !ok {
					break
				}
				if err := findUnknownFields(v, ft.Elem()); err != nil {
					return err
				}
			}
		}
		delete(fields, tag)
	}

	for key := range fields {
		return fmt.Errorf("json: key '%s' not supported", key)
	}

	return nil
}
