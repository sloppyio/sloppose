package converter_test

import (
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/sloppyio/sloppose/pkg/converter"
)

var expectedSloppyYml = `project: sloppy-test
services:
  apps:
    busy_env:
      cmd: sleep 20
      dependencies:
      - ../apps/wordpress
      env:
      - VAR_A=1
      - VAR_B=test
      image: busybox
    db:
      cmd: mysqld
      env:
      - MYSQL_DATABASE=wordpress
      - MYSQL_PASSWORD=wordpress
      - MYSQL_ROOT_PASSWORD=somewordpress
      - MYSQL_USER=wordpress
      image: mysql:8.0.0
      volumes:
      - container_path: /var/lib/mysql
    wordpress:
      dependencies:
      - ../apps/db
      domain: mywords.sloppy.zone
      env:
      - WORDPRESS_DB_HOST=db.apps.sloppy-test:3306
      - WORDPRESS_DB_PASSWORD=wordpress
      - WORDPRESS_DB_USER=wordpress
      image: wordpress:4.7.4
      port: 80
      volumes:
      - container_path: /var/www/html
version: v1
`

// output should be the same as described above
var testFiles = []string{
	"/testdata/docker-compose-v2.yml",
	"/testdata/docker-compose-v3-simple.yml",
}

func loadSloppyFile(filename string) (cf *converter.ComposeFile, sf *converter.SloppyFile) {
	reader := &converter.ComposeReader{}
	b, err := reader.Read(filename)
	if err != nil {
		panic(err)
	}
	cf, err = converter.NewComposeFile([][]byte{b}, "sloppy-test")
	if err != nil {
		panic(err)
	}

	sf, err = converter.NewSloppyFile(cf)
	if err != nil {
		panic(err)
	}
	linker := &converter.Linker{}
	err = linker.Resolve(cf, sf)
	if err != nil {
		panic(err)
	}
	return
}

func TestNewSloppyFile(t *testing.T) {
	wantLines := strings.Split(expectedSloppyYml, "\n")
	for _, testFile := range testFiles {
		_, have := loadSloppyFile(testFile)
		haveBuf, err := yaml.Marshal(have)
		if err != nil {
			t.Error(err)
		}

		haveLines := strings.Split(string(haveBuf), "\n")
		if diff := cmp.Diff(haveLines, wantLines); diff != "" {
			t.Errorf("Case: %q\nResult differs: (-got +want)\n%s", testFile, diff)
		}
	}
}
