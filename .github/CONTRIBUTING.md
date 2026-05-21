# Contributing

**Thank you for your interest in making this project more awesome!**

There are multiple ways of getting involved:

* [Reporting bugs](#reporting-bugs)
* [Suggesting features](#suggesting-features)
* [Contributing code](#contributing-code)

Below are a few guidelines we would like you to follow. If you need help, please
reach via issue.


## Reporting bugs

Reporting bugs is one of the best ways to contribute. Before creating a bug
report, please check whether an [issue](../../issues) reporting the same problem
already exist. If there is such an issue, please add your information as a
comment.

To report a new bug you should open an issue that summarizes the bug and set
the label to "bug". If you want to provide a fix along with your bug report:
That is great!

In this case please create a pull request as described in section
[Contributing code](#contributing-code) and report the bug directly in the
pull request body.


## Suggesting features

To request a new feature you should open an [issue](../../issues/new) and
summarize the desired functionality and its use case with an example. Set the
issue label to "enhancement".


## Contributing code

This is a rough outline of what the workflow for code contributions looks like:

* Check whether your contribution is aligned with the goals and principles of
  the project and fits into the current scope.
* Fork the repository on GitHub and create a feature branch, from where you want
  to base your work. The base is usually `main`.
* Make commits of logical units and use `git commit --sign-off` to comply with
  [DCO](https://developercertificate.org/).
* Write good commit messages (see below) and push your changes to the feature
  branch in your fork of the repository.
* All features need additionally include documentation for developers as
* Submit a pull request.

Thanks for your contributions!


### Code style

The code is quality checked and formatted using [`go-make`][go-make]. Please
run `make test lint` on your code before making any commit or pull request:

* The coding style suggested by the Golang community is the preferred one.
  The project applies the style guidelines fostered by [`gofmt`][gofmt],
  [`gofumpt`][gofumpt], [`goimports`][goimports], and [`golines`][golines].
  For the cases that are not covered by these formatters, see the
  [style doc][go-style] for details.

  Please run `make format lint` before committing any code and fix the
  remaining formatting issues.

* The code quality is checked by using the `max` coding standard defined for
  [`golangci-lint`][golangci] by [`go-make`][go-make]. This requires a high
  commitment to extensive quality standards defined by the go community.

  Please run `make lint fix` before any commit and fix the remaining code
  quality issues.

* All changes need to include unit or component tests using the idiomatic
  patterns defined by the [`go-testing`][go-testing] framework providing a
  [sensible, high-quality code coverage][testing].

[go-make]: <https://golang.org/cmd/gofmt/>
[go-style]: <https://go.dev/wiki/CodeReviewComments>
[go-testing]: <https://github.com/tkrop/go-testing>
[golangci]: <https://github.com/golangci/golangci-lint>
[gofmt]: <https://pkg.go.dev/cmd/gofmt>
[gofumpt]: <https://github.com/mvdan/gofumpt>
[goimports]: <https://pkg.go.dev/golang.org/x/tools/cmd/goimports>
[golines]: <https://github.com/segmentio/golines>
[testing]: <https://ricomariani.medium.com/100-unit-testing-now-its-ante-f0e2384ffedf>


### Commit messages

The project is applying the standard for [conventional commits][convent-commit]
with the addition, that the subject line should reference at least one issue.
Your commit messages ideally answers two questions: what is changed and why.
The subject line should feature the "what" and the body of the commit should
describe the "why" - if not already provided by the issue reference in the
subject line (we encourage you to leverage the pull request body for the
latter).

When creating a pull request, the body comment should also provide a detailed
explanation about what was changed and why it was changed, and reference all
the related issues - if available.

**Have fun and enjoy hacking!**

[convent-commit]: <https://www.conventionalcommits.org/en/v1.0.0/>


## Governance - final decisions

In case of disagreement between contributors and maintainers, the project owner
and lead will make all final decisions.
