# === do not change this Makefile! ===
# See MAKEFILE.md for documentation.

SHELL := /bin/bash

RUNDIR := $(CURDIR)/run
BUILDDIR := $(CURDIR)/build
CREDDIR := $(RUNDIR)/creds

# Include required custom variables.
ifneq ("$(wildcard Makefile.vars)","")
  include Makefile.vars
else
  $(info info: please define variables in Makefile.vars)
endif

# Setup sensible defaults for configuration variables.
CODE_QUALITY ?= base
TEST_TIMEOUT ?= 10s

FILE_CONTAINER ?= Dockerfile
FILE_GOLANGCI ?= .golangci.yaml
FILE_MARKDOWN ?= .markdownlint.yaml
FILE_GITLEAKS ?= .gitleaks.toml
FILE_REVIVE ?= revive.toml
FILE_CODACY ?= .codacy.yaml
FILE_DELIVERY ?= delivery.yaml
FILE_DELIVERY_REGEX ?= (cdp-runtime\/go-|go-version: \^?)[0-9.]*

REPOSITORY ?= $(shell git remote get-url origin | \
	sed "s/^https:\/\///; s/^git@//; s/.git$$//; s/:/\//")
GITHOSTNAME ?= $(word 1,$(subst /, ,$(REPOSITORY)))
GITORGNAME ?= $(word 2,$(subst /, ,$(REPOSITORY)))
GITREPONAME ?= $(word 3,$(subst /, ,$(REPOSITORY)))

TEAM ?= $(shell cat .zappr.yaml | grep "X-Zalando-Team" | \
	sed "s/.*:[[:space:]]*\([a-z-]*\).*/\1/")

TOOLS_GO ?= $(TOOLS_GOGEN) \
	github.com/golangci/golangci-lint/cmd/golangci-lint \
	github.com/zalando/zally/cli/zally \
	github.com/mgechev/revive@v1.2.3 \
	github.com/securego/gosec/v2/cmd/gosec \
	github.com/tsenart/deadcode \
	honnef.co/go/tools/cmd/staticcheck \
	github.com/zricethezav/gitleaks/v8 \
	github.com/icholy/gomajor \
	github.com/golang/mock/mockgen \
	github.com/tkrop/go-testing/cmd/mock

# function for correct gol-package handling.
define go-pkg
$(shell awk -v mode="$(1)" -v filter="$(3)" ' \
  BEGIN { FS = "[/@]"; RS = "[ \n\r]" } { \
	field = NF; \
	if (version = ($$0 ~ "@")) { field--; } \
	if ($$(field) ~ "v[0-9]+") { field--; } \
	if (!filter || ($$(field) ~ filter)) { \
	  if (mode == "cmd") { \
	    print $$(field); \
	  } else if (mode == "normal") { \
		if (version) { print $$0; } else { \
		  print $$0 "@latest"; \
		}; \
	  } else if (mode == "strip") { \
		if (version) { \
		  print substr($$0, 1, index($$0, "@") - 1); \
		} else { print $$0 } \
	  } \
	} \
  }' <<<"$(2)")
endef

TOOLS_NPMLINT ?= markdownlint-cli
TOOLS_NPM := $(TOOLS_NPMLINT)

IMAGE_PUSH ?= pulls
IMAGE_VERSION ?= snapshot

ifeq ($(words $(subst /, ,$(IMAGE_NAME))),3)
  IMAGE_HOST ?= $(word 1,$(subst /, ,$(IMAGE_NAME)))
  IMAGE_TEAM ?= $(word 2,$(subst /, ,$(IMAGE_NAME)))
  IMAGE_ARTIFACT ?= $(word 3,$(subst /, ,$(IMAGE_NAME)))
else
  IMAGE_HOST ?= pierone.stups.zalan.do
  IMAGE_TEAM ?= $(TEAM)
  IMAGE_ARTIFACT ?= $(GITREPONAME)
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

# Setup codacy integration.
CODACY ?= enabled
ifdef CDP_PULL_REQUEST_NUMBER
  CODACY_CONTINUE ?= true
else
  CODACY_CONTINUE ?= false
endif
CODACY_PROVIDER ?= ghe
CODACY_USER ?= $(GITORGNAME)
CODACY_PROJECT ?= $(GITREPONAME)
CODACY_API_BASE_URL ?= https://codacy.bus.zalan.do
CODACY_CLIENTS ?= aligncheck deadcode
CODACY_BINARIES ?= gosec staticcheck
CODACY_GOSEC_VERSION ?= 0.4.5
CODACY_STATICCHECK_VERSION ?= 3.0.12

# Default target list for all and cdp builds.
TARGETS_ALL ?= init test lint build image
TARGETS_INIT ?= init-hooks init-packages \
	$(if $(filter $(CODACY),enabled),init-codacy,)
TARGETS_CLEAN ?= clean-build
TARGETS_UPDATE ?= update-go update-deps update-make
TARGETS_COMMIT ?= test-go test-unit lint-leaks? lint-$(CODE_QUALITY) lint-markdown
TARGETS_TEST ?= test-go test-all test-upload
TARGETS_TEST ?= test-all $(if $(filter $(CODACY),enabled),test-upload,)
TARGETS_LINT ?= lint-$(CODE_QUALITY) lint-leaks? lint-markdown lint-apis \
    $(if $(filter $(CODACY),enabled),lint-codacy,)

UPDATE_MAKE ?= Makefile MAKEFILE.md $(FILE_GOLANGCI) $(FILE_MARKDOWN) \
	$(FILE_GITLEAKS) $(FILE_CODACY) $(FILE_REVIVE)

# Initialize golang modules - if not done before.
ifneq ($(shell ls go.mod), go.mod)
  $(shell go mod init $(REPOSITORY))
endif

# Setup go to use desired and consistent go versions.
GOVERSION := $(shell go version | sed -E "s/.*go([0-9]+\.[0-9]+).*/\1/")

# Export private repositories not to be downloaded.
export GOPRIVATE := github.bus.zalan.do
export GOBIN ?= $(shell go env GOPATH)/bin


# General setup of tokens for run-targets (not to be modified)

# Often used token setup functions.
define run-token-create
	ztoken > $(CREDDIR)/token; echo "Bearer" > $(CREDDIR)/type
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
  $(info info: please define custom functions in Makefile.defs)
endif

# Setup conversion variables.
define upper
$(shell echo "$(1)" | tr '[:lower:]' '[:upper:]')
endef


