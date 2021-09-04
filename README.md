# Go global update

Update go global executables.

## Usage

First, install the executable:

```sh
go install github.com/Gelio/go-global-update@latest
```

Then, run it with:

```sh
go-global-update
```

## How it works

`go-global-update` consists of the following steps:

1. List all executables installed in your `go env GOBIN` (or
   `$(go env GOPATH)/bin`)

2. Inspect where each executable came from (by running
   `go version -m [executable name]` and checking the `path`),

3. Run `go install [package path]@latest` to update each package.
