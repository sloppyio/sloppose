.DEFAULT_GOAL := all

BUILDPATH := ./build
SRCPATH := ./cmd
APPNAME := sloppose

.PHONY: all test build-dev osx linux win clean

define build
	GOOS=$(1) GOARCH=$(2) go build -o $(BUILDPATH)/$(APPNAME)_$(1)_$(2)$(3) $(SRCPATH)
endef

define zip
	cd build && zip $(1)_$(2).zip $(APPNAME)_$(1)_$(2)$(3) && rm $(APPNAME)_$(1)_$(2)$(3)
endef

test:
	go test -v -race ./pkg/...

build-dev:
	go build -o ./$(APPNAME) $(SRCPATH)

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