# Setup default environment variables.
COMMANDS := $(shell grep -lr "func main()" cmd/*/main.go 2>/dev/null | \
	sed -E "s/^cmd\/([^/]*)\/main.go$$/\1/" | sort -u)
SOURCES := $(shell find . -name "*.go" ! -name "mock_*_test.go")

# Setup optimized golang mock setup environment.
MOCK_MATCH_DST := ^.\/(.*)\/(.*):\/\/go:generate.*-destination=([^ ]*).*$$
MOCK_MATCH_SRC := ^.\/(.*)\/(.*):\/\/go:generate.*-source=([^ ]*).*$$
MOCK_TARGETS := $(shell grep "//go:generate[[:space:]]*mockgen" $(SOURCES) | \
	sed -E "s/$(MOCK_MATCH_DST)/\1\/\3=\1\/\2/" | sort -u)
MOCK_SOURCES := $(shell grep "//go:generate[[:space:]]*mockgen.*-source" $(SOURCES) | \
	sed -E "s/$(MOCK_MATCH_SRC)/\1\/\3/" | sort -u | \
	xargs -r readlink -f | sed "s|$(PWD)/||g")
MOCKS := $(shell for TARGET in $(MOCK_TARGETS); \
	do echo "$${TARGET%%=*}"; done | sort -u)


# Prepare phony make targets lists.
TARGETS_INIT_CODACY := $(addprefix init-, $(CODACY_BINARIES))
TARGETS_LINT_CODACY_CLIENTS := $(addprefix lint-, $(CODACY_CLIENTS))
TARGETS_LINT_CODACY_BINARIES := $(addprefix lint-, $(CODACY_BINARIES))
TARGETS_IMAGE := $(shell find * -type f -name "$(FILE_CONTAINER)*")
TARGETS_IMAGE_BUILD := $(addprefix image-build/, $(TARGETS_IMAGE))
TARGETS_IMAGE_PUSH := $(addprefix image-push/, $(TARGETS_IMAGE))
TARGETS_BUILD := $(addprefix build-, $(COMMANDS))
TARGETS_BUILD_LINUX := $(addprefix $(BUILDDIR)/linux/, $(COMMANDS))
TARGETS_INSTALL := $(addprefix install-, $(COMMANDS))
TARGETS_INSTALL_GO := $(addprefix install-, $(call go-pkg,cmd,$(TOOLS_GO)))
TARGETS_INSTALL_NPM := $(addprefix install-, $(TOOLS_NPM:-cli=))
TARGETS_INSTALL_ALL := $(TARGETS_INSTALL) $(TARGETS_INSTALL_GO) $(TARGETS_INSTALL_NPM)
TARGETS_UNINSTALL := $(addprefix uninstall-, $(COMMANDS))
TARGETS_UNINSTALL_GO := $(addprefix uninstall-, $(call go-pkg,cmd,$(TOOLS_GO)))
TARGETS_UNINSTALL_NPM := $(addprefix uninstall-, $(TOOLS_NPM:-cli=))
TARGETS_UNINSTALL_CODACY := $(addprefix uninstall-codacy-, $(CODACY_BINARIES))
TARGETS_UNINSTALL_ALL := $(TARGETS_UNINSTALL) $(TARGETS_UNINSTALL_GO) \
	$(TARGETS_UNINSTALL_NPM) $(TARGETS_UNINSTALL_CODACY)
TARGETS_RUN := $(addprefix run-, $(COMMANDS))
TARGETS_RUN_GO := $(addprefix run-go-, $(COMMANDS))
TARGETS_RUN_IMAGE := $(addprefix run-image-, $(COMMANDS))
TARGETS_RUN_CLEAN := $(addprefix run-clean-, $(COMMANDS) db aws)
TARGETS_CLEAN_ALL := clean-init $(TARGETS_CLEAN) clean-run $(TARGETS_UNINSTALL_ALL)
TARGETS_CLEAN_RUN := $(addprefix clean-run-, $(COMMANDS) db aws)
TARGETS_UPDATE_MAKE := $(addprefix update/,$(UPDATE_MAKE))
TARGETS_UPDATE_MAKE? := $(addsuffix ?,$(TARGETS_UPDATE_MAKE))
TARGETS_UPDATE_ALL := update-go update-deps update-make update-tools
TARGETS_UPDATE_ALL? := udate-go? update-deps? update-make?
TARGETS_UPDATE? := $(filter $(addsuffix ?,$(TARGETS_UPDATE)), \
	$(TARGETS_UPDATE_ALL?) $(TARGETS_UPDATE_MAKE?))

# Setup phony make targets to always be executed.
.PHONY: all all-clean help list commit bump release
.PHONY: init init-all init-hooks init-packages init-sources
.PHONY: init-codacy $(TARGETS_INIT_CODACY)
.PHONY: test test-all test-unit test-bench test-clean test-cover
.PHONY: test-upload test-go
.PHONY: lint lint-min lint-base lint-plus lint-max lint-all lint-config
.PHONY: lint-markdown lint-apis lint-codacy lint-revive
.PHONY: $(TARGETS_LINT_CODACY_CLIENTS) $(TARGETS_LINT_CODACY_BINARIES)
.PHONY: build build-native build-linux build-image $(TARGETS_BUILD)
.PHONY: image image-build $(TARGETS_IMAGE_BUILD)
.PHONY: image-push $(TARGETS_IMAGE_PUSH)
.PHONY: install install-all $(TARGETS_INSTALL_ALL)
.PHONY: uninstall uninstall-all $(TARGETS_UNINSTALL_ALL)
.PHONY: run-native run-image run-clean $(TARGETS_RUN) run-db run-aws
.PHONY: $(TARGETS_RUN_GO) $(TARGETS_RUN_IMAGE) $(TARGETS_RUN_CLEAN)
.PHONY: clean clean-all $(TARGETS_CLEAN)
.PHONY: $(TARGETS_CLEAN_ALL) $(TARGETS_CLEAN_RUN)
.PHONY: update update? update-all update-all? update-base
.PHONY: $(TARGETS_UPDATE_ALL) $(TARGETS_UPDATE_ALL?)
.PHONY: $(TARGETS_UPDATE_MAKE) $(TARGETS_UPDATE_MAKE?)


# Setup docker or podman command.
IMAGE_CMD ?= $(shell command -v docker || command -v podman)
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
	  $(findstring run ,$(MAKECMDGOALS)),\
	  $(findstring test-,$(MAKECMDGOALS)),\
	  $(findstring test ,$(MAKECMDGOALS)),\
	  $(findstring lint-,$(MAKECMDGOALS)),\
	  $(findstring lint ,$(MAKECMDGOALS)),\
	  $(findstring bump,$(MAKECMDGOALS)),\
	  $(findstring update-,$(MAKECMDGOALS)),\
	  $(findstring update ,$(MAKECMDGOALS)),\
	  $(findstring list ,$(MAKECMDGOALS)),\
	)%
# If any argument contains "run*", "test*", "lint*", "bump" ...
ifneq ($(CMDMATCH),%)
  CMD := $(filter $(CMDMATCH),$(MAKECMDGOALS))
  POS = $(call pos,$(CMD), $(MAKECMDGOALS))
  # ... then use the rest as arguments for "run/test-" ...
  CMDARGS := $(wordlist $(shell expr $(POS) + 1),\
    $(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets.
  $(eval $(CMDARGS):;@:)
  RUNARGS ?= $(CMDARGS)
  $(shell if [ -n "$(RUNARGS)" ]; then \
    echo "info: captured arguments [$(RUNARGS)]" >/dev/stderr; \
  fi)
endif


## Standard: default targets to test, lint, and build.

#@ executes the default targets.
all: $(TARGETS_ALL)
#@ executes the default targets after cleaning up.
all-clean: clean clean-run all
#@ executes the pre-commit check targets.
commit: $(TARGETS_COMMIT)
$(BUILDDIR) $(RUNDIR) $(CREDDIR):
	@if [ ! -d "$@" ]; then mkdir -p $@; fi;


## Support: targets to support processes and users.

#@ prints this help.
help:
	@cat Makefile | awk ' \
	  BEGIN { \
	    printf("\n\033[1mUsage:\033[0m $(MAKE) \033[36m<target>\033[0m\n"); \
	  }; \
	  /^## .*/ { \
	    if (i = index($$0, ":")) { \
	      printf("\n\033[1m%s\033[0m%s\n", \
	        substr($$0, 4, i - 4), substr($$0, i)); \
	    } else { \
	      printf("\n\033[1m%s\033[0m\n", substr($$0, 4)); \
	    }; next; \
	  }; \
	  /^#@ / { \
	    if (i = index($$0, ":")) { \
	      printf("  \033[36m%-25s\033[0m %s\n", \
	        substr($$0, 4, i - 4), substr($$0, i + 2)); \
	    } else { line = substr($$0, 4); }; next; \
	  }; \
	  /^[a-zA-Z_0-9?-]+:/ { \
	    if (line) { \
	      target = substr($$1, 1, length($$1) - 1); \
	      if (i = index(line, " # ")) { \
	        target = target " " substr(line, 1, i - 1); \
	        line = substr(line, i + 3); \
	      }; \
	      printf("  \033[36m%-25s\033[0m %s\n", target, line); \
	      line = ""; \
	    }; \
	  };' | less -R;

#@ show actual makefile.
show:
	@$(MAKE) --no-builtin-rules --no-builtin-variables --print-data-base \
          --question --makefile=$(firstword $(MAKEFILE_LIST)) : 2>/dev/null;
