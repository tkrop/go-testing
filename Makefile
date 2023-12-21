SHELL := /bin/bash

# Include custom variables to modify behavior.
ifneq ("$(wildcard Makefile.vars)","")
	include Makefile.vars
else
	$(warning warning: please customize variables in Makefile.vars)
endif

GOBIN ?= $(shell go env GOPATH)/bin
GOMAKE ?= github.com/tkrop/go-make@v0.0.12
TARGETS := $(shell command -v go-make >/dev/null || \
	go install $(GOMAKE) && go-make targets)

# Declare all targets phony to make them available for auto-completion.
.PHONY: $(TARGETS)

# Delegate all targets to go-make in one call.
# TODO: consider solution that does not delegate local goals.
$(firstword $(MAKECMDGOALS) all)::
	$(GOBIN)/go-make $(MAKEFLAGS) $(MAKECMDGOALS);
