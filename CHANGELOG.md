# Changelog

## Unreleased

### Added

- [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) containing a description of
  common problems faced while updating packages with `go-global-update`.

## v0.1.2 (2022-03-21)

A patch release adding support for go 1.18.

### Fixed

- Support go 1.18 <https://github.com/Gelio/go-global-update/pull/7>

  Go 1.18 changed the format of the `go version -m [binary-name]` command to no
  longer include the `mod` information for binaries built from source using `go build main.go`.

  go-global-update now parses such outputs and does not attempt to check for
  the latest version of such binaries.

## v0.1.1 (2022-03-21)

A patch release containing fixes for some of the edge cases found in
<https://github.com/Gelio/go-global-update/issues/3>.

### Fixed

- Skip updating binaries built from source.

  Binaries built from source (either using `go build main.go` or `go install`
  in the cloned repository) likely have been modified prior to being built.
  Updating them would likely throw away these changes and end up being annoying
  for engineers who want to keep their modified versions.

  Moreover, packages built using `go build main.go` have
  `command-line-arguments` set as their `path` in `go version -m binary-name`.
  This makes it impossible to update automatically.

- Filesystem path handling on Windows.

  Use correct separator for filesystem paths on Windows. This allows using this
  tool on Windows.

## v0.1.0 (2022-03-14)

- Complete the basic functionality of upgrading globally installed executables.
- Add `--dry-run`, `--verbose`, `--debug` flags.
- Support upgrading only specified binaries by accepting arguments.
- Add integration tests and configure CI.
