SHELL := /bin/bash

GOBIN ?= $(shell go env GOPATH)/bin
GOMAKE := github.com/tkrop/go-make@latest
TARGETS := $(shell command -v go-make >/dev/null || \
	go install $(GOMAKE) && go-make list)


# Include custom variables to modify behavior.
ifneq ("$(wildcard Makefile.vars)","")
	include Makefile.vars
else
	$(warning warning: please customize variables in Makefile.vars)
endif


# Include standard targets from base makefile.
.PHONY: $(TARGETS) $(addprefix targets/,$(TARGETS))
$(TARGETS):; $(GOBIN)/go-make $(MAKEFLAGS) $(MAKECMDGOALS);
$(addprefix targets/,$(TARGETS)): targets/%:
	$(GOBIN)/go-make $(MAKEFLAGS) $*;


# Include custom targets to extend scripts.
ifneq ("$(wildcard Makefile.ext)","")
	include Makefile.ext
endif
