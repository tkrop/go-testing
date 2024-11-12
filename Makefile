SHELL := /bin/bash

# Include custom variables to modify behavior.
GITROOT ?= $(shell git rev-parse --show-toplevel 2>/dev/null || echo $(CURDIR))
ifneq ("$(wildcard $(GITROOT)/Makefile.vars)","")
	include Makefile.vars
endif

# Setup default go compiler environment.
export GO ?= go
export GOPATH ?= $(shell $(GO) env GOPATH)
export GOBIN ?= $(GOPATH)/bin

# Setup go-make version to use desired build and config scripts.
GOMAKE_DEP ?= github.com/tkrop/go-make@v0.0.115
INSTALL_FLAGS ?= -mod=readonly -buildvcs=auto
# Request targets from go-make targets target.
TARGETS := $(shell command -v $(GOBIN)/go-make >/dev/null || \
	$(GO) install $(INSTALL_FLAGS) $(GOMAKE_DEP) >/dev/stderr && \
	MAKEFLAGS="" $(GOBIN)/go-make show-targets 2>/dev/null)
# Declare all targets phony to make them available for auto-completion.
.PHONY:: $(TARGETS)

# Delegate all targets to go-make in a single stubbing call.
GOAL := $(firstword $(MAKECMDGOALS) all)
$(eval $(MAKECMDGOALS)::;@:)
$(GOAL):: $(GOBIN)/go-make
	@$(GOBIN)/go-make $(MAKECMDGOALS);