#@ lists all available targets.
list:
	@$(MAKE) --no-builtin-rules --no-builtin-variables --print-data-base \
	  --question --makefile=$(firstword $(MAKEFILE_LIST)) : 2>/dev/null | \
	  if [ "$(RUNARGS)" != "raw" ]; then awk -v RS= -F: ' \
	    /(^|\n)# Files(\n|$$)/,/(^|\n)# Finished Make data base/ { \
	      if ($$1 !~ "^[#.]") { print $$1 } \
	    }' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'; \
	  else cat -; fi;

#@ updates version and prepare release of the software as library.
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

#@ releases a fixed version of the software as library.
release:
	@VERSION="$$(cat VERSION)" && \
	if [ -n "$${VERSION}" -a \
	     -z "$$(git tag -l "v$${VERSION}")" ]; then \
	  git gh-release "v$${VERSION}" && \
	  echo "Added release tag v$${VERSION} to repository!"; \
	fi; \

#@ install all software created by the project.
install: $(TARGETS_INSTALL)
#@ install all software created and used by the project.
install-all: $(TARGETS_INSTALL_ALL)
#@ install-*: install the matched software command or service.

# install go tools used by the project.
$(addprefix $(GOBIN)/,$(call go-pkg,cmd,$(TOOLS_GO))): $(GOBIN)/%:
	go install $(call go-pkg,normal,$(TOOLS_GO),^$*$$);
$(TARGETS_INSTALL_GO): install-%:
	go install $(call go-pkg,normal,$(TOOLS_GO),^$*$$);
# install npm tools used by the project.
$(addprefix $(NVM_BIN)/,$(TOOLS_NPM:-cli=)): $(NVM_BIN)/%: install-%
$(TARGETS_INSTALL_NPM): install-%:
	@if command -v npm &> /dev/null && ! command -v $*; then \
	  echo "npm install --global ^$*$$"; \
	  npm install --global $(filter $*-cli,$(TOOLS_NPM)); \
	fi;
#@ install software command or service created by the project.
$(TARGETS_INSTALL): install-%: $(BUILDDIR)/%; cp $< $(GOBIN)/$*;

#@ uninstall all software created by the project.
uninstall: $(TARGETS_UNINSTALL)
#@ uninstall all software created and used by the projct.
uninstall-all: $(TARGETS_UNINSTALL_ALL)
#@ uninstall-*: uninstall the matched software command or service.

# uninstall go tools used by the project.
$(TARGETS_UNINSTALL_GO): uninstall-%:
	@#PACKAGE=$(call go-pkg,strip,$(TOOLS_GO),^$*$$); \
	#go clean -i $${PACKAGE}; # This does not work as expected.
	rm -f $(GOBIN)/$*;
# uninstall npm based tools used by the project.
$(TARGETS_UNINSTALL_NPM): uninstall-%:
	@if command -v npm &> /dev/null; then \
	  echo "npm uninstall --global $*"; \
	  npm uninstall --global $(filter $*-cli,$(TOOLS_NPM)); \
	fi;
# uninstall codacy tools used by the project.
$(TARGETS_UNINSTALL_CODACY): uninstall-codacy-%:
	@VERSION="$(CODACY_$(call upper,$*)_VERSION)"; \
	rm -f "$(GOBIN)/codacy-$*-$${VERSION}";
# uninstall software command or service created by the project.
$(TARGETS_UNINSTALL): uninstall-%:; rm -f $(GOBIN)/$*;


## Init: targets to initialize tools and (re-)sources.

#@ initializes the project to prepare it for building.
init: $(TARGETS_INIT)
#@ initializes the pre-commit hook.
init-hooks: .git/hooks/pre-commit
.git/hooks/pre-commit:
	@echo -ne "#!/bin/sh\n$(MAKE) commit" >$@; chmod 755 $@;

#@ initializes the package dependencies.
init-packages:
	go build ./...;

#@ initializes the generated sources.
init-sources: $(MOCKS)
$(MOCKS): go.sum $(MOCK_SOURCES) | $(GOBIN)/mockgen $(GOBIN)/mock
	go generate "$(shell echo $(MOCK_TARGETS) | \
	  sed -E "s:.*$@=([^ ]*).*$$:\1:;")";

#@ initializes the codacy support - if enabled.
init-codacy: $(TARGETS_INIT_CODACY)
$(TARGETS_INIT_CODACY): init-%:
	@VERSION="$(CODACY_$(call upper,$*)_VERSION)"; \
	FILE="$(GOBIN)/codacy-$*-$${VERSION}"; \
	if [ ! -f $${FILE} ]; then \
	  BASE="https://github.com/codacy/codacy-$*/releases/download"; \
	  echo curl --silent --location --output $${FILE} \
	    $${BASE}/$${VERSION}/codacy-$*-$${VERSION}; \
	  curl --silent --location --output $${FILE} \
	    $${BASE}/$${VERSION}/codacy-$*-$${VERSION}; \
	  chmod 700 $${FILE}; \
	fi; \


## Test: targets to test the source code (and compiler environment).
TEST_ALL := $(BUILDDIR)/test-all.cover
TEST_UNIT := $(BUILDDIR)/test-unit.cover
TEST_BENCH := $(BUILDDIR)/test-bench.cover

#@ executes default test set.
test: $(TARGETS_TEST)
#@ <pkg>[/<test>] # executes all tests.
test-all: test-clean init-sources $(TEST_ALL)
#@ <pkg>[/<test>] # executes only unit tests.
test-unit: test-clean init-sources $(TEST_UNIT)
#@ <pkg>[/<test>] # executes benchmarks.
test-bench: test-clean init-sources $(TEST_BENCH)

# removes all coverage files of test and benchmarks.
test-clean:
	@if [ -f "$(TEST_ALL)" ]; then rm -vf $(TEST_ALL); fi; \
	if [ -f "$(TEST_UNIT)" ]; then rm -vf $(TEST_UNIT); fi; \
	if [ -f "$(TEST_BENCH)" ]; then rm -vf $(TEST_BENCH); fi; \

#@ starts the test coverage report.
test-cover:
	@FILE=$$(ls -Art "$(TEST_ALL)" "$(TEST_UNIT)" \
	  "$(TEST_BENCH)" 2>/dev/null); \
	go tool cover -html="$${FILE}"; \

#@ uploads the test coverage report to codacy.
test-upload:
	@FILE=$$(ls -Art "$(TEST_ALL)" "$(TEST_UNIT)" \
	  "$(TEST_BENCH)" 2>/dev/null); \
	COMMIT="$$(git log --max-count=1 --pretty=format:"%H")"; \
	SCRIPT="https://coverage.codacy.com/get.sh"; \
	if [ -n "$(CODACY_PROJECT_TOKEN)" ]; then \
	  bash <(curl --silent --location $${SCRIPT}) report \
	    --project-token $(CODACY_PROJECT_TOKEN) --commit-uuid $${COMMIT} \
	    --codacy-api-base-url $(CODACY_API_BASE_URL) --language go \
	    --force-coverage-parser go --coverage-reports "$${FILE}"; \
	elif [ -n "$(CODACY_API_TOKEN)" ]; then \
	  bash <(curl --silent --location $${SCRIPT}) report \
	    --organization-provider $(CODACY_PROVIDER) \
	    --username $(CODACY_USER) --project-name $(CODACY_PROJECT) \
	    --api-token $(CODACY_API_TOKEN) --commit-uuid $${COMMIT} \
	    --codacy-api-base-url $(CODACY_API_BASE_URL) --language go \
	    --force-coverage-parser go --coverage-reports "$${FILE}"; \
	fi; \

