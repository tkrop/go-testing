<!-- do not change!!! updated by 'make update-make' if not source of truth -->

# Makefile

This repository provides a generic [Makefile](Makefile) that is majorly working
by the following conventions:

1. Place all available commands under `cmd/<name>/main.go`.
2. Use a single 'config' package to read the configuration for all commands.
3. Use a common 'Dockerfile' to install all commands in a container image.

All targets in the [Makefile](Makefile) are designated to work out-of-the-box
taking care to setup the project, installing the necessary tools (except for the
golang compiler), and triggering the precondition targets as far as required.

The [Makefile](Makefile) supports targets to `test`, `lint`, and `build` the
source code to create commands (services), `install` the commands local or to a
container (`image-build`) and `run` them using with a customized setup. Targets
for running and testing allow to add arguments, e.g. `make run/test-* [args]`.

**Example:** `make test-unit base` runs all the unit tests in the package
`base`.

**Warning:** The [Makefile](Makefile) installs a `pre-commit` hook overwriting
and deleting any pre-existing hook, i.e. `make commit`, that enforces a basic
+unit testing and linting to run successfully before allowing to commit, i.e.
`test-go`, `test-unit`, `lint-base` (or what code quality level is defined),
and `lint-markdown`.


## Setup and customization

The [Makefile](Makefile) is using sensitive defaults that are supposed to work
out-of-the-box for most targets. Please see documentation of the target groups
for more information on setup and customization:

