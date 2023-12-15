SHELL := /bin/bash

GOBIN ?= $(shell go env GOPATH)/bin
GOMAKE ?= github.com/tkrop/go-make@latest
TARGETS := $(shell command -v go-make >/dev/null || \
	go install $(GOMAKE) && go-make targets)


# Include custom variables to modify behavior.
ifneq ("$(wildcard Makefile.vars)","")
	include Makefile.vars
else
	$(warning warning: please customize variables in Makefile.vars)
endif


# Include standard targets from go-make providing group targets as well as
# single target targets. The group target is used to delegate the remaining
# request targets, while the single target can be used to define the
# precondition of custom target.
.PHONY: $(TARGETS) $(addprefix target/,$(TARGETS))
$(eval $(lastwords $(MAKECMDGOALS)):;@:)
$(firstword $(MAKECMDGOALS)):
	$(GOBIN)/go-make $(MAKEFLAGS) $(MAKECMDGOALS);
$(addprefix target/,$(TARGETS)): target/%:
	$(GOBIN)/go-make $(MAKEFLAGS) $*;


# Include custom targets to extend scripts.
ifneq ("$(wildcard Makefile.ext)","")
	include Makefile.ext
endif