#@ tests whether project is using the latest go-version.
test-go:
	@ERROR="error: local go version is $(GOVERSION)"; \
	VERSION=$$(grep "^go [0-9.]*$$" go.mod | cut -f2 -d' '); \
	if [ "$(GOVERSION)" != "$${VERSION}" ]; then \
	  echo "$${ERROR}: 'go.mod' requires $${VERSION}!"; \
	  if [[ "$(GOVERSION)" < "$${VERSION}" ]]; then \
	    GOCHANGE="upgrade"; CHANGE="downgrade"; \
	  else \
	    GOCHANGE="downgrade"; CHANGE="upgrade"; \
	  fi; \
	  echo -e "\t$${GOCHANGE} your local go version to $${VERSION} or "; \
	  echo -e "\trun 'make update-go $${VERSION}' to adjust the project!"; \
	  exit -1; \
	fi; \
	VERSIONS="$$(if [ -f "$(FILE_DELIVERY)" ]; then \
	    grep -Eo "$(FILE_DELIVERY_REGEX)" "$(FILE_DELIVERY)" | \
	    grep -Eo "[0-9.]*" | sort -u; else echo "$${VERSION}"; fi)"; \
	for VERSION in $${VERSIONS}; do \
	  if [ "$(GOVERSION)" != "$${VERSION}" ]; then \
	    echo "$${ERROR}: $(FILE_DELIVERY) requires $${VERSION}!"; \
	    if [[ "$(GOVERSION)" < "$${VERSION}" ]]; then \
	      GOCHANGE="upgrade"; CHANGE="downgrade"; \
	    else \
	      GOCHANGE="downgrade"; CHANGE="upgrade"; \
	    fi; \
	    echo -e "\t$${GOCHANGE} your local go version to $${VERSION} or "; \
	    echo -e "\trun 'make update-go $${VERSION}' to adjust the project!"; \
	    exit -1; \
	  fi; \
	done; \

# process test arguments.
define testargs
  if [ -n "$(RUNARGS)" ]; then ARGS=($(RUNARGS)); \
    if [[ -f "$${ARGS[0]}" && ! -d "$${ARGS[0]}" && \
			 "$${ARGS[0]}" == *_test.go ]]; then \
	  find $$(dirname $(RUNARGS) | sort -u) \
		-maxdepth 1 -a -name "*.go" -a ! -name "*_test.go" \
		-o -name "common_test.go" -o -name "mock_*_test.go" | \
		sed "s|^|./|"; \
	  echo $(addprefix ./,$(RUNARGS)); \
	elif [[ -d "$${ARGS[0]}" && ! -f "$${ARGS[0]}" ]]; then \
	  echo $(addprefix ./,$(RUNARGS)); \
	elif [[ ! -f "$${ARGS[0]}" && ! -d "$${ARGS[0]}" ]]; then \
	  for ARG in $${ARGS[0]}; do \
		if [ -z "$${PACKAGES}" ]; then PACKAGES="$${ARG%/*}"; \
		else PACKAGES="$${PACKAGES}\n$${ARG%/*}"; fi; \
		if [ -z "$${TESTCASE}" ]; then TESTCASES="-run $${ARG##*/}"; \
		else TESTCASES="$${TESTCASES} -run $${ARG##*/}"; fi; \
	  done; \
	  echo -en "$${PACKAGES}" | sort -u | sed "s|^|./|"; \
	  echo "$${TESTCASES}"; \
	else
	  echo "warning: invalid test parameters [$${ARGS[@]}]";
	fi;\
  else echo "./..."; fi
endef

TESTFLAGS ?= -race -mod=readonly -count=1
TESTARGS ?= $(shell $(testargs))

# actuall targets for testing.
$(TEST_ALL): $(SOURCES) init-sources $(TEST_DEPS) | $(BUILDDIR)
	go test $(TESTFLAGS) -timeout $(TEST_TIMEOUT) \
	  -cover -coverprofile $@ $(TESTARGS);
$(TEST_UNIT): $(SOURCES) init-sources | $(BUILDDIR)
	go test $(TESTFLAGS) -timeout $(TEST_TIMEOUT) \
	  -cover -coverprofile $@ -short $(TESTARGS);
$(TEST_BENCH): $(SOURCES) init-sources | $(BUILDDIR)
	go test $(TESTFLAGS) -benchtime=8s \
	  -cover -coverprofile $@ -short -bench=. $(TESTARGS);

#@ executes a kind of self-test running the main targets.
test-self: clean-all all-clean clean-all
	@echo "Self test finished successfully!!!";

# Variables and definitions for linting of source code.
COMMA := ,
SPACE := $(null) #

# disabled (deprecated): deadcode golint interfacer ifshort maligned musttag
#   nosnakecase rowserrcheck scopelint structcheck varcheck wastedassign
# diabled (distructive): nlreturn ireturn nonamedreturns varnamelen exhaustruct
#   exhaustivestruct gochecknoglobals gochecknoinits tagliatelle
# disabled (conflicting): godox gci paralleltest
# not listed (unnecessary): forcetypeassert wsl

LINTERS_DISCOURAGED ?= \
	deadcode exhaustivestruct exhaustruct gci gochecknoglobals gochecknoinits \
	godox golint ifshort interfacer ireturn maligned musttag nlreturn \
	nonamedreturns nosnakecase paralleltest rowserrcheck scopelint structcheck \
	tagliatelle varcheck varnamelen wastedassign
LINTERS_MINIMUM ?= \
	asasalint asciicheck bidichk bodyclose dogsled dupl dupword durationcheck \
	errchkjson execinquery exportloopref funlen gocognit goconst gocyclo \
	godot gofmt gofumpt goimports gosimple govet importas ineffassign maintidx \
	makezero misspell nestif nilerr prealloc predeclared promlinter reassign \
	sqlclosecheck staticcheck typecheck unconvert unparam unused \
	usestdlibvars whitespace
LINTERS_BASELINE ?= \
	containedctx contextcheck cyclop decorder depguard errname forbidigo \
	ginkgolinter gocheckcompilerdirectives goheader gomodguard \
	goprintffuncname gosec grouper interfacebloat lll loggercheck nakedret \
	nilnil noctx nolintlint nosprintfhostport
LINTERS_EXPERT ?= \
	errcheck errorlint exhaustive gocritic goerr113 gomnd gomoddirectives \
	revive stylecheck tenv testableexamples testpackage thelper tparallel \
	wrapcheck

LINT_FILTER := $(LINTERS_CUSTOM) $(LINTERS_DISABLED)
LINT_DISABLED ?= $(subst $(SPACE),$(COMMA),$(strip \
	$(filter-out $(LINT_FILTER),$(LINTERS_DISCOURAGED)) $(LINTERS_DISABLED)))
LINT_MINIMUM ?= $(subst $(SPACE),$(COMMA),$(strip \
	$(filter-out $(LINT_FILTER),$(LINTERS_MINIMUM)) $(LINTERS_CUSTOM)))
LINT_BASELINE ?= $(subst $(SPACE),$(COMMA),$(strip $(LINT_MINIMUM) \
	$(filter-out $(LINT_FILTER),$(LINTERS_BASELINE))))
LINT_EXPERT ?= $(subst $(SPACE),$(COMMA),$(strip $(LINT_BASELINE) \
	$(filter-out $(LINT_FILTER),$(LINTERS_EXPERT))))

ifeq ($(shell ls $(FILE_GOLANGCI) 2>/dev/null), $(FILE_GOLANGCI))
  LINT_CONFIG := --config $(FILE_GOLANGCI)
endif

LINT_MIN := --enable $(LINT_MINIMUM) --disable $(LINT_DISABLED)
LINT_BASE := --enable $(LINT_BASELINE) --disable $(LINT_DISABLED)
LINT_PLUS := --enable $(LINT_EXPERT) --disable $(LINT_DISABLED)
LINT_MAX := --enable-all --disable $(LINT_DISABLED)
LINT_ALL := --enable $(LINT_EXPERT),$(LINT_DISABLED) --disable-all

LINT_FLAGS ?= --allow-serial-runners --sort-results --color always
LINT_CMD ?= golangci-lint run $(LINT_CONFIG) $(LINT_FLAGS)
LINT_CMD_CONFIG := make --no-print-directory lint-config
ifeq ($(RUNARGS),linters)
  LINT_CMD := golangci-lint linters $(LINT_CONFIG) $(LINT_FLAGS)
