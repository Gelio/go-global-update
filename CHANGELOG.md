# Changelog

## Unreleased

- Skip updating binaries built from source.

  Binaries built from source (either using `go build main.go` or `go install`
  in the cloned repository) likely have been modified prior to being built.
  Updating them would likely throw away these changes and end up being annoying
  for engineers who want to keep their modified versions.

  Moreover, packages built using `go build main.go` have
  `command-line-arguments` set as their `path` in `go version -m binary-name`.
  This makes it impossible to update automatically.

## v0.1.0 (2022-03-14)

- Complete the basic functionality of upgrading globally installed executables.
- Add `--dry-run`, `--verbose`, `--debug` flags.
- Support upgrading only specified binaries by accepting arguments.
- Add integration tests and configure CI.
