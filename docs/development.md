

## Tests

**All code changes in this Action should be accompanied by new or updated tests documenting and
preserving the new behaviour.**

Run `make test` to run the tests.

There are also tests that exercise the action itself, see
[`.github/workflows/test.yml`](https://github.com/hashicorp/actions-go-build/blob/main/.github/workflows/test.yml).
These tests use a reusable workflow for brevity, and assert both passing and failing conditions.

The example code is also tested to ensure it really works, see
[`.github/workflows/example.yml`](https://github.com/hashicorp/actions-go-build/blob/main/.github/workflows/example.yml)
and
[`.github/workflows/example-matrix.yml`](https://github.com/hashicorp/actions-go-build/blob/main/.github/workflows/example-matrix.yml).

## Documentation

Wherever possible, the documentation in this README is generated from source code to ensure
that it is accurate and up-to-date. Run `make docs` to update it.

### Changelog

All changes should be accompanied by a corresponding changelog entry.
Each version has a file named `dev/changes/v<VERSION>.md` which contains
the changes added during development of that version.

## Releasing

You can release a new version by running `make release`.
This uses the version string from `dev/VERSION` to add tags,
get the corresponding changelog entries, and create a new GitHub
release.

## Implementation

### TODO

- Add a reusable workflow for better optimisation (i.e. running in parallel jobs)
- See **ENGSRV-083** (internal only) for future plans.
