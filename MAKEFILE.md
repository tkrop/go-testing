<!-- do not change!!! updated by 'make update-make' if not source of truth -->

# Makefile

This repository provides a generic [Makefile](Makefile) that is majorly working
by the following conventions:

1. Place all available commands under `cmd/<name>/main.go`.
2. Use a single 'config' package to read the configuration for all commands.
3. Use a common 'Dockerfile' to install all commands in a container image.

The [Makefile](Makefile) also allows to call run the commands/services and run
test via `make run/test-* [args]`, e.g. `make test-unit app/service` runs all
the unit tests in the directory `app/service`.

**Warning:** The [Makefile](Makefile) installs a `pre-commit` hook overwriting
and deleting any pre-existing hook that requires `make lint test-unit` to run
successfully before allowing to commit.


## Standard targets

The [Makefile](Makefile) supports the following often used standard targets
beside a high number of specialized and even dynamic targets.

```bash
make all     # first choice target to init, build, and test
make cdp     # select of targets to init, build, and test as in pipeline
make init    # inits project by downloading dependencies
make test    # generates and builds sources to execute tests
make lint    # generates and builds sources and lints sources
make build   # creates binary files of commands
make install # installs binary files of commands in '${GOPATH}/bin'
make delete  # deletes binary files of commands from '${GOPATH}/bin'
make clean   # cleans up the project removing all created files
```

All this targets are supposed to work out of the box to setup the project and
execute the most important tasks, however, some specialized commands provide
more features and may provide a quicker response on building and testing.


### Test targets

Often it is more efficient or even necessary to execute the finegrained test
targets to complete a task.

```bash
make test        # short cut for 'test-all'
make test-all    # executes the complete tests suite
make test-unit   # executes only unit tests by setting the short flag
make test-cover  # opens the test coverage report in the browser
```

In addition, it is possible to restrict test target execution to packages,
files and test cases as follows:

* For a single package use `make test-(unit|all) <package> ...`).
* For a single test file (`make test[-(unit|all) <package>/<file>_test.go ...`).
* For a single test case (`make test[-(unit|all) <package>/<test-name> ...`).


### Linter targets

The [Makefile](Makefile) supports different targets that can help with linting
as well as with fixing the linter problems - if possible.

```bash
make lint        # short cut to execute 'lint-base lint-apis'
make lint-base   # lints the go-code using a baseline setting
make lint-plus   # lints the go-code using an advanced setting
make lint-all    # lints the go-code using an all-in expert setting
make lint-api    # lints the api specifications in '/zalando-apis'

make format      # formats the code to fix selected linter violations
```

The `lint-*` targets allow command line arguments:

1. The keyword `list` to display the linter configurations, or
2. `<linter>,...` a list of linters to enable for a quick checks.

To default target for `make lint` can be customized via `TARGETS_LINT` in
`Makefile.vars`. The default is `lint-base lint-apis`.

The default linter setup is providing a golden path with three expert levels
out-of-the-box, i.e. a `base` setting, a challanging `plus` setting, and an
expert `all` setting that runs all but the disabled linters.

The lint expert levels can be customized in two ways.

1. Linters can be enabled and disabled providing the linter names to the space
   separated lists via the variables `LINT_ENABLED`, `LINT_DISABLED`, and
   `LINT_ADVANCED`.
2. The linters settings can be configuration via the `.golangci.yaml` file.

Howver, customizing the `.golangci.yaml` is currently not advised, since the
`Makefile` is updating and enforcing a common version.


## Install targets

The install targets installs the latest build version of a command in the
`${GOPATH}/bin` directory for simple command line execution.

```bash
make install      # installs all commands using the platform binaries
make install-(*)  # installs the matched command using the platform binary
```

If a command, service, job has not been build before, it is first build.

**Note:** Please use carefully, if your project uses common command names.


## Delete targets

The delete targets delete the latest installed command from `${GOPATH}/bin`.

```bash
make install      # Deletes all commands
make install-(*)  # Deletes the matched command
```

**Note:** Please use carefully, if your project uses common command names.


## Image targets

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


## Run targets

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
make update       # short cut to execute update-deps
make update-deps  # updates the project dependencies to the latest version
make update-go    # updates go to the latest available version
make update-make  # updates the Makefile to the latest available version
```

It is advised to use and extend this targets when necessary.


### Cleaning targets

The [Makefile](Makefile) is designed to clean up everything it has created by
executing the following targets.

```bash
make clean         # short cut for clean-init, clean-build, and clean-test
make clean-init    # cleans up all resources created by the init targets
make clean-build   # cleans up all resources created by the build targets
make clean-test    # cleans up all resources created for the test targets
make clean-run(-*) # cleans up all resources created for the run targets
```


## Initialization targets

The [Makefile](Makefile) supports initialization targets that are usally
already added as prequisits for targets that need them. So there is usually
no need to call them directly.

```bash
make init           # short cut for init-tools init-hooks init-packages
make init-tools     # initializes the requeired tool using `go install`
make init-hooks     # initializes github hooks for pre-commit, etc
make init-packages  # initializes and downloads packages dependencies
make init-sources   # initializes sources by generating mocks, etc
```


## Releasing targets

Finally, the [Makefile](Makefile) supports targets for releasing the
provided packages as library.

```bash
make bump      # bumps version to prepare release
make release   # creates release tags in repository
```


## Setup and customization

To customize the behavior of the Makefile there exist multiple extension points
that can be used to setup additional variables and targets that modify the
behavior of the [Makefile](Makefile).

* [Makefile.vars](Makefile.vars) allows to modify the behavior of standard
  targets by customizing and defining additional variables (see Section
  [Modifying variables](#modifying-variables) for more details).
* [Makefile.defs](Makefile.defs) allows to customize the runtime environment
  for executing of commands (see Section [Running commands](#running-commands)
  for more details).
* [Makefile.targets](Makefile.targets) is an optional extension point that
  allows to define arbitrary custom targets.


### Modifying variables

TODO: add content!

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
RUN_DEPS as needed to [Makefile.vars](Makefile.vars).

You can also override the default setup via the `DB_HOST`, `DB_PORT`,
`DB_NAME`, `DB_USER`, and `DB_PASSWORD` variables, but this is optional.

**Note:** when running test against a DB you usually have to extend the
default `TEST_TIMEOUT` of 10s to a less aggressive value.

To enable AWS localstack you have to add `run-aws` to the `TEST_DEPS` and
`RUN_DEPS`. You may also need to provide a sensible setup of AWS services via
the `AWS_SERVICES` variable (default is `sqs s3`).

**Note:** Currently, we the [Makefile](Makefile) does not support command
specific command-line arguments or environment variables. It is assuming that
command are following the principles of the [Twelf Factor App][12factor]
supportig setup via application specific environment variables.

[12factor]: https://12factor.net/ "Twelf Factor App"
