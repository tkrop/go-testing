# === do not change this Makefile! ===
# === extended via Makefile.{vars,defs,targets} ===
#
# This Makefile is majorly working by convention.
#
# Please visit
#
# https://github.bus.zalan.do/builder-knowledge/go-base/MAKEFILE.md
#
# for more information and help.
#
# Note: You can discover make targets using the tab-completion of your shell.
#
# Warning: The Makefile automatically installs a 'pre-commit' hook (overwritng
# any pre-existing hook) that runs 'make lint test' before allowing to commit.
#

SHELL := /bin/bash

RUNDIR := $(CURDIR)/run
BUILDIR := $(CURDIR)/build
CREDDIR := $(RUNDIR)/creds
TEMPDIR := $(RUNDIR)/temp

TEST_ALL := $(BUILDIR)/test-all.cover
TEST_UNIT := $(BUILDIR)/test-unit.cover
LINT_ALL := lint-src lint-api

# Include required custom variables.
ifneq ("$(wildcard Makefile.vars)","")
  include Makefile.vars
else
  $(error error: please define variables in Makefile.vars)
endif

# Setup sensible defaults for configuration variables.
TEST_TIMEOUT ?= 10s

CONTAINER ?= Dockerfile
REPOSITORY ?= $(shell git remote get-url origin | \
	sed "s/^https:\/\///; s/^git@//; s/.git$$//; s/:/\//;")
TEAM ?= $(shell cat .zappr.yaml | grep "X-Zalando-Team" | \
	sed "s/.*: *\([a-z-]*\).*/\1/")


IMAGE_PUSH ?= test
IMAGE_VERSION ?= snapshot

ifeq ($(words $(subst /, ,$(IMAGE_NAME))),3)
  IMAGE_HOST ?= $(wordlist 1,1,$(subst /, ,$(IMAGE_NAME)))
  IMAGE_TEAM ?= $(wordlist 2,2,$(subst /, ,$(IMAGE_NAME)))
  IMAGE_ARTIFACT ?= $(wordlist 3,3,$(subst /, ,$(IMAGE_NAME)))
else
  IMAGE_HOST ?= pierone.stups.zalan.do
  IMAGE_TEAM ?= $(TEAM)
  IMAGE_ARTIFACT ?= $(wordlist 3,3,$(subst /, ,$(REPOSITORY)))
endif
IMAGE ?= $(IMAGE_HOST)/$(IMAGE_TEAM)/$(IMAGE_ARTIFACT):$(IMAGE_VERSION)


DB_HOST ?= 127.0.0.1
DB_PORT ?= 5432
DB_NAME ?= db
DB_USER ?= user
DB_PASSWORD ?= pass
DB_VERSION ?= latest
DB_IMAGE ?= postgres:$(DB_VERSION)

AWS_SERVICES ?= sqs s3
AWS_VERSION ?= latest
AWS_IMAGE ?= localstack/localstack:$(AWS_VERSION)

# default target list for all and cdp builds.
TARGETS_ALL ?= init test lint build
TARGETS_CDP ?= clean clean-run init test lint \
	$(if $(filter $(IMAGE_PUSH),never),,\
	  $(if $(wildcard $(CONTAINER)),image-push,))
TARGETS_LINT ?= $(LINT_ALL)


# General setup of tokens for run-targets (not to be modified)

# Often used token setup functions.
define run-token-create
	mkdir -p $(CREDDIR); ztoken > $(CREDDIR)/token; echo "Bearer" > $(CREDDIR)/type
endef
define run-token-link
	test -n "$(1)" && test -n "$(2)" && test -n "$(3)" && ( \
		test -h "$(CREDDIR)/$(1)-$(2)" || ln -s type "$(CREDDIR)/$(1)-$(2)" && \
		test -h "$(CREDDIR)/$(1)-$(3)" || ln -s token "$(CREDDIR)/$(1)-$(3)" \
	) || test -n "$(1)" && test -n "$(2)" && test -z "$(3)" && ( \
		test -h "$(CREDDIR)/$(1)" || ln -s type "$(CREDDIR)/$(1)" && \
		test -h "$(CREDDIR)/$(2)" || ln -s token "$(CREDDIR)/$(2)" \
	) || test -n "$(1)" && test -z "$(2)" && test -z "$(3)" && ( \
		test -h "$(CREDDIR)/$(1)" || ln -s token "$(CREDDIR)/$(1)" \
	) || true
