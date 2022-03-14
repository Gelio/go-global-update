# go-global-update

![go-global-update screenshot](https://user-images.githubusercontent.com/889383/158154631-60da69af-7159-4a33-b564-1a4ae1ce3328.png)

Update globally installed go binaries.

The missing go command similar to `npm -g update` or [cargo install-update](https://github.com/nabijaczleweli/cargo-update).

## Requirements

- Go 1.16 or higher

## Installation

```sh
go install github.com/Gelio/go-global-update@latest
```

## Usage

```sh
go-global-update
```

You can also do a dry run without update the binaries:

```sh
go-global-update --dry-run
```

or update just a handful of binaries:

```sh
go-global-update gofumpt
```

For more information, see

```sh
go-global-update --help
```

## Upgrading `go-global-update`

`go-global-update` will take care of updating itself when it updates other binaries.

## How it works

`go-global-update` consists of the following steps:

1. Determine binaries to inspect.

   Either use the list of provided arguments or all executables installed in
   your `go env GOBIN` (or `$(go env GOPATH)/bin`).

1. Inspect where each executable came from (by running
   `go version -m [executable name]` and checking the `path`),

1. Check the latest version for each binary using
   `go list -m -f "{{.Version}}" [path]`

1. If the binary has a newer version, run `go install [package path]@latest` to
   update it.