else ifeq ($(RUNARGS),config)
  LINT_CMD := @
  LINT_MIN := LINT_ENABLED=$(LINT_MINIMUM) \
    LINT_DISABLED=$(LINT_DISABLED) $(LINT_CMD_CONFIG)
  LINT_BASE := LINT_ENABLED=$(LINT_BASELINE) \
    LINT_DISABLED=$(LINT_DISABLED) $(LINT_CMD_CONFIG)
  LINT_PLUS := LINT_ENABLED=$(LINT_EXPERT) \
    LINT_DISABLED=$(LINT_DISABLED) $(LINT_CMD_CONFIG)
  LINT_MAX := LINT_DISABLED=$(LINT_DISABLED) $(LINT_CMD_CONFIG)
  LINT_ALL := LINT_ENABLED=$(LINT_EXPERT),$(LINT_DISABLED) $(LINT_CMD_CONFIG)
else ifeq ($(RUNARGS),fix)
  LINT_CMD := golangci-lint run $(LINT_CONFIG) $(LINT_FLAGS) --fix
else ifneq ($(RUNARGS),)
  LINT_CMD := golangci-lint run $(LINT_CONFIG) $(LINT_FLAGS)
  LINT_MIN := --disable-all --enable $(RUNARGS)
  LINT_BASE := --disable-all --enable $(RUNARGS)
  LINT_PLUS := --disable-all --enable $(RUNARGS)
  LINT_MAX := --disable-all --enable $(RUNARGS)
  LINT_ALL := --disable-all --enable $(RUNARGS)
endif


## Lint: targets to lint source code.

#@ <fix|linter> # execute linters for custom code quality level.
lint: $(TARGETS_LINT)
#@ <fix|linter> # execute golangci linters for minimal code quality level.
lint-min: init-sources $(GOBIN)/golangci-lint; $(LINT_CMD) $(LINT_MIN)
#@ <fix|linter> # execute golangci linters for base code quality level.
lint-base: init-sources $(GOBIN)/golangci-lint;	$(LINT_CMD) $(LINT_BASE)
#@ <fix|linter> # execute golangci linters for plus code quality level.
lint-plus: init-sources $(GOBIN)/golangci-lint;	$(LINT_CMD) $(LINT_PLUS)
#@ <fix|linter> # execute golangci linters for maximal code quality level.
lint-max: init-sources $(GOBIN)/golangci-lint; $(LINT_CMD) $(LINT_MAX)
#@ <fix|[linter]> # execute all golangci linters for insane code quality level.
lint-all: init-sources $(GOBIN)/golangci-lint; $(LINT_CMD) $(LINT_ALL)

#@ creates a golangci linter config using the custom code quality level.
lint-config:
	@LINT_START=$$(awk '($$0 == "linters:"){print NR-1}' $(FILE_GOLANGCI)); \
	LINT_STOP=$$(awk '(start && $$0 ~ "^[^ ]+:$$"){print NR; start=0} \
	  ($$0 == "linters:"){start=1}' $(FILE_GOLANGCI)); \
	(sed -ne '1,'$${LINT_START}'p' $(FILE_GOLANGCI); \
	echo -e "linters:\n  enable:"; \
	for LINTER in "# List of min-set linters (min)" \
	    $(LINTERS_MINIMUM) \
		"# List of base-set linters (base)" $(LINTERS_BASELINE) \
		"# List of plus-set linters (plus)" $(LINTERS_EXPERT); do \
	  if [ "$${LINTER:0:1}" == "#" ]; then X=""; else X="- "; fi; \
	  echo -e "    $${X}$${LINTER}"; \
	done; echo -e "\n  disable:"; \
	for LINTER in "# List of to-avoid linters (avoid)" \
	    $(LINTERS_DISABLED); do \
	  if [ "$${LINTER:0:1}" == "#" ]; then X=""; else X="- "; fi; \
	  echo -e "    $${X}$${LINTER}"; \
	done; \
	echo; sed -ne $${LINT_STOP}',$$p' $(FILE_GOLANGCI)) | \
	awk -v enabled=$${LINT_ENABLED} -v disabled=$${LINT_DISABLED} ' \
	  ((start == 2) && ($$1 == "-")) { \
	    enabled = enable[$$2]; disabled = disable[$$2]; \
	    if (enabled && disabled) { \
	      print $$0 "  # (conflicting)" \
	    } else if (!enabled && disabled) { \
	      print gensub("-","# -", 1) "  # (disabled)" \
	    } else if (!enabled && !disabled && enone) { \
	      print gensub("-","# -", 1) "  # (missing)" \
	    } else { print $$0 } \
	    enable[$$2] = 2; next; \
	  } \
	  ((start == 3) && ($$1 == "-")) { \
	    enabled = enable[$$2]; disabled = disable[$$2]; \
	    if (enabled && disabled) { \
	      print gensub("-","# -", 1) "  # (conflicting)" \
	    } else if (enabled && !disabled) { \
	      print gensub("-","# -", 1) "  # (enabled)" \
	    } else if (!enabled && !disabled && dnone) { \
	      print gensub("-","# -", 1) "  # (missing)" \
	    } else { print $$0 } \
	    disable[$$2] = 2; next; \
	  } \
	  ((start != 0) && ($$0 ~ "^[^ ]+:$$")) { start=0 } \
	  ((start >= 2) && ($$0 ~ "^  [^ ]+:$$")) { none=1; \
	    if (start == 2) { for (key in enable) { \
	      if (key && enable[key] == 1) { \
	        if (none) { print "    # Additional linters" } \
	        print "    - " key; none=0 \
	      } \
	    } } else if (start == 3) { for (key in disable) { \
	      if (key && disable[key] == 1) { \
	        if (none) { print "    # Additional linters" } \
	        print "    - " key; none=0 \
	      } \
	    } } \
	  } \
	  ((start != 0) && ($$0 == "  disable:")) { start = 3 }\
	  ((start != 0) && ($$0 == "  enable:")) { start = 2 } \
	  ((start == 0) && ($$0 == "linters:")) { start = 1; \
	    enone = split(enabled, array, ","); \
		for (i in array){ enable[array[i]] = 1 } \
	    dnone = split(disabled, array, ","); \
		for (i in array){ disable[array[i]] = 1 } \
	  } { print $$0 }' \

#@ execute all codacy linters.
lint-codacy: lint-revive $(TARGETS_LINT_CODACY_CLIENTS) $(TARGETS_LINT_CODACY_BINARIES)
$(TARGETS_LINT_CODACY_CLIENTS): lint-%: init-sources
	@LARGS=("--allow-network" "--skip-uncommitted-files-check"); \
	LARGS+=("--codacy-api-base-url" "$(CODACY_API_BASE_URL)"); \
	if [ -n "$(CODACY_PROJECT_TOKEN)" ]; then \
	  LARGS+=("--project-token" "$(CODACY_PROJECT_TOKEN)"); \
	  LARGS+=("--upload" "--verbose"); \
	elif [ -n "$(CODACY_API_TOKEN)" ]; then \
	  LARGS+=("--provider" "$(CODACY_PROVIDER)"); \
	  LARGS+=("--username" "$(CODACY_USER)"); \
	  LARGS+=("--project" "$(CODACY_PROJECT)"); \
	  LARGS+=("--api-token" "$(CODACY_API_TOKEN)"); \
	  LARGS+=("--upload" "--verbose"); \
	fi; \
	echo """$(IMAGE_CMD) run --rm=true --env CODACY_CODE="/code" \
	  --volume /var/run/docker.sock:/var/run/docker.sock \
	  --volume /tmp:/tmp --volume ".":"/code" \
	  codacy/codacy-analysis-cli analyze "$${LARGS[@]}" --tool $*"""; \
	$(IMAGE_CMD) run --rm=true --env CODACY_CODE="/code" \
	  --volume /var/run/docker.sock:/var/run/docker.sock \
	  --volume /tmp:/tmp --volume "$$(pwd):/code" \
	  codacy/codacy-analysis-cli analyze "$${LARGS[@]}" --tool $* || \
	    $(CODACY_CONTINUE); \

