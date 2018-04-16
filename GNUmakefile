.DEFAULT_GOAL := all

GOVERSION := 1.9.4
BUILDPATH := ./build
SRCPATH := ./cmd
APPNAME := sloppose
VERSION_NAMESPACE := github.com/sloppyio/sloppose/command

.PHONY: all generate test test-update build-dev osx linux win clean

define build
	GOOS=$(1) GOARCH=$(2) go build -ldflags "-X ${VERSION_NAMESPACE}.VersionName=`git describe --exact-match --abbrev=0` -X ${VERSION_NAMESPACE}.BuildName=`git log -1 --format=%h`" -o $(BUILDPATH)/$(APPNAME)_$(1)_$(2)$(3) $(SRCPATH)
endef

define zip
	cd build && zip $(1)_$(2).zip $(APPNAME)_$(1)_$(2)$(3)
endef

test:
	go test -v -race -timeout 30s -covermode=atomic -coverprofile=coverage.txt ./pkg/converter

test-update:
	@echo updating golden file...
	go test -v -timeout 30s ./pkg/converter -update

coverage-show:
	go tool cover -html=coverage.txt

coverage-stats:
	go tool cover -func=coverage.txt

coverage-report:
	goveralls -coverprofile=coverage.txt -service=travis-ci

dev-osx:
	docker run -v $(CURDIR):/go/src/github.com/sloppyio/sloppose --workdir /go/src/github.com/sloppyio/sloppose -e GOOS=darwin golang:$(GOVERSION) make build-dev

build-dev:
	go build -ldflags "-X ${VERSION_NAMESPACE}.VersionName=`git describe --exact-match --abbrev=0`" -o ./$(APPNAME) $(SRCPATH)

generate:
	go run pkg/config/schemas/generate.go

osx:
	@echo "Building osx binaries..."
	@$(call build,darwin,amd64,)
	@$(call zip,darwin,amd64,)

linux:
	@echo "Building linux binaries..."
	@$(call build,linux,amd64,)
	@$(call zip,linux,amd64,)
	@$(call build,linux,386,)
	@$(call zip,linux,386,)

win:
	@echo "Building windows binaries..."
	@$(call build,windows,amd64,.exe)
	@$(call zip,windows,amd64,.exe)

all: test osx linux win

clean:
	@rm -rf build/*