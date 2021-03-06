# Sloppose

[![Build Status](https://travis-ci.org/sloppyio/sloppose.svg?branch=master)](https://travis-ci.org/sloppyio/sloppose) [![Coverage Status](https://coveralls.io/repos/github/sloppyio/sloppose/badge.svg?branch=master)](https://coveralls.io/github/sloppyio/sloppose?branch=master)

Library to convert docker-compose files to sloppy.io compatible ones. Integrated into the sloppy.io CLI ([learn more](https://kb.sloppy.io/features/cli-command-reference/12-start)). Can also be used as a standalone tool for one-time conversions.

Supports docker-compose version **3**.

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

Create an osx build with docker: `make dev-osx`

To run tests with Docker: `test-in-docker`