endef

# Stub definition for general setup in run-targets.
define run-setup
  true
endef

# Stub definition for common variables in run-targets.
define run-vars
endef

# Stub definition for local runtime variables in run-targets.
define run-vars-local
endef

# Stub definition for container specific runtime variables in run-targets.
define run-vars-image
  $(call run-vars-docker)
endef

# Stub definition to setup aws localstack run-target.
define run-aws-setup
  true
endef

# Include function definitions to override defaults.
ifneq ("$(wildcard Makefile.defs)","")
  include Makefile.defs
else
  $(warning warning: please define custom functions in Makefile.defs)
endif


# Setup default environment variables.
COMMANDS := $(shell grep -lr "func main()" cmd/*/main.go 2>/dev/null | \
	sed -E "s/^cmd\/([^/]*)\/main.go$$/\1/;" | sort -u)
SOURCES := $(shell find . -name "*.go" ! -name "mock_*_test.go")

# Setup golang mock setup environment.
MOCK_MATCH_DST := ^.\/(.*)\/(.*):\/\/go:generate.*-destination=([^ ]*).*$$
MOCK_MATCH_SRC := ^.\/(.*)\/(.*):\/\/go:generate.*-source=([^ ]*).*$$
MOCK_TARGETS := $(shell grep "//go:generate.*mockgen" $(SOURCES) | \
	sed -E "s/$(MOCK_MATCH_DST)/\1\/\3=\1\/\2/;" | sort -u)
MOCK_SOURCES := $(shell grep "//go:generate.*mockgen.*-source" $(SOURCES) | \
	sed -E "s/$(MOCK_MATCH_SRC)/\1\/\3/;" | sort -u | \
	xargs realpath --relative-base=.)
MOCKS := $(shell for TARGET in $(MOCK_TARGETS); \
	do echo "$${TARGET%%=*}"; done | sort -u)


# Setup phony make targets to always be executed.
.PHONY: all cdp bump release
.PHONY: update update-go update-deps update-make
.PHONY: clean clean-init clean-build clean-run
.PHONY: $(addprefix clean-run-, $(COMMANDS) db aws)
.PHONY: init init-tools init-hooks init-packages init-sources
.PHONY: test test-all test-unit test-clean test-upload test-cover
.PHONY: lint lint-src lint-api format
.PHONY: build build-native build-linux build-image build-docker
.PHONY: $(addprefix build-, $(COMMANDS))
.PHONY: image image-build image-push docker docker-build docker-push
.PHONY: run-native run-image run-docker run-clean
.PHONY: $(addprefix run-, $(COMMANDS) db aws)
.PHONY: $(addprefix run-go-, $(COMMANDS))
.PHONY: $(addprefix run-image-, $(COMMANDS))
.PHONY: $(addprefix run-clean-, $(COMMANDS) db aws)


# Setup docker or podman command.
IMAGE_CMD ?= $(shell command -v podman || command -v docker)
ifndef IMAGE_CMD
  $(error error: docker/podman command not found)
endif


# Helper definitions to resolve match position of word in text.
define pos-recursion
  $(if $(findstring $(1),$(2)),$(call pos-recursion,$(1),\
    $(wordlist 2,$(words $(2)),$(2)),x $(3)),$(3))
endef
define pos
  $(words $(call pos-recursion,$(1),$(2)))
endef


# match commands that support arguments ...
CMDMATCH = $(or \
	  $(findstring run-,$(MAKECMDGOALS)),\
	  $(findstring test-,$(MAKECMDGOALS)),\
	  $(findstring bump,$(MAKECMDGOALS)),\
	)%
# If any argument contains "run-", "test-", "bump" ...
ifneq ($(CMDMATCH),%)
  CMD := $(filter $(CMDMATCH),$(MAKECMDGOALS))
  POS = $(call pos,$(CMD), $(MAKECMDGOALS))
  # ... then use the rest as arguments for "run/test-" ...
  CMDARGS := $(wordlist $(shell expr $(POS) + 1),\
    $(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets.
  $(eval $(CMDARGS):;@:)
  RUNARGS ?= $(CMDARGS)
  $(info translated targets to arguments (ARGS=[$(RUNARGS)]))
endif


# Initialize golang modules - if not done before.
ifneq ($(shell ls go.mod), go.mod)
  $(shell go mod init $(REPOSITORY))
endif

# Setup go to use desired and consistent golang versions.
GOVERSION := $(shell go version | sed -Ee "s/.*go([0-9]+\.[0-9]+).*/\1/")
GOVERSION_MOD := $(shell grep "^go [1-9.]*$$" go.mod | cut -f2 -d' ')
GOVERSION_DELIVERY := $(shell if [ -f delivery.yaml ]; then \
    grep -o "cdp-runtime/go-[0-9.]*" delivery.yaml | grep -o "[0-9.]*" | sort -u; \
  else echo $(GOVERSION); fi)
ifneq (update-go,$(MAKECMDGOALS))
  ifneq ($(firstword $(GOVERSION_DELIVERY)), $(GOVERSION_DELIVERY))
    $(error "inconsistent go versions: delivery.yaml uses $(GOVERSION_DELIVERY)")
  endif
  ifneq ($(GOVERSION), $(GOVERSION_DELIVERY))
    $(error "unsupported go version $(GOVERSION): delivery.yaml requires $(GOVERSION_DELIVERY)")
  endif
  ifneq ($(GOVERSION), $(GOVERSION_MOD))
    $(error "unsupported go version $(GOVERSION): go.mod requires $(GOVERSION_MOD)")
  endif
endif

# Export private repositories not to be downloaded.
export GOPRIVATE := github.bus.zalan.do


# Standard targets for default build processes.
all: $(TARGETS_ALL)
cdp: $(TARGETS_CDP)


# Update dependencies of all packages.
update: update-deps
update-go:
	@sed -i "s/go $(GOVERSION_MOD)/go $(GOVERSION)/" go.mod; \
	sed -E -i "s/(cdp-runtime\/go)[0-9.-]*/\1-$(GOVERSION)/" delivery.yaml; \

update-make:
	@TEMPDIR=$$(mktemp -d) && echo "update Makefile" &&  \
	BASEREPO=git@github.bus.zalan.do:builder-knowledge/go-base.git && \
	git clone --no-checkout --depth 1 $${BASEREPO} $${TEMPDIR} 2>/dev/null && ( \
	  cd $${TEMPDIR}; \
	    git show HEAD:Makefile > Makefile; \
		git show HEAD:MAKEFILE.md > MAKEFILE.md; \
	  cd - \
	); cp $${TEMPDIR}/Makefile $${TEMPDIR}/MAKEFILE.md .; \
	rm -rf $${TEMPDIR}; \

update-make-would-be-better:
	BASEREPO=git://github.bus.zalan.do/builder-knowledge/go-base.git; \
	git archive --remote=$${BASEREPO} HEAD Makefile | tar -xvf -; \

update-deps:
	@for DIR in $$(find . -name "*.go" | xargs dirname | sort -u); do \
	  echo -n "update: $${DIR} -> "; cd $${DIR} && \
	  go get -u && go mod tidy -compat=${GOVERSION} && \
	  cd -; \
	done; \


# Bump version to prepare release of software.
bump:
	@if [ -z "$(RUNARGS)" ]; then \
	  echo "error: missing new version"; exit 1; \
	fi; \
	VERSION="$$(echo $(RUNARGS) | \
	  grep -E -o "[0-9]+\.[0-9]+\.[0-9]+(-.*)?")"; \
	if [ -z "$${VERSION}" ]; then \
	  echo "error: invalid new version [$(RUNARGS)]"; exit 1; \
	fi; \
	echo "$${VERSION}" > VERSION; \
	echo "Bumped version to $${VERSION} for auto release!"; \

# Release fixed version of software.
release:
	@VERSION="$$(cat VERSION)" && \
	if [ -n "$${VERSION}" -a \
	     -z "$$(git tag -l "v$${VERSION}")" ]; then \
	  git gh-release "v$${VERSION}" && \
	  echo "Added release tag v$${VERSION} to repository!"; \
	fi; \


# Default cleanup of sources.
clean: clean-build
clean-build: clean-init
	find . -name "mock_*_test.go" -exec rm -v {} \;; \

clean-init:
	rm -vrf $(RUNDIR) $(BUILDIR);
	rm -vrf .git/hooks/pre-commit;

# Clean up all running container images.
clean-run: $(addprefix clean-run-, $(COMMANDS) db aws)
$(addprefix clean-run-, $(COMMANDS) db aws): clean-run-%: run-clean-%


# Initialize tooling and packages for building.
init: clean-init init-tools init-hooks init-packages

init-tools:
	go install github.com/golang/mock/mockgen@latest;
	go install golang.org/x/lint/golint@latest;
	go install github.com/zalando/zally/cli/zally@latest;
	go install golang.org/x/tools/cmd/goimports@latest;
	go mod tidy -compat=${GOVERSION};

init-hooks: .git/hooks/pre-commit
.git/hooks/pre-commit:
	@echo -ne "#!/bin/sh\nmake lint test-unit" >$@; chmod 755 $@;

init-packages:
	go build ./...;

init-sources: $(MOCKS)
$(MOCKS): go.sum $(MOCK_SOURCES)
	GO111MODULE=on go generate "$(shell echo $(MOCK_TARGETS) | \
	  sed -E "s:.*$@=([^ ]*).*$$:\1:;")";


test: test-all
test-all: test-clean init-sources $(TEST_ALL)
test-unit: test-clean init-sources $(TEST_UNIT)
test-clean:
	@if [ -f "$(TEST_ALL)" ]; then rm -v $(TEST_ALL); fi; \
	 if [ -f "$(TEST_UNIT)" ]; then rm -v $(TEST_UNIT); fi;
test-upload:
test-cover:
	@if [ "$(TEST_ALL)" -nt "$(TEST_UNIT)" ]; then \
	  go tool cover -html=$(TEST_ALL); \
	else \
	  go tool cover -html=$(TEST_UNIT); \
	fi; \

TESTFLAGS ?= -race -mod=readonly -count=1
ifneq ($(RUNARGS),)
  TESTARGS ?= $(addprefix ./,$(RUNARGS))
else
  TESTARGS ?= ./...
endif

$(TEST_ALL): $(SOURCES) init-sources $(TEST_DEPS)
	@if [ ! -d "$(BUILDIR)" ]; then mkdir -p $(BUILDIR); fi;
	go test $(TESTFLAGS) -timeout $(TEST_TIMEOUT) \
	  -cover -coverprofile $@ $(TESTARGS);
$(TEST_UNIT): $(SOURCES) init-sources
	@if [ ! -d "$(BUILDIR)" ]; then mkdir -p $(BUILDIR); fi;
	go test $(TESTFLAGS) -timeout $(TEST_TIMEOUT) \
	  -cover -coverprofile $@ -short $(TESTARGS);


lint: $(TARGETS_LINT)
lint-src: init-sources
	go vet $$(go list ./...);
	golint -set_exit_status $$(go list ./...);
	@GOIMPORT="$$(goimports -l -local "$(REPOSITORY)" \
	    $$(find . -name "*.go" ! -name "mock_*_test.go"))"; \
	if [ -n "$${GOIMPORT}" ]; then \
	  echo -e "Error: Sources are not formatted correctly:"; \
	  echo -e "$${GOIMPORT}\n\tTo fix run: make format"; false; \
	fi;
lint-api:
	@ARGS=("--linter-service" "https://infrastructure-api-linter.zalandoapis.com"); \
	if command -v ztoken > /dev/null; then ARGS+=("--token" "$$(ztoken)"); fi; \
	for APISPEC in $$(find zalando-apis -name "*.yaml" > /dev/null 2>&1); do \
	  echo "check API: zally \"$${APISPEC}\""; \
	  zally "$${ARGS[@]}" lint "$${APISPEC}" || exit 1; \
	done;

format:
	goimports -w -local "$(REPOSITORY)" $$(find . -name "*.go" ! -name "mock_*_test.go")


# Setup container specific build flags
BUILDOS ?= ${shell grep "^FROM [^ ]*$$" $(CONTAINER) 2>/dev/null | \
	grep -v " as " | sed -e "s/.*\(alpine\|ubuntu\).*/\1/g"}
BUILDARCH ?= amd64
ifeq ($(BUILDOS),alpine)
  BUILDFLAGS ?= -v -mod=readonly
  GOCGO := 0
else
  BUILDFLAGS ?= -v -race -mod=readonly
  GOCGO := 1
endif
GOOS ?= linux
GOARCH := $(BUILDARCH)

# Define flags propagate versions to build commands.
LDFLAGS ?= -X $(shell go list ./... | grep "config$$").Version=$(IMAGE_VERSION) \
           -X $(shell go list ./... | grep "config$$").GitHash=$(shell git rev-parse --short HEAD)

# Build targets for native platform builds and linux builds.
build: build-native
build-image: image-build
build-native: $(addprefix build/, $(COMMANDS))
$(addprefix build-, $(COMMANDS)): build-%: build/%
build/%: cmd/%/main.go $(SOURCES)
	@mkdir -p "$(dir $@)"
	CGO_ENABLED=1 go build \
	  $(BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $@ $<;

build-linux: $(addprefix $(BUILDIR)/linux/, $(COMMANDS))
$(BUILDIR)/linux/%: cmd/%/main.go $(SOURCES)
	@mkdir -p $(dir $@)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(GOCGO) go build \
	  $(BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $@ $<;


# Image build and push targets.
image: image-build
image-build: $(CONTAINER) build-linux
	@if [ "$(IMAGE_PUSH)" == "never" ]; then \
	  echo "We never build images, aborting."; exit 0; \
	else \
	  $(IMAGE_CMD) build -t $(IMAGE) -f $< .; \
	fi; \

image-push: image-build
	@if [ "$(IMAGE_PUSH)" == "never" ]; then \
	  echo "We never push images, aborting."; exit 0; \
	elif [ "$(IMAGE_VERSION)" == "snapshot" ]; then \
	  echo "We never push snapshot images, aborting."; exit 0; \
	elif [ -n "$(CDP_PULL_REQUEST_NUMBER)" -a "$(IMAGE_PUSH)" != "test" ]; then \
	  echo "We never push pull request images, aborting."; exit 0; \
	fi; \
	$(IMAGE_CMD) push $(IMAGE); \


# Target for running a postgres database.
run-db:
	@if [[ ! "$(TEST_DEPS) $(RUN_DEPS)" =~ run-db ]]; then exit 0; fi; \
	mkdir -p $(RUNDIR) && HOST="127.0.0.1" && \
	echo "info: ensure $(DB_IMAGE) running on $${HOST}:$(DB_PORT)"; \
	if [ -n "$$($(IMAGE_CMD) ps | grep "$(DB_IMAGE).*$${HOST}:$(DB_PORT)")" ]; then \
	  echo "warning: port allocated, try existing db container!"; exit 0; \
	fi; \
	$(IMAGE_CMD) start ${IMAGE_ARTIFACT}-db 2>/dev/null || ( \
	$(IMAGE_CMD) run -dt \
	  --name ${IMAGE_ARTIFACT}-db \
	  --publish $${HOST}:$(DB_PORT):5432 \
	  --env POSTGRES_USER="$(DB_USER)" \
	  --env POSTGRES_PASSWORD="$(DB_PASSWORD)" \
	  --env POSTGRES_DB="$(DB_NAME)" \
	  $(DB_IMAGE) $(RUNARGS) 2>&1 & \
	until [ "$$($(IMAGE_CMD) inspect -f {{.State.Running}} \
	         $(IMAGE_ARTIFACT)-db 2>/dev/null)" == "true" ]; \
	do echo "waiting for db container" >/dev/stderr; sleep 1; done && \
	until $(IMAGE_CMD) exec $(IMAGE_ARTIFACT)-db \
	  pg_isready -h localhost -U $(DB_USER) -d $(DB_NAME); \
	do echo "waiting for db service" >/dev/stderr; sleep 1; done) |\
	tee -a $(RUNDIR)/$(IMAGE_ARTIFACT)-db; \

# Target for running the AWS localstack.
run-aws:
	@if [[ ! "$(TEST_DEPS) $(RUN_DEPS)" =~ run-aws ]]; then exit 0; fi; \
	mkdir -p $(RUNDIR) && HOST="127.0.0.1" && \
	echo "info: ensure $(AWS_IMAGE) is running on $${HOST}:4566/4571" && \
	if [ -n "$$($(IMAGE_CMD) ps | \
	    grep "$(AWS_IMAGE).*$${HOST}:4566.*$${HOST}:4571")" ]; then \
	  echo "warning: ports allocated, try existing aws container!"; \
	  $(call run-aws-setup); exit 0; \
	fi; \
	$(IMAGE_CMD) start ${IMAGE_ARTIFACT}-aws 2>/dev/null || ( \
	$(IMAGE_CMD) run -dt --name ${IMAGE_ARTIFACT}-aws \
	  --publish $${HOST}:4566:4566 \
	  --publish $${HOST}:4571:4571 \
	  --env SERVICES="$(AWS_SERVICES)" \
	  $(AWS_IMAGE) $(RUNARGS) 2>&1 && \
	until [ "$$($(IMAGE_CMD) inspect -f {{.State.Running}} \
	         $(IMAGE_ARTIFACT)-aws 2>/dev/null)" == "true" ]; \
	do echo "waiting for aws container" >/dev/stderr; sleep 1; done && \
	until [ -n "$$($(IMAGE_CMD) exec $(IMAGE_ARTIFACT)-aws \
	         curl -is http://$${HOST}:4566 | grep -o "HTTP/1.1 200")" ]; \
	do echo "waiting for aws service" >/dev/stderr; sleep 1; done && \
	$(call run-aws-setup)) | \
	tee -a $(RUNDIR)/$(IMAGE_ARTIFACT)-aws.log; \

# Targets for running the provide commands natively.
$(addprefix run-, $(COMMANDS)): run-%: build/% $(RUN_DEPS)
	@mkdir -p $(RUNDIR) && $(call run-setup);
	$(call run-vars) $(call run-vars-local) \
	  $(BUILDIR)/$* $(RUNARGS) 2>&1 | \
	  tee -a $(RUNDIR)/$(IMAGE_ARTIFACT)-$*.log; \
	exit $${PIPESTATUS[0]};

# Targets for running the provide commands via golang.
$(addprefix run-go-, $(COMMANDS)): run-go-%: $(BUILDIR)/% $(RUN_DEPS)
	@mkdir -p $(RUNDIR) && $(call run-setup);
	$(call run-vars) $(call run-vars-local) \
	  go run cmd/$*/main.go $(RUNARGS) 2>&1 | \
	  tee -a $(RUNDIR)/$(IMAGE_ARTIFACT)-$*.log; \
	exit $${PIPESTATUS[0]};

# Target to run commands in container images.
$(addprefix run-image-, $(COMMANDS)): run-image-%: $(RUN_DEPS)
	@mkdir -p $(RUNDIR) && $(call run-setup); \
	trap "$(IMAGE_CMD) rm $(IMAGE_ARTIFACT)-$* >/dev/null" EXIT; \
	trap "$(IMAGE_CMD) kill $(IMAGE_ARTIFACT)-$* >/dev/null" INT TERM;
	$(IMAGE_CMD) run --name $(IMAGE_ARTIFACT)-$* --network=host \
	  --volume $(CREDDIR):/meta/credentials --volume $(RUNDIR)/temp:/tmp \
      $(call run-vars, --env) $(call run-vars-image, --env) \
      $(IMAGE) /$* $(RUNARGS) 2>&1 | \
	tee -a $(RUNDIR)/$(IMAGE_ARTIFACT)-$*.log; \
	exit $${PIPESTATUS[0]};

# Clean up all running container images.
run-clean: $(addprefix run-clean-, $(COMMANDS) db aws)
$(addprefix run-clean-, $(COMMANDS) db aws): run-clean-%:
	@echo "check container $(IMAGE_ARTIFACT)-$*"; \
	if [ -n "$$($(IMAGE_CMD) ps | grep "$(IMAGE_ARTIFACT)-$*")" ]; then \
	  $(IMAGE_CMD) kill $(IMAGE_ARTIFACT)-$* > /dev/null && \
	  echo "killed container $(IMAGE_ARTIFACT)-$*"; \
	fi; \
	if [ -n "$$($(IMAGE_CMD) ps -a | grep "$(IMAGE_ARTIFACT)-$*")" ]; then \
	  $(IMAGE_CMD) rm $(IMAGE_ARTIFACT)-$* > /dev/null && \
	  echo "removed container $(IMAGE_ARTIFACT)-$*"; \
	fi; \


# Include custom targets to extend scripts.
ifneq ("$(wildcard Makefile.targets)","")
  include Makefile.targets
endif
