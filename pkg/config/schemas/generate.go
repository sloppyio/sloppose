// +build ignore

package main

import (
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/sevenval/structgen"
)

func main() {
	cwd, err := os.Getwd()
	must(err)

	schemaPath := path.Join(cwd, "pkg/config/schemas")

	typeName := "DockerComposeV3"
	inFileName := "config_schema_v3.6.json"
	outFileName := "compose_v3.go"
	namespace := "config"

	buf, err := ioutil.ReadFile(path.Join(schemaPath, inFileName))
	must(err)
	schema, err := structgen.NewSchema(buf)
	must(err)
	generator := structgen.NewGenerator(typeName, namespace, schema)
	f, err := os.OpenFile(path.Join(schemaPath, "../", outFileName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	must(err)
	_, err = io.Copy(f, generator)
	must(err)
	must(f.Close())
}

func must(err error) {
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
