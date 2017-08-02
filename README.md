# Sloppose

[![Build Status](https://travis-ci.org/sloppyio/sloppose.svg?branch=master)](https://travis-ci.org/sloppyio/sloppose) [![Coverage Status](https://coveralls.io/repos/github/sloppyio/sloppose/badge.svg?branch=feature%2Fsloppy-yml-test)](https://coveralls.io/github/sloppyio/sloppose?branch=feature%2Fsloppy-yml-test)

Small tool to convert docker-compose files to sloppy.io compatible ones.

Supports docker-compose versions **2** and **3**.

## Usage

`sloppose [command]`

**Commands**:
* `convert [options] [files]`
    * Example: `sloppose convert -o outFile.yml -projectname example`

## Configuration

**Projectname**:
* can be set with `COMPOSE_PROJECT_NAME` environment variable or with parameter as seen above.
* defaults to current working dir

## Development

Checkout to `$GOPATH/src/github.com/sloppyio/sloppose`

Create a development build within the Go environment: `make build-dev`

Create an osx build with docker: `docker run -v $PWD:/go/src/github.com/sloppyio/sloppose --workdir /go/src/github.com/sloppyio/sloppose -e GOOS=darwin golang:1.8.3 make build-dev`
