package structgen

import (
	"bytes"
	"fmt"
	"text/template"
)

// Struct contains all relevant infos to generate a go struct string.
type Struct struct {
	Comment  string
	Fields   map[string]*Struct
	ID       string
	JSONName string
	Name     string
	Required bool
	Type     string
}

type StructMap map[string]*Struct

// String returns the self go-field-string representation.
func (s *Struct) String() string {
	n := s.JSONName
	if !s.Required {
		n += ",omitempty"
	}
	str := fmt.Sprintf("%s\t%s\t`json:%q`", s.Name, s.Type, n)
	if s.Comment != "" {
		str = fmt.Sprintf("%s // %s", str, s.Comment)
	}
	return str
}

// StructString returns the structs go-string representation.
func (s *Struct) StructString() (string, error) {
	if len(s.Fields) == 0 { // don't return structs
		return "", nil
	}

	const str = "type {{.Name}} struct {\n{{range .Fields}}\t{{.}}\n{{end}}}\n\n"
	t := template.New(s.Name)
	t, err := t.Parse(str)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = t.Execute(buf, s)
	return buf.String(), err
}
