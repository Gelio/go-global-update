# Contributing to go-global-update

Contributions are welcome!

Found a bug or want a new feature? Report an issue.

Want to take a stab at making a code change? Pull requests are also welcome!

## Pull request checklist

Before submitting a pull request, make sure to complete the following steps:

1. Format and lint the Go code with [golangci-lint](https://golangci-lint.run/).
   Run `go tool -modfile=tools/go.mod golangci-lint fmt` and
   `go tool -modfile=tools/go.mod golangci-lint run ./...` (or use a locally
   installed golangci-lint of the same version).
1. Ensure the tests pass by running `go test ./...`
1. Regenerate the Markdown table of contents in case you changed Markdown files.
   Run `npm run generate-toc`
1. Format the Markdown files using [prettier](https://prettier.io/). Run
   `npm run format-docs:write`

It is best if your change is covered by tests. If there are none, consider
adding new test cases. Do not worry if you do not know how, we can figure it out
during the pull request review.

To run `npm` commands you will need [Node](https://nodejs.org/en/) installed,
which comes bundled with `npm`. Run `npm install` to install the necessary
dependencies before running those `npm run ...` commands.

## Changelog

[CHANGELOG.md](CHANGELOG.md) is for users. Add new entries under
`## Unreleased`, using these sections:

- `### ⚠ Breaking Changes` (only when needed)
- `### Features`
- `### Bug Fixes`
- `### Internal`

Only changes in behavior users can observe go under Breaking Changes, Features,
or Bug Fixes. Anything users cannot see (refactors, dependency updates, CI,
build and release plumbing, how `--version` is derived) goes under Internal.

The sections match [Conventional Commits](https://www.conventionalcommits.org/)
types: `feat` goes under Features, `fix` under Bug Fixes, `!` or
`BREAKING CHANGE` under Breaking Changes, and every other type under Internal.
Pick the commit type by its user-visible effect.

Entries for v0.2.5 and earlier use the previous section headings.

## Releasing

The version reported by `--version` comes from the Go build info, so there is no
version string to edit in the code. To release, tag the commit as `vX.Y.Z` and
add the matching entry to [CHANGELOG.md](CHANGELOG.md).
