SHELL := /bin/bash

# Include custom variables to modify behavior.
ifneq ("$(wildcard Makefile.vars)","")
	include Makefile.vars
else
	$(warning warning: please customize variables in Makefile.vars)
endif

# Setup desired go compiler environemt or defaults.
export GO ?= go
ifndef GOSETUP
  export GOBIN ?= $(shell $(GO) env GOPATH)/bin
else ifeq ($(GOSETUP),local)
  export GOBIN := $(DIR_BUILD)/bin
  export PATH := $(GOBIN):$(PATH)
else
  $(error error: unsupported go setup ($(GOSETUP)))
endif

# Setup go-make to utilize desired build and config scripts.
GOMAKE_DEP ?= github.com/tkrop/go-make@v0.0.41
# Request targets from go-make targets target.
TARGETS := $(shell command -v $(GOBIN)/go-make >/dev/null || \
	go install $(GOMAKE_DEP) >/dev/stderr && \
	$(GOBIN)/go-make targets 2>/dev/null)
# Declare all targets phony to make them available for auto-completion.
.PHONY:: $(TARGETS)

# Delegate all targets to go-make in a single call stubbing other targets.
$(eval $(wordlist 1,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))::;@:)
$(firstword $(MAKECMDGOALS) all)::
	@$(GOBIN)/go-make $(MAKECMDGOALS);
