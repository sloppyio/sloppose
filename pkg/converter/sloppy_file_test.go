package converter_test

import (
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/sloppyio/sloppose/pkg/converter"
)

var expectedSloppyYml = `project: pkg
services:
  apps:
    busy_env:
      cmd: sleep 20
      dependencies:
      - ../apps/db
      - ../apps/wordpress
      domain: $URI
      env:
      - VAR_A=1
      - VAR_B=test
      image: busybox
      instances: 1
      mem: 256
    db:
      cmd: mysqld
      domain: $URI
      env:
      - MYSQL_DATABASE=wordpress
      - MYSQL_PASSWORD=wordpress
      - MYSQL_ROOT_PASSWORD=somewordpress
      - MYSQL_USER=wordpress
      image: mysql:8.0.0"
      instances: 1
      mem: 256
      volumes:
      - container_path: /var/lib/mysql
        size: 8GB
    wordpress:
      dependencies:
      - ../apps/db
      domain: mywords.sloppy.zone
      env:
      - WORDPRESS_DB_HOST=db.apps.pkg:3306
      - WORDPRESS_DB_PASSWORD=wordpress
      - WORDPRESS_DB_USER=wordpress
      image: wordpress:4.7.4
      instances: 1
      mem: 256
      port: 80
      port_mappings:
      - container_port: 80
version: v1
`

func TestNewSloppyFile(t *testing.T) {
	reader := &converter.ComposeReader{}
	b, err := reader.Read("/testdata/docker-compose-v2.yml")
	if err != nil {
		t.Error(err)
	}
	cf, err := converter.NewComposeFile([][]byte{b}, "")
	if err != nil {
		t.Error(err)
	}

	have, err := converter.NewSloppyFile(cf)
	if err != nil {
		t.Error(err)
	}
	linker := &converter.Linker{}
	linker.Resolve(cf, have)

	haveBuf, err := yaml.Marshal(have)
	if err != nil {
		t.Error(err)
	}

	haveLines := strings.Split(string(haveBuf), "\n")
	wantLines := strings.Split(expectedSloppyYml, "\n")
	if diff := cmp.Diff(haveLines, wantLines); diff != "" {
		t.Errorf("Result differs: (-got +want)\n%s", diff)
	}
}
