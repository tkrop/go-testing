## Maintained by: github.com/tkrop/go-make
## Manual: http://github.com/tkrop/go-make/MANUAL.md
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

# Setup default temporary directory for go-make.
TMPDIR ?= /tmp
# Setup default go-make installation flags.
INSTALL_FLAGS ?= -mod=readonly -buildvcs=auto
# Setup go-make version to use desired build and config scripts.
GOMAKE_DEP ?= github.com/tkrop/go-make@v0.0.164
# Request targets from go-make show-targets target.
TARGETS := $(shell command -v $(GOBIN)/go-make >/dev/null || \
	$(GO) install $(INSTALL_FLAGS) $(GOMAKE_DEP) >&2 && \
	DIR="$(abspath $(TMPDIR))/go-make-$(USER)$(realpath $(CURDIR))" && \
	cat "$${DIR}/targets.make" 2>/dev/null; \
	MAKEFLAGS="" $(GOBIN)/go-make show-targets-make >/dev/null 2>&1 &)
# Declare all targets phony to make them available for auto-completion.
.PHONY:: $(TARGETS)

# Delegate all targets to go-make in a single stubbing call.
$(eval $(MAKECMDGOALS)::;@:)
$(firstword $(MAKECMDGOALS) all)::
	@+$(GOBIN)/go-make $(MAKECMDGOALS);
