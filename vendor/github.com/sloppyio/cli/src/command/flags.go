package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
)

// stringMap is used to handle multiple flag
type stringMap map[string]string

func (s *stringMap) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringMap) Get() interface{} { return stringMap(*s) }

func (s *stringMap) Set(value string) error {
	if (*s) == nil {
		(*s) = make(stringMap)
	}

	// Backwards compatibility
	if strings.Contains(value, ",") {
		for _, keyValue := range strings.Split(value, ",") {
			s.Set(keyValue)
		}
		return nil
	}

	i := strings.Index(value, ":")
	if i != -1 && value[i+1:] != "" {
		(*s)[value[:i]] = value[i+1:]
	}

	return nil
}

func newFlagSet(name string, errorHandling flag.ErrorHandling) *flag.FlagSet {
	flag := flag.NewFlagSet(name, errorHandling)
	flag.SetOutput(ioutil.Discard)
	return flag
}