* [Standard targets](#standard-targets)
* [Test targets](#test-targets)
* [Linter targets](#linter-targets)
* [Install targets](#install-targets)
* [Delete targets](#delete-targets)
* [Image targets](#image-targets)
* [Update targets](#update-targets)
* [Cleanup targets](#cleanup-targets)
* [Init targets](#init-targets)
* [Release targets](#release-targets)

To customize the behavior of the Makefile there exist multiple extension points
that can be used to setup additional variables, definitions, and targets that
modify the behavior of the [Makefile](Makefile).

* [Makefile.vars](Makefile.vars) allows to modify the behavior of standard
  targets by customizing and defining additional variables (see section
  [Modifying variables](#modifying-variables) for more details).
* [Makefile.defs](Makefile.defs) allows to customize the runtime environment
  for executing of commands (see Section [Running commands](#running-commands)
  for more details).
* [Makefile.targets](Makefile.targets) is an optional extension point that
  allows to define arbitrary custom targets.


### Modifying variables

While there exist sensible defaults for all configurations variables, some of
them might need to be adjusted. The following list provides an overview of the
most prominent ones

```Makefile
# Setup code quality level (default: base).
CODE_QUALITY := plus
# Setup codacy integration (default: enabled [enabled, disabled]).
CODACY := enabled

# Setup required targets before testing (default: <empty>).
TEST_DEPS := run-db
# Setup required targets before running commands (default: <empty>).
RUN_DEPS := run-db
# Setup required aws services for testing (default: <empty>).
AWS_SERVICES :=

# Setup when to push images (default: pulls [never, pulls, merges])
IMAGE_PUSH ?= never

# Setup default test timeout (default: 10s).
TEST_TIMEOUT := 15s

# Setup custom delivery file (default: delivery.yaml).
FILE_DELIVERY := delivery-template.yaml
# Setup custom local build targets (default: init test lint build).
TARGETS_ALL := init delivery test lint build

# Custom linters applied to prepare next level (default: <empty>).
LINTERS_CUSTOM := nonamedreturns gochecknoinits tagliatelle
```

You can easily lookup a list using `grep -r " ?= " Makefile`, however, most
will not be officially supported unless mentioned in the above list.


### Running commands

To `run-*` commands as expected, you need to setup the environment variables
for your designated runtime by defining the custom functions for setting it up
via `run-setup`, `run-vars`, `run-vars-local`, and `run-vars-image` in
[Makefile.defs](Makefile.defs).

While tests are supposed to run with global defaults and test specific config,
the setup of the `run-*` commands strongly depends on the commands execution
context and its purpose. Still, there are common patterns that can be copied
from other commands and projects.

To enable postgres database support you must add `run-db` to `TEST_DEPS` and
`RUN_DEPS` variables to [Makefile.vars](Makefile.vars).

You can also override the default setup via the `DB_HOST`, `DB_PORT`, `DB_NAME`,
`DB_USER`, and `DB_PASSWORD` variables, but this is optional.

**Note:** when running test against a DB you usually have to extend the default
`TEST_TIMEOUT` of 10s to a less aggressive value.

To enable AWS localstack you have to add `run-aws` to the default`TEST_DEPS` and
`RUN_DEPS` variables, as well as to add your list of required aws services to
the `AWS_SERVICES` variable.

```Makefile
# Setup required targets before testing (default: <empty>).
TEST_DEPS := run-aws
# Setup required targets before running commands (default: <empty>).
RUN_DEPS := run-aws
# Setup required aws services for testing (default: <empty>).
AWS_SERVICES := s3 sns
```

**Note:** Currently, the [Makefile](Makefile) does not support all command-line
arguments since make swallows arguments starting with `-`. To compensate this
shortcoming the commands need to support setpu via command specific environment
variables following the principles of the [Twelf Factor App][12factor].

[12factor]: https://12factor.net/ "Twelf Factor App"


## Standard targets

The [Makefile](Makefile) supports the following often used standard targets.

```bash
make all     # short cut target to init, test, and build binaries locally
make cdp     # short cut target to init, test, and build containers in pipeline
make commit  # short cut target to execute pr-commit test and lint steps
make init    # short cut target to setup the project installing the latest tools
make test    # short cut target to generates sources to execute tests
make lint    # short cut target to generates and lints sources
```

The short cut targets can be customized by setting up the variables `TARGETS_*`
(in upper letters), according to your preferences in `Makefile.vars` or in your
+environment.

Other less customizable commands are targets to build, install, delete, and
cleanup project resources:

```bash
make build   # creates binary files of commands
make install # installs binary files of commands in '${GOPATH}/bin'
make delete  # deletes binary files of commands from '${GOPATH}/bin'
make clean   # cleans up the project removing all created files
```

While these targets allow to execute the most important tasks out-of-the-box,
there exist a high number of specialized (sometimes project specific) commands
that provide more features with quicker response times for building, testing,
releasing, and executing of components.

**Note:** All targets automatically trigger their preconditions and install the
latest version of the required tools, if some are missing. To enforce the setup
of a new tool, you need to run `make init` explicitly.


### Test targets

Often it is more efficient or even necessary to execute the fine grained test
targets to complete a task.

```bash
make test        # short cut for default test targets
make test-all    # executes the complete tests suite
make test-unit   # executes only unit tests by setting the short flag
make test-cover  # opens the test coverage report in the browser
make test-upload # uploads the test coverage files
make test-clean  # cleans up the test files
make test-go     # test go versions
```

In addition, it is possible to restrict test target execution to packages,
files and test cases as follows:

* For a single package use `make test-(unit|all) <package> ...`.
* For a single test file `make test[-(unit|all) <package>/<file>_test.go ...`.
* For a single test case `make test[-(unit|all) <package>/<test-name> ...`.

The default test target can be customized by defining the `TARGETS_TEST`
variable in `Makefile.vars`. Usually this is not necessary.


### Linter targets

The [Makefile](Makefile) supports different targets that help with linting
according to different quality levels, i.e. `min`,`base` (default), `plus`,
`max`, (and `all`) as well as automatically fixing the issues.

```bash
make lint          # short cut to execute the default lint targets
make lint-min      # lints the go-code using a minimal config
make lint-base     # lints the go-code using a baseline config
make lint-plus     # lints the go-code using an advanced config
make lint-max      # lints the go-code using an expert config
make lint-all      # lints the go-code using an insane all-in config
make lint-codacy   # lints the go-code using codacy client side tools
make lint-markdown # lints the documentation using markdownlint
make lint-api      # lints the api specifications in '/zalando-apis'
```

The default target for `make lint` is determined by the selected `CODE_QUALITY`
level (`min`, `base`, `plus`, and `max`), and the `CODACY` setup (`enabled`,
`disabled`). The default setup is to run the targets `lint-base`, `lint-apis`,
`lint-markdown`, and `lint-codacy`. It can be further customized via changing
the `TARGETS_LINT` in `Makefile.vars` - if necessary.

The `lint-*` targets for `golangci-lint` allow some command line arguments:

1. The keyword `fix` to lint with auto fixing enabled (when supported),
2. The keyword `config` to shows the effective linter configuration,
3. The keyword `linters` to display the linters with description, or
4. `<linter>,...` comma separated list of linters to enable for a quick checks.

The default linter config is providing a golden path with different levels
out-of-the-box, i.e. a `min` for legacy code, `base` as standard for active
projects, and `plus` for experts and new projects, and `max` enabling all
but the conflicting disabled linters. Besides, there is an `all` level that
allows to experience the full linting capability.

Independen of the golden path this setting provides, the lint expert levels
can be customized in three ways.

1. The default way is to add additional linters for any level by setting the
  `LINTERS_CUSTOM` variable adding a white space separated list of linters.
2. Less comfortable and a bit trickier is the approach to override the linter
  config variables `LINTERS_DISABLED`, `LINTERS_DEFAULT`, `LINTERS_MINIMUM`,
  `LINTERS_BASELINE`, and `LINTERS_EXPERT`, to change the standards.
3. Last the linter configs can be changed via `.golangci.yaml`, as well as
  via `.codacy.yaml`, `.markdownlint.yaml`, and `revive.toml`.

However, customizing `.golangci.yaml` and other config files is currently not
advised, since the `Makefile` is designed to update and enforce a common
version on running `update-*` targets.


### Install targets

The install targets installs the latest build version of a command in the
`${GOPATH}/bin` directory for simple command line execution.

```bash
make install      # installs all commands using the platform binaries
make install-(*)  # installs the matched command using the platform binary
```

If a command, service, job has not been build before, it is first build.

**Note:** Please use carefully, if your project uses common command names.


### Delete targets

The delete targets delete the latest installed command from `${GOPATH}/bin`.

```bash
make delete      # Deletes all commands
make delete-(*)  # Deletes the matched command
```

**Note:** Please use carefully, if your project uses common command names.


### Image targets

Based on the convention that all binaries are installed in a single container
image, the [Makefile](Makefile) supports to create and push the container image
as required for a pipeline.

```bash
make image        # short cut for 'image-build'
make image-build  # build a container image after building the commands
make image-push   # pushes a container image after building it
```

The targets are checking silently whether there is an image at all, and whether
it should be build and pushed according to the pipeline setup. You can control
this behavior by setting `IMAGE_PUSH` to `never` or `test` to disable pushing
(and building) or enable it in addition for pull requests. Any other value will
ensure that images are only pushed for `main`-branch and local builds.


### Run targets

The [Makefile](Makefile) supports targets to startup a common DB and a common
AWS container image as well as to run the commands provided by the repository.

```bash
make run-db     # runs a postgres container image to provide a DBMS
make run-aws    # runs a localstack container image to simulate AWS
make run-(*)    # runs the matched command using its before build binary
make run-go-(*) # runs the matched command using 'go run'
make run-image-(*) # runs the matched command in the container image
```

To run commands successfully the environment needs to be setup to run the
commands in its runtim. Please visit [Running commands](#running-commands) for
more details on how to do this.

**Note:** The DB (postgres) and AWS (localstack) containers can be used to
support any number of parallel applications, if they use different tables,
queues, and buckets. Developers are encouraged to continue with this approach
and only switch application ports and setups manually when necessary.


### Update targets

The [Makefile](Makefile) supports targets for common maintainance tasks.

```bash
make update        # short cut to execute update-deps
make update-go     # updates go version to reflect the current compiler
make update-deps   # updates the project dependencies to the latest version
make update-tools  # updates the project tools to the latest versions
make update-make   # updates the Makefile to the latest version
make update-codacy # updates the codacy configs to the latest versions
```

It is advised to use and extend this targets when necessary.


### Cleanup targets

The [Makefile](Makefile) is designed to clean up everything it has created by
executing the following targets.

```bash
make clean         # short cut for clean-init, clean-build, and clean-test
make clean-init    # cleans up all resources created by the init targets
make clean-build   # cleans up all resources created by the build targets
make clean-test    # cleans up all resources created for the test targets
make clean-run(-*) # cleans up all resources created for the run targets
```


### Init targets

The [Makefile](Makefile) supports initialization targets that are usually
already added as perquisites for targets that need them. So there is usually
no need to call them directly.


```bash
make init           # short cut for 'init-tools init-hooks init-packages'
make init-codacy    # initializes the tools for running the codacy targets
make init-hooks     # initializes github hooks for pre-commit, etc
make init-packages  # initializes and downloads packages dependencies
make init-sources   # initializes sources by generating mocks, etc
```


### Release targets

Finally, the [Makefile](Makefile) supports targets for releasing the provided
packages as library.

```bash
make bump <version>  # bumps version to prepare a new release
make release         # creates the release tags in the repository
```


## Compatibility

This [Makefile](Makefile) is making extensive use of GNU tools but is supposed
to be compatible to all recent Linux and MacOS versions. Since MacOS is usually
a couple of years behind in applying the GNU standards, we document the
restrictions this requires here.


### `sed` in place substitution

In MacOS we need to execute need to add `-e '<cmd>'` after `sed -i` since else
the command section is not automatically restricted to a single argument. In
linux this restriction is automatically applied to the first argument.


### `realpath` not supporting relative offsets

In MacOS we need to manually remove the path-prefix from `realpath`, since the
default in-`bash` fallback version does not provide the `--relative-base`
argument option.