LINT_ARGS_GOSEC_LOCAL := -log /dev/null -exclude G105,G307 ./...
LINT_ARGS_GOSEC_UPLOAD := -log /dev/null -exclude G105,G307 -fmt json ./...
LINT_ARGS_STATICCHECK_LOCAL := -tests ./...
LINT_ARGS_STATICCHECK_UPLOAD := -tests -f json ./...

$(TARGETS_LINT_CODACY_BINARIES): lint-%: init-sources init-% | $(GOBIN)/%
	@COMMIT=$$(git log --max-count=1 --pretty=format:"%H"); \
	VERSION="$(CODACY_$(call upper,$*)_VERSION)"; \
	LARGS=("--silent" "--location" "--request" "POST" \
	  "--header" "Content-type: application/json"); \
	if [ -n "$(CODACY_PROJECT_TOKEN)" ]; then \
	  BASE="$(CODACY_API_BASE_URL)/2.0/commit/$${COMMIT}"; \
	  LARGS+=("-H" "project-token: $(CODACY_PROJECT_TOKEN)"); \
	elif [ -n "$(CODACY_API_TOKEN)" ]; then \
	  SPATH="$(CODACY_PROVIDER)/$(CODACY_USER)/$(CODACY_PROJECT)"; \
	  BASE="$(CODACY_API_BASE_URL)/2.0/$${SPATH}/commit/$${COMMIT}"; \
	  LARGS+=("-H" "api-token: $(CODACY_API_TOKEN)"); \
	fi; \
	if [ -n "$${BASE}" ]; then \
	  echo -e """$(GOBIN)/$* $(LINT_ARGS_$(call upper,$*)_UPLOAD) | \
	    $(GOBIN)/codacy-$*-$${VERSION} | \
	    curl "$${LARGS[@]}" --data @- "$${BASE}/issuesRemoteResults"; \
	    curl "$${LARGS[@]}" "$${BASE}/resultsFinal""""; \
	  ( $(GOBIN)/$* $(LINT_ARGS_$(call upper,$*)_UPLOAD) | \
	    $(GOBIN)/codacy-$*-$${VERSION} | \
	    curl "$${LARGS[@]}" --data @- "$${BASE}/issuesRemoteResults"; echo; \
	    curl "$${LARGS[@]}" "$${BASE}/resultsFinal"; echo ) || \
	      $(CODACY_CONTINUE); \
	else \
	  echo $(GOBIN)/$* $(LINT_ARGS_$(call upper,$*)_LOCAL); \
	  $(GOBIN)/$* $(LINT_ARGS_$(call upper,$*)_LOCAL) || $(CODACY_CONTINUE); \
	fi; \

#@ execute revive linter.
lint-revive: init-sources | $(GOBIN)/revive
	revive -formatter friendly -config=revive.toml $(SOURCES) || \
	  $(CODACY_CONTINUE);

#@ execute markdown linter.
lint-markdown: init-sources $(NVM_BIN)/markdownlint
	@echo markdownlint --config .markdownlint.yaml .; \
	if command -v markdownlint &> /dev/null; then \
	  markdownlint --config .markdownlint.yaml .; \
	else $(IMAGE_CMD) run --tty --volume $$(pwd):/src:ro \
	  container-registry.zalando.net/library/node-18-alpine:latest \
	  /bin/sh -c "npm install --global markdownlint-cli >/dev/null 2>&1 && \
	    cd /src && markdownlint --config .markdownlint.yaml ."; \
	fi; \

#@ execute gitleaks linter to check committed code.
lint-leaks: | $(GOBIN)/gitleaks
	gitleaks detect --no-banner --verbose --source .;
#@ execute gitleaks linter to check un-committed code changes.
lint-leaks?: | $(GOBIN)/gitleaks
	gitleaks protect --no-banner --verbose --source .;
#@ execute zally api-linter.
lint-apis: | $(GOBIN)/zally
	@LINTER="https://infrastructure-api-linter.zalandoapis.com"; \
	if ! curl --silent $${LINTER}; then \
	  echo "warning: API linter not available;"; exit 0; \
	fi; \
	ARGS=("--linter-service" "$${LINTER}"); \
	if command -v ztoken > /dev/null; then ARGS+=("--token" "$$(ztoken)"); fi; \
	for APISPEC in $$(find zalando-apis -name "*.yaml" 2>/dev/null); do \
	  echo "check API: zally \"$${APISPEC}\""; \
	  zally "$${ARGS[@]}" lint "$${APISPEC}" || exit 1; \
	done;


# Variables for setting up container specific build flags.
BUILDOS ?= $(go env GOOS)
BUILDARCH ?= $(go env GOARCH)
IMAGEOS ?= ${shell grep "^FROM [^ ]*$$" $(FILE_CONTAINER) 2>/dev/null | \
	grep -v " as " | sed "s/.*\(alpine\|ubuntu\).*/\1/g"}
ifeq ($(IMAGEOS),alpine)
  BUILDFLAGS ?= -v -mod=readonly
  GOCGO := 0
else
  BUILDFLAGS ?= -v -race -mod=readonly
  GOCGO := 1
endif

# Define flags propagate versions to build commands.
LDFLAGS ?= -X $(shell go list ./... | grep "config$$").Version=$(IMAGE_VERSION) \
	-X $(shell go list ./... | grep "config$$").GitHash=$(shell git rev-parse --short HEAD) \
	-X main.Version=$(IMAGE_VERSION) -X main.GitHash=$(shell git rev-parse --short HEAD)


## Build: targets to build native and linux platform executables.

