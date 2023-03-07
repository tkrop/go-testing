# Mock Generator

* The `source package` is provided via `--source(-pkg)=...`. It can also be
  provided by a command line argument starting with a lower case letter since
  valid mocked interfaces are expected to start with an upper letter.
* The `source file` is provided via `--source(-file)=...`. It can also be
  provided by a command line argument matching the regular expression
  `^[a-z.].*\.go$`.
* The `target package` is provided via `--target(-pkg)=...`. It can also be
  provided by a command line argument matching the regular expression
  `^[a-z_]*`.
* The `target file` is provided via `--target(-file)=...`. It can be also be
  provided by a command line argument matching the regular expression
  `^[a-z.].*/mock_[^/]*\.go`).
