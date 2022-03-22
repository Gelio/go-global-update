# Contributing to go-global-update

Contributions are welcome!

Found a bug or want a new feature? Report an issue.

Want to take a stab at making a code change? Pull requests are also welcome!

## Pull request checklist

Before submitting a pull request, make sure to complete the following steps:

1. Format the Go code usign [gofumpt](https://github.com/mvdan/gofumpt)
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
