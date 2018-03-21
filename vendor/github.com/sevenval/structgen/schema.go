package structgen

import (
	"encoding/json"
	"path"
	"unicode"
	"unicode/utf8"
)

const (
	TypeArray  = "array"
	TypeBool   = "boolean"
	TypeInt    = "integer"
	TypeNull   = "null"
	TypeNumber = "number"
	TypeObject = "object"
	TypeString = "string"
)

type Schema struct {
	AdditionalProperties interface{} `json:"additionalProperties,omitempty"` // bool or []*Schema
	Definitions          SchemaMap   `json:"definitions,omitempty"`
	Description          string      `json:"description,omitempty"`
	Format               string      `json:"format,omitempty"`
	ID                   string      `json:"id,omitempty"` // $id in draft 7
	Items                *Schema     `json:"items,omitempty"`
	OneOf                []*Schema   `json:"oneOf,omitempty"`
	AnyOf                []*Schema   `json:"anyOf,omitempty"`
	AllOf                []*Schema   `json:"allOf,omitempty"`
	PatternProperties    SchemaMap   `json:"patternProperties,omitempty"`
	Properties           SchemaMap   `json:"properties,omitempty"`
	Reference            string      `json:"$ref,omitempty"`
	Required             []string    `json:"required,omitempty"`
	Schema               string      `json:"$schema,omitempty"`
	Title                string      `json:"title,omitempty"`
	Type                 interface{} `json:"type,omitempty"` // string or array
	UniqueItems          bool        `json:"uniqueItems,omitempty"`
}

type SchemaMap map[string]*Schema

func NewSchema(from []byte) (*Schema, error) {
	s := &Schema{}
	err := json.Unmarshal(from, s)
	return s, err
}

func (s *Schema) toCamelCase(str string) string {
	var result []byte
	toUpper := true
	for i, width := 0, 0; i < len(str); i += width {
		r, w := utf8.DecodeRuneInString(str[i:])
		width = w
		if toUpper {
			r = unicode.ToUpper(r)
			toUpper = false
		}
		if r == '_' || r == '-' {
			toUpper = true
		} else {
			result = append(result, byte(r))
		}
	}
	return string(result)
}

func (s *Schema) GetType() (t string, tlist []string) {
	switch s.Type.(type) {
	case string:
		t = s.Type.(string)
		tlist = append(tlist, t)
	case []interface{}: // ["number", "string"]
		a := s.Type.([]interface{})
		for _, str := range a {
			tlist = append(tlist, str.(string))
		}
		t = tlist[0]
	}
	return t, tlist
}

// name have to be given by parent. schema doesn't know own prop name
func (s *Schema) getGoType(name string) string {
	t, tlist := s.GetType()
	if len(tlist) > 1 && ((tlist[0] == TypeObject && tlist[1] == TypeNull) ||
		(tlist[1] == TypeObject && tlist[0] == TypeNull)) {
		t = TypeObject
	} else if len(tlist) > 1 {
		t = "interface{}"
	}

	// handle complex types
	if t == "" && (len(s.OneOf) > 0 || len(s.AnyOf) > 0 ||
		len(s.AllOf) > 0 || s.Reference != "") {
		t = "interface{}"
	}

	// TODO resolve later on?
	if name == "" {
		name = "#"
	}

	// assume a map of objects if object type and patternProbs given
	if t == TypeObject && len(s.PatternProperties) > 0 {
		for _, schema := range s.PatternProperties {
			if schema.Reference != "" {
				_, n := path.Split(schema.Reference)
				t = "map[string]*" + s.toCamelCase(n)
			} else {
				t = schema.getGoType(name)
			}
			break
		}
	}

	switch t {
	case TypeArray:
		t = "[]" + s.Items.getGoType(name)
	case TypeNull:
		fallthrough
	case TypeObject:
		t = "*" + s.toCamelCase(name)
	case TypeInt:
		t = "int"
	case TypeBool:
		t = "bool"
	case TypeNumber:
		t = "float64"
	}
	return t
}