#@ build default executables (native).
build: build-native
#@ build native platform executables using system architecture.
build-native: $(TARGETS_BUILD)
$(TARGETS_BUILD): build-%: $(BUILDDIR)/%
$(BUILDDIR)/%: cmd/%/main.go $(SOURCES)
	@mkdir -p "$(dir $@)";
	GOOS=$(BUILDOS) GOARCH=$(BUILDARCH) CGO_ENABLED=1 go build \
	  $(BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $@ $<;

#@ build linux platform executables using default (system) architecture.
build-linux: $(TARGETS_BUILD_LINUX)
$(BUILDDIR)/linux/%: cmd/%/main.go $(SOURCES)
	@mkdir -p "$(dir $@)";
	GOOS=$(BUILDOS) GOARCH=$(BUILDARCH) CGO_ENABLED=$(GOCGO) go build \
	  $(BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $@ $<;

#@ build container image (alias for image-build).
build-image: image-build


## Image: targets to build and push container images.

#@ build and push container images - if setup.
image: $(if $(filter $(IMAGE_PUSH),never),,image-push)
#@ build container images.
image-build: $(TARGETS_IMAGE_BUILD)

define setup-image-file
  FILE="$*"; IMAGE=$(IMAGE); \
  PREFIX="$$(basename "$${FILE%$(FILE_CONTAINER)*}")"; \
  SUFFIX="$$(echo $${FILE#*$(FILE_CONTAINER)} | sed "s/^[^[:alnum:]]*//")"; \
  INFIX="$${SUFFIX:-$${PREFIX}}"; \
  if [ -n "$${INFIX}" ]; then \
    IMAGE="$${IMAGE/:/-$${INFIX}:}"; \
  fi;
endef

$(TARGETS_IMAGE_BUILD): image-build/%: build-linux
	@$(call setup-image-file $*) \
	if [ "$(IMAGE_PUSH)" == "never" ]; then \
	  echo "We never build images, aborting [$${IMAGE}]."; exit 0; \
	fi; \
	REVISION=$$(git rev-parse HEAD 2>/dev/null); \
	BASE=$$(grep "^FROM[[:space:]]" $${FILE} | tail -n1 | cut -d" " -f2); \
	$(IMAGE_CMD) build --tag $${IMAGE} --file=$${FILE} . \
	  --label=org.opencontainers.image.created="$$(date --rfc-3339=seconds)" \
	  --label=org.opencontainers.image.vendor=zalando/$(TEAM) \
	  --label=org.opencontainers.image.source=$(REPOSITORY) \
	  --label=org.opencontainers.image.version=$${IMAGE##*:} \
	  --label=org.opencontainers.image.revision=$${REVISION} \
	  --label=org.opencontainers.image.base.name=$${BASE} \
	  --label=org.opencontainers.image.ref.name=$${IMAGE}; \

#@ push contianer images - if setup and allowed.
image-push: $(TARGETS_IMAGE_PUSH)
$(TARGETS_IMAGE_PUSH): image-push/%: image-build/%
	@$(call setup-image-file $*) \
	if [ "$(IMAGE_PUSH)" == "never" ]; then \
	  echo "We never push images, aborting [$${IMAGE}]."; exit 0; \
	elif [ "$(IMAGE_VERSION)" == "snapshot" ]; then \
	  echo "We never push snapshot images, aborting [$${IMAGE}]."; exit 0; \
	elif [ -n "$(CDP_PULL_REQUEST_NUMBER)" -a "$(IMAGE_PUSH)" != "pulls" ]; then \
	  echo "We never push pull request images, aborting [$${IMAGE}]."; exit 0; \
	fi; \
	$(IMAGE_CMD) push $${IMAGE}; \


## Run: targets for starting services and commands for testing.
HOST := "127.0.0.1"

#@ starts a postgres database instance for testing.
run-db: $(RUNDIR)
	@if [[ ! "$(TEST_DEPS) $(RUN_DEPS)" =~ run-db ]]; then exit 0; fi; \
	echo "info: ensure $(DB_IMAGE) running on $(HOST):$(DB_PORT)"; \
	if [ -n "$$($(IMAGE_CMD) ps | grep "$(DB_IMAGE).*$(HOST):$(DB_PORT)")" ]; then \
	  echo "info: port allocated, use existing db container!"; exit 0; \
	fi; \
	$(IMAGE_CMD) start ${IMAGE_ARTIFACT}-db 2>/dev/null || ( \
	$(IMAGE_CMD) run --detach --tty \
	  --name ${IMAGE_ARTIFACT}-db \
	  --publish $(HOST):$(DB_PORT):5432 \
	  --env POSTGRES_USER="$(DB_USER)" \
	  --env POSTGRES_PASSWORD="$(DB_PASSWORD)" \
	  --env POSTGRES_DB="$(DB_NAME)" $(DB_IMAGE) \
	    -c 'shared_preload_libraries=pg_stat_statements' \
        -c 'pg_stat_statements.max=10000' \
        -c 'pg_stat_statements.track=all' \
	  $(RUNARGS) 2>&1 & \
	until [ "$$($(IMAGE_CMD) inspect --format {{.State.Running}} \
	         $(IMAGE_ARTIFACT)-db 2>/dev/null)" == "true" ]; \
	do echo "waiting for db container" >/dev/stderr; sleep 1; done && \
	until $(IMAGE_CMD) exec $(IMAGE_ARTIFACT)-db \
	  pg_isready -h localhost -U $(DB_USER) -d $(DB_NAME); \
	do echo "waiting for db service" >/dev/stderr; sleep 1; done) |\
	tee -a $(RUNDIR)/$(IMAGE_ARTIFACT)-db; \

#@ starts an AWS localstack instance for testing.
run-aws: $(RUNDIR)
	@if [[ ! "$(TEST_DEPS) $(RUN_DEPS)" =~ run-aws ]]; then exit 0; fi; \
	echo "info: ensure $(AWS_IMAGE) is running on $(HOST):4566/4571" && \
	if [ -n "$$($(IMAGE_CMD) ps | \
	    grep "$(AWS_IMAGE).*$(HOST):4566.*$(HOST):4571")" ]; then \
	  echo "info: ports allocated, use existing aws container!"; \
	  $(call run-aws-setup); exit 0; \
	fi; \
	$(IMAGE_CMD) start ${IMAGE_ARTIFACT}-aws 2>/dev/null || ( \
	$(IMAGE_CMD) run --detach --tty --name ${IMAGE_ARTIFACT}-aws \
	  --publish $(HOST):4566:4566 --publish $(HOST):4571:4571 \
	  --env SERVICES="$(AWS_SERVICES)" $(AWS_IMAGE) $(RUNARGS) 2>&1 && \
	until [ "$$($(IMAGE_CMD) inspect --format {{.State.Running}} \
	         $(IMAGE_ARTIFACT)-aws 2>/dev/null)" == "true" ]; \
	do echo "waiting for aws container" >/dev/stderr; sleep 1; done && \
	until $(IMAGE_CMD) exec $(IMAGE_ARTIFACT)-aws \
	      curl --silent http://$(HOST):4566; \
	do echo "waiting for aws service" >/dev/stderr; sleep 1; done && \
	$(call run-aws-setup)) | \
	tee -a $(RUNDIR)/$(IMAGE_ARTIFACT)-aws.log; \

#@ run-*: starts the provide command using the native binary.
$(TARGETS_RUN): run-%: $(BUILDDIR)/% $(RUN_DEPS) $(RUNDIR) $(CREDDIR)
	@$(call run-setup); $(call run-vars) $(call run-vars-local) \
	  $(BUILDDIR)/$* $(RUNARGS) 2>&1 | \
	  tee -a $(RUNDIR)/$(IMAGE_ARTIFACT)-$*.log; \
	exit $${PIPESTATUS[0]};

#@ run-go-*: starts the provide command using go run.
$(TARGETS_RUN_GO): run-go-%: $(BUILDDIR)/% $(RUN_DEPS) $(RUNDIR) $(CREDDIR)
	@$(call run-setup); $(call run-vars) $(call run-vars-local) \
	  go run cmd/$*/main.go $(RUNARGS) 2>&1 | \
	  tee -a $(RUNDIR)/$(IMAGE_ARTIFACT)-$*.log; \
	exit $${PIPESTATUS[0]};

#@ run-image-*: starts the matched command via the container images.
$(TARGETS_RUN_IMAGE): run-image-%: $(RUN_DEPS) $(RUNDIR) $(CREDDIR)
	@trap "$(IMAGE_CMD) rm $(IMAGE_ARTIFACT)-$* >/dev/null" EXIT; \
	trap "$(IMAGE_CMD) kill $(IMAGE_ARTIFACT)-$* >/dev/null" INT TERM; \
	if [ -n "$(filter %$*,$(TARGETS_IMAGE))" ]; then \
	  IMAGE="$$(echo "$(IMAGE)" | sed -e "s/:/-$*:/")"; \
	else IMAGE="$(IMAGE)" fi; $(call run-setup); \
	$(IMAGE_CMD) run --name $(IMAGE_ARTIFACT)-$* --network=host \
	  --volume $(CREDDIR):/meta/credentials --volume $(RUNDIR)/temp:/tmp \
	  $(call run-vars, --env) $(call run-vars-image, --env) \
	  ${IMAGE} /$* $(RUNARGS) 2>&1 | \
	tee -a $(RUNDIR)/$(IMAGE_ARTIFACT)-$*.log; \
	exit $${PIPESTATUS[0]};

#@ clean up all running container images.
run-clean: $(TARGETS_RUN_CLEAN)
#@ run-clean-*: kills and removes the container image of the matched command.
$(TARGETS_RUN_CLEAN): run-clean-%:
	@echo "check container $(IMAGE_ARTIFACT)-$*"; \
	if [ -n "$$($(IMAGE_CMD) ps | grep "$(IMAGE_ARTIFACT)-$*")" ]; then \
	  $(IMAGE_CMD) kill $(IMAGE_ARTIFACT)-$* > /dev/null && \
	  echo "killed container $(IMAGE_ARTIFACT)-$*"; \
	fi; \
	if [ -n "$$($(IMAGE_CMD) ps -a | grep "$(IMAGE_ARTIFACT)-$*")" ]; then \
	  $(IMAGE_CMD) rm $(IMAGE_ARTIFACT)-$* > /dev/null && \
	  echo "removed container $(IMAGE_ARTIFACT)-$*"; \
	fi; \


## Cleanup: targets to clean up sources, tools, and containers.

#@ clean up resources created during build processes.
clean: $(TARGETS_CLEAN)
#@ clean up all resources, i.e. also tools installed for the build.
clean-all: $(TARGETS_CLEAN_ALL)
#@ clean up all resources created by initialization.
clean-init:; rm -vrf .git/hooks/pre-commit;
#@ clean up all resources created by building and testing.
clean-build:; rm -vrf $(BUILDDIR) $(RUNDIR);
	find . -name "mock_*_test.go" -exec rm -v {} \;;
#@ clean up all running container images.
clean-run: $(TARGETS_CLEAN_RUN)
#@ clean-run-*: clean up matched running container image.
$(TARGETS_CLEAN_RUN): clean-run-%: run-clean-%


## Update: targets to update tools, configs, and dependencies.

#@ update default or customized update targets.
update: $(TARGETS_UPDATE)
#@ check default or customized update targets for updates.
update?: $(TARGETS_UPDATE?)
#@ update all components.
update-all: $(TARGETS_UPDATE_ALL)
#@ check all components for udpdates.
update-all?: $(TARGETS_UPDATE_ALL?)

#@ <version> # update go to latest or given version.
update-go:
	@if ! [[ "$(RUNARGS)" =~ "[0-9.]*" ]]; then \
	  ARCH="linux-amd64"; BASE="https://go.dev/dl"; \
	  VERSION="$$(curl --silent --header "Accept-Encoding: gzip" \
	    "$${BASE}/" | gunzip | grep -o "go[0-9.]*$${ARCH}.tar.gz" | \
	      head -n 1 | sed "s/go\(.*\)\.[0-9]*\.$${ARCH}\.tar\.gz/\1/")"; \
	else VERSION="$(RUNARGS)"; fi; \
	echo "info: update golang version to $${VERSION}"; \
	go mod tidy -go=$${VERSION}; \
	if [ -f $(FILE_DELIVERY) ]; then \
	  sed -E -i -e "s/$(FILE_DELIVERY_REGEX)/\1$${VERSION}/" \
	    $(FILE_DELIVERY); \
	fi; \
	if [ "$(GOVERSION)" != "$${VERSION}" ]; then \
	  echo "warning: current compiler is using $(GOVERSION)" >/dev/stderr; \
	fi; \

#@ check whether a new go version exists.
update-go?:
	@ARCH="linux-amd64"; BASE="https://go.dev/dl"; \
	VERSION="$$(curl --silent --header "Accept-Encoding: gzip" \
	  "$${BASE}/" | gunzip | grep -o "go[0-9.]*$${ARCH}.tar.gz" | \
	    head -n 1 | sed "s/go\(.*\)\.[0-9]*\.$${ARCH}\.tar\.gz/\1/")"; \
	if [ "$(GOVERSION)" != "$${VERSION}" ]; then \
	  echo "info: new golang version $${VERSION} available"; \
	  echo -e "\trun 'make update-go' to update the project!"; \
	  exit -1; \
	fi; \

#@ [major] # updates minor (and major) dependencies to latest versions.
update-deps: test-go update/go.mod | $(GOBIN)/gomajor
update/go.mod:
	@if [ "$(RUNARGS)" == "major" ]; then \
	  readarray -t UPDATES < <(gomajor list -major 2>&1); \
	  for UPDATE in "$${UPDATES[@]}"; do ARGS=($${UPDATE}); \
	    if [ "$${ARGS[1]}" == "failed:" ]; then continue; fi; \
		SVERSION="$${ARGS[1]%%.*}"; TVERSION="$${ARGS[3]%%.*}"; \
		if [ "$${SVERSION}" == "$${TVERSION}" ]; then continue; fi; \
	    SPACKAGE="$${ARGS[0]%%:*}"; \
		if [[ "$${SPACKAGE}" == *$${SVERSION}* ]]; then \
		  TPACKAGE="$${SPACKAGE/$${SVERSION}/$${TVERSION}}"; \
		else TPACKAGE="$${SPACKAGE}/$${TVERSION}"; fi; \
		echo "update: $${SPACKAGE} => $${TPACKAGE}"; \
		sed -i -e "s#$${SPACKAGE}#$${TPACKAGE}#g" $(SOURCES); \
	  done; \
	  readarray -t UPDATES < <(gomajor list 2>&1); \
	  for UPDATE in "$${UPDATES[@]}"; do ARGS=($${UPDATE}); \
	    if [ "$${ARGS[1]}" == "failed:" ]; then continue; fi; \
	    PACKAGE="$${ARGS[0]%%:*}"; \
	    SVERSION="$${ARGS[1]}"; TVERSION="$${ARGS[3]%%]}"; \
		echo "update: $${PACKAGE} $${SVERSION} => $${TVERSION}"; \
		go get "$${PACKAGE}@$${TVERSION}"; \
	  done; \
	fi; \
	ROOT=$$(pwd); \
	for DIR in $$(find . -name "*.go" | xargs dirname | sort -u); do \
	  echo "update: $${DIR}"; cd $${ROOT}/$${DIR##./} && \
	  go mod tidy -x -v -e -compat=${GOVERSION} && go get -u || exit -1; \
	done; \

#@ checks for major and minor updates to latest versions.
update-deps?: test-go update/go.mod? | $(GOBIN)/gomajor
update/go.mod?:
	gomajor list .;

#@ update this build environment to latest version.
update-make: $(TARGETS_UPDATE_MAKE)
$(TARGETS_UPDATE_MAKE): update/%: update-base
	@DIR="$$(pwd)"; cd $(BASEDIR); \
	if [ ! -e "$${DIR}/$*" ]; then touch "$${DIR}/$*"; fi; \
	DIFF="$$(diff <(git show HEAD:$* 2>/dev/null) $${DIR}/$*)"; \
	if [ -n "$${DIFF}" ]; then \
	  if [ -n "$$(cd $${DIR}; git diff $*)" ]; then \
	    echo "info: $* is blocked (has been changed)"; \
	  else echo "info: $* is updated"; \
	    git show HEAD:$* > $${DIR}/$* 2>/dev/null; \
	  fi; \
	fi; \

#@ check whether updates for build environment exist.
update-make?: $(TARGETS_UPDATE_MAKE?)
$(TARGETS_UPDATE_MAKE?): update/%?: update-base
	@DIR="$$(pwd)"; cd $(BASEDIR); \
	if [ ! -e "$${DIR}/$*" ]; then \
	  echo "info: $* is not tracked"; \
	else DIFF="$$(diff <(git show HEAD:$*) $${DIR}/$*)"; \
	  if [ -n "$${DIFF}" ]; then \
	    if [ -n "$$(cd $${DIR}; git diff $*)" ]; then \
	      echo "info: $* is blocked (has been changed)"; \
	    else echo "info: $* needs update"; fi; \
	  fi; \
	fi; \

#@ creates a clone of the base repository to update from.
BASE := git@github.bus.zalan.do:builder-knowledge/go-base.git
update-base:
	@$(eval BASEDIR = $(shell mktemp -d)) ( \
	  trap "rm -rf $${DIR}" INT TERM EXIT; \
	  while pgrep $(MAKE) > /dev/null; do sleep 1; done; \
	  rm -rf $${DIR} \
	) & \
	git clone --no-checkout --depth 1 $(BASE) $(BASEDIR) 2> /dev/null; \

#@ updates all tools required by the project.
update-tools: $(TARGETS_INSTALL_GO) $(TARGETS_INSTALL_NPM)
	go mod tidy -compat=${GOVERSION};


# Include custom targets to extend scripts.
ifneq ("$(wildcard Makefile.targets)","")
  include Makefile.targets
endif
