# Development

This action's core functionality is contained in a Go CLI. The `action.yml` at the root of
this repo defines the action itself, as a composite action, that builds and then calls the CLI.

## Manually Testing the CLI

It's useful to be able to just run the CLI locally to manually poke around and see your changes.

The workflow for this is to run `make install` any time you want to test your local changes,
this runs tests and does a fully dogfooded from-scratch build and installs it into your PATH.

You can override the destination by setting `DESTDIR` e.g. `make install DESTDIR=/some/path`

You can then run `actions-go-build <blah>` to test things out.

## Tests

**All code changes in this Action should be accompanied by new or updated tests documenting and
preserving the new behaviour.**

### CLI Tests

Run `make test` to run the CLI tests locally. These tests are run in CI by the
`test.yml` workflow.

### End-to-End Action Tests

There are tests that exercise the action itself, see
[`test-build.yml`](.github/workflows/test-build.yml).

That test file invokes the entire test suite `self-test-suite.yml` twice, once on Linux
runners and again on macOS runners. They are the two runner OSs supported by this action.

The `self-test-suite.yml` calls the test harness action (see below) multiple times to
simulate different conditions and assert on the results.

There is a special action defined in `self-test/action.yml` which is a test harness for
the main action. It has all the same inputs as the action itself, and passes these through
but sets some alternative default input values that make testing easier, and provides built-in
assertions. This keeps the test suite nice and concise.

### Example Code

The example code in the readme is real, working code that we run on every push.
This ensures that it doesn't go out of date.

[`.github/workflows/example.yml`](.github/workflows/example.yml)
and
[`.github/workflows/example-matrix.yml`](.github/workflows/example-matrix.yml).

## Documentation

Wherever possible, the documentation in this README is generated from source code to ensure
that it is accurate and up-to-date. Run `make docs` to update it.

The `make docs` target updates both the readme and the changelog
(`make readme` and `make changelog`).

### Readme

The Readme contains blocks like this:

```
<!-- insert:path/to/executable -->
...
... Some content here.
...
<!-- end:insert:path/to/executable -->
```

These blocks are auto-updated when we run `make readme` or `make docs`.
They are updated by the script at `dev/docs/readme_update` which calls
`dev/docs/insert` to replace the contents of those blocks with the output
from the named executable.

For example, the executable at `dev/docs/print_example_workflow` gets the contents
of an example workflow, and formats it as markdown.

Another more complex example is `dev/docs/environment_doc` which uses the `run`
script at the root of this repo to build the CLI, then execute it to get the
output of `inspect -describe-build-env` and format it as a markdown table.

Doing all these things by hand would be error-prone, and easy to forget.
With these automations the source code and documentation are guaranteed to
be congruent.

### Changelog

All changes should be accompanied by a corresponding changelog entry.
Each version has a file named `dev/changes/v<VERSION>.md` which contains
the changes added during development of that version.

When a release is made (see below) the contents of the changelog file are
copied into the release description.

## Releasing

You can release a new version by running `make release`.
This uses the version string from `dev/VERSION` to add tags,
get the corresponding changelog entries, and create a new GitHub
release.
