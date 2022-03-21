# Troubleshooting go-global-update

This document describes known problems that can occur while trying to update
globally-installed binaries using go-global-update.

<!-- vim-markdown-toc GFM -->

- [E001 - binaries built from source](#e001---binaries-built-from-source)
- [E002 - module found but does not contain package](#e002---module-found-but-does-not-contain-package)
- [E003 - module declares its path as ... but was required as ...](#e003---module-declares-its-path-as--but-was-required-as-)
- [E004 - go.mod contains `replace` directives](#e004---gomod-contains-replace-directives)

<!-- vim-markdown-toc -->

## E001 - binaries built from source

go-global-update can only update binaries in `GOBIN` installed using
`go install [path URL]@latest`. go-global-update cannot update binaries which
were built from source (inside of the source repository of the binary) using either:

- `go install` (without arguments)
- `go build` (either without arguments or pointing to a specific Go file)

This is because the version of the installed binary is discarded and will
appear as `(devel)` in `go version -m [binary-name]`. Moreover, for binaries
built using `go build`, the path URL is also discarded and replaced with
`command-line-arguments`, which makes it impossible to know what is the path
URL for the binary. The path URL is required when updating the binary.

```sh
$ git clone git@github.com:StevenACoffman/toolbox.git
$ go build -o bin/jp cmd/jira-pull/jp.go
$ cd bin
$ go version -m jp
jp: go1.17.5
	path	command-line-arguments
	mod	github.com/StevenACoffman/toolbox	(devel)
	dep	github.com/emirpasic/gods	v1.12.0	h1:QAUIPSaCu4G+POclxeqb3F+WPpdKqFGlw36+yOzGlrg=
	dep	github.com/go-git/gcfg	v1.5.0	h1:Q5ViNfGF8zFgyJWPqYwA7qGFoMTEiBmdlkcfRmpIMa4=
	dep	github.com/go-git/go-billy/v5	v5.0.0	h1:7NQHvd9FVid8VL4qVUMm8XifBK+2xCoZ2lSk0agRrHM=
  # the rest of the output ...
```

Additionally, if you are using a binary built from source instead of some
publicly available version, this can mean that you introduced some modifications
to it and expect them to be retained. Updating to the latest version could
remove those modification, which could quickly become annoying.

## E002 - module found but does not contain package

Sometimes repository maintainers decide to extract the CLI binary of some module
to a separate module in a separate repository or a different directory in the
same repository. Regardless, the path URL to the directory with the binary
module changes. It means updating the previous path URL will not work, because
it will not have a CLI binary there.

```sh
$ go-global-update
Upgrading cobra to v1.4.0 ... ❌
    Could not upgrade package
go: downloading github.com/spf13/cobra v1.4.0
go install: github.com/spf13/cobra/cobra@latest:
  module github.com/spf13/cobra@latest found (v1.4.0),
  but does not contain package github.com/spf13/cobra/cobra
```

In that case, you have to find the new path for the CLI, remove the old binary,
and install the new one once:

```sh
cd $(which cobra)/..
rm cobra
go install github.com/spf13/cobra-cli@latest
```

Next runs of `go-global-update` will correctly keep `cobra-cli` up-to-date.

Known extracted binaries:

- [`cobra`](https://github.com/spf13/cobra) was extracted to
  [`cobra-cli`](https://github.com/spf13/cobra-cli) in [v1.3.0](https://github.com/spf13/cobra-cli/releases/tag/v1.3.0)

## E003 - module declares its path as ... but was required as ...

Similarly to [E002](#e002---module-found-but-does-not-contain-package), the
whole repository containing the binary may be moved to a different name or to a
different organization. This would manifest itself in the error:

```sh
$ go-global-update
Upgrading gnostic to v0.6.6 ... ❌
  Could not upgrade package
go install: github.com/googleapis/gnostic@latest: github.com/googleapis/gnostic@v0.6.6: parsing go.mod:
  module declares its path as: github.com/google/gnostic
          but was required as: github.com/googleapis/gnostic
```

This happens because GitHub automatically redirects [the old
URL](https://github.com/googleapis/gnostic) to [the new URL of the
repository](https://github.com/google/gnostic), but [the `go.mod` only lists the
new path to the
repository](https://github.com/google/gnostic/blob/418d86c152e3f607fa625e9aca135091e574811f/go.mod#L1),
which `go` does not like.

To mitigate the problem, remove the old binary and install it once using the new
path URL:

```sh
cd $(which gnostic)/..
rm gnostic
go install github.com/google/gnostic@latest
```

Next runs of `go-global-update` will correctly keep `gnostic` up-to-date.

Known moved repositories:

- <https://github.com/google/gnostic> (moved to the
  [google](https://github.com/google) organization from
  [googleapis](https://github.com/googleapis))

## E004 - go.mod contains `replace` directives

Some binaries contain `go.mod` files which contain [`replace`
directives](https://go.dev/ref/mod#go-mod-file-replace). `go install` does not
handle such directives and exits with an error when trying to install or update
such a binary.

```sh
$ go-global-update
Upgrading dive to v0.10.0 ... ❌
  Could not upgrade package
go: downloading github.com/wagoodman/dive v0.10.0
go install: github.com/wagoodman/dive@latest (in github.com/wagoodman/dive@v0.10.0):
    The go.mod file for the module providing named packages contains one or
    more replace directives. It must not contain directives that would cause
    it to be interpreted differently than if it were the main module.
```

There are 2 ways to solve this problem:

1. Ask the module maintainer to remove the `replace` directive. This way `go install` will be able to correctly install it.

2. Clone the repository locally and run `go install` on it.

   This will install the binary into your `GOBIN`. Future runs of
   `go-global-update` will not update such a binary, because it was built from
   source and will trigger [E001](#e001---binaries-built-from-source).
