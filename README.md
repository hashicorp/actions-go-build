# Go Build Action

_**Build and package a (reproducible) Go binary.**_

- **Define** the build.
- **Assert** that it is reproducible (optionally).
- **Use** the resultant artifacts in your workflow.

_This is intended for internal HashiCorp use only; Internal folks please refer to RFC **ENGSRV-084** for more details._

<!-- insert:dev/docs/table_of_contents -->
## Table of Contents
* [Table of Contents](#table-of-contents)
* [Features](#features)
* [Usage](#usage)
  * [Example Workflows](#example-workflows)
  * [Inputs](#inputs)
  * [Build Environment](#build-environment)
  * [Reproducibility Assertions](#reproducibility-assertions)
  * [Build Instructions](#build-instructions)
  * [Ensuring Reproducibility](#ensuring-reproducibility)
* [Development](#development)
  * [Tests](#tests)
  * [Documentation](#documentation)
  * [Releasing](#releasing)
  * [Bash, dreaded bash.](#bash-dreaded-bash)
  * [Future Implementation Options](#future-implementation-options)
  * [TODO](#todo)
<!-- end:insert:dev/docs/table_of_contents -->

## Features

- **Results are zipped** using standard HashiCorp naming conventions.
- **You can include additional files** in the zip like licenses etc.
- **Convention over configuration** means minimal config required.
- **Reproducibility** is checked at build time.
- **Fast feedback** if accidental nondeterminism is introduced.

## Usage

This Action can run on both Ubuntu and macOS runners.

### Example Workflows

#### Minimal(ish) Example

This example shows building a single `linux/amd64` binary.

[See this simple example workflow running here](https://github.com/hashicorp/actions-reproducible-build/actions/workflows/example.yml).

<!-- insert:dev/docs/print_example_workflow example.yml -->
```yaml
name: Minimal(ish) Example
on: [push]
jobs:
  example:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build
        uses: hashicorp/actions-reproducible-build@main
        with:
          product_name: example-app
          product_version: 1.2.3
          go_version: 1.18
          os: linux
          arch: amd64
          instructions: |-
            cd ./testdata/example-app
            go build -o "$BIN_PATH" -trimpath -buildvcs=false
```
<!-- end:insert:dev/docs/print_example_workflow example.yml -->

#### More Realistic Example

This example shows usage of the action inside a matrix configured to produce
binaries for different platforms. It also injects the version, revision, and
revision time into the binary via `-ldflags`, uses the `netcgo` tag for darwin,
and disables CGO for linux and windows builds.

[See this matrix example workflow running here](https://github.com/hashicorp/actions-reproducible-build/actions/workflows/example-matrix.yml).

<!-- insert:dev/docs/print_example_workflow example-matrix.yml -->
```yaml
name: Matrix Example
on: [push]
jobs:
  example:
    runs-on: ${{ matrix.runner }}
    strategy:
      matrix:
        include:
          - { runner: macos-latest,  os: darwin,  arch: amd64, tags: netcgo        }
          - { runner: macos-latest,  os: darwin,  arch: arm64, tags: netcgo        }
          - { runner: ubuntu-latest, os: linux,   arch: amd64, env:  CGO_ENABLED=0 }
          - { runner: ubuntu-latest, os: linux,   arch: amd64, env:  CGO_ENABLED=0 }
          - { runner: ubuntu-latest, os: windows, arch: amd64, env:  CGO_ENABLED=0 }
    steps:
      - uses: actions/checkout@v3
      - name: Build
        uses: hashicorp/actions-reproducible-build@main
        with:
          product_name: example-app
          product_version: 1.2.3
          go_version: 1.18
          os: ${{ matrix.os }}
          arch: ${{ matrix.arch }}
          instructions: |-
            cd ./testdata/example-app && \
            ${{ matrix.env }} \
              go build \
                -o "$BIN_PATH" \
                -trimpath \
                -buildvcs=false \
                -tags="${{ matrix.tags }}" \
                -ldflags "
                  -X 'main.Version=$PRODUCT_VERSION'
                  -X 'main.Revision=$PRODUCT_REVISION'
                  -X 'main.RevisionTime=$PRODUCT_REVISION_TIME'
                "
```
<!-- end:insert:dev/docs/print_example_workflow example-matrix.yml -->

### Inputs

<!-- insert:dev/docs/inputs_doc -->
|  Name                                     |  Description                                                                                              |
|  -----                                    |  -----                                                                                                    |
|  `product_name`&nbsp;_(optional)_         |  Used to calculate default `bin_name` and `zip_name`. Defaults to repository name.                        |
|  **`product_version`**&nbsp;_(required)_  |  Version of the product being built.                                                                      |
|  **`go_version`**&nbsp;_(required)_       |  Version of Go to use for this build.                                                                     |
|  **`os`**&nbsp;_(required)_               |  Target product operating system.                                                                         |
|  **`arch`**&nbsp;_(required)_             |  Target product architecture.                                                                             |
|  `reproducible`&nbsp;_(optional)_         |  Assert that this build is reproducible. Options are `assert` (the default), `report`, or `nope`.         |
|  `bin_name`&nbsp;_(optional)_             |  Name of the product binary generated. Defaults to `product_name` minus any `-enterprise` suffix.         |
|  `zip_name`&nbsp;_(optional)_             |  Name of the product zip file. Defaults to `<product_name>_<product_version>_<os>_<arch>.zip`.            |
|  **`instructions`**&nbsp;_(required)_     |  Build instructions to generate the binary. See [Build Instructions](#build-instructions) for more info.  |
<!-- end:insert:dev/docs/inputs_doc -->

### Build Environment

When the `instructions` are executed, there are a set of environment variables
already exported that you can make use of
(see [Environment Variables](#environment-variables) below).

#### Environment Variables

<!-- insert:dev/docs/environment_doc -->
|  Name                     |  Description                                                         |
|  -----                    |  -----                                                               |
|  `TARGET_DIR`             |  Absolute path to the zip contents directory.                        |
|  `PRODUCT_NAME`           |  Same as the `product_name` input.                                   |
|  `PRODUCT_VERSION`        |  Same as the `product_version` input.                                |
|  `PRODUCT_REVISION`       |  The git commit SHA of the product repo being built.                 |
|  `PRODUCT_REVISION_TIME`  |  UTC timestamp of the `PRODUCT_REVISION` commit in iso-8601 format.  |
|  `BIN_NAME`               |  Name of the Go binary file inside `TARGET_DIR`.                     |
|  `BIN_PATH`               |  Same as `TARGET_DIR/BIN_NAME`.                                      |
|  `OS`                     |  Same as the `os` input.                                             |
|  `ARCH`                   |  Same as the `arch` input.                                           |
|  `GOOS`                   |  Same as `OS`                                                        |
|  `GOARCH`                 |  Same as `ARCH`.                                                     |
<!-- end:insert:dev/docs/environment_doc -->

### Reproducibility Assertions

The `reproducible` input has three options:

- `assert` (the default) means perform a verification build and fail if it's not identical to the primary build.
- `report` means perform a verification build, log the results, but do not fail.
- `nope`   means do not perform a verification build at all.

See [Ensuring Reproducibility](#ensuring-reproducibility), below for tips on making your build reproducible.

### Build Instructions

The `instructions` input is a bash script that builds the product binary.
It should be kept as simple as possible.
Typically this will be a simple `go build` invocation,
but it could hit a make target, or call another script.
See [Example Build Instructions](#example-build-instructions)
below for examples of valid instructions.

The instructions _must_ use the environment variable `$BIN_PATH`
because the minimal thing they can do is to write the compiled binary to `$BIN_PATH`.

In order to add other files like licenses etc to the zip file, you need to
write them into `$TARGET_DIR` in your build instructions.

#### Example Build Instructions

The examples below all illustrate valid build instructions using `go build` flags
that give the build some chance at being reproducible.

---

Simplest Go 1.17 invocation. (Uses `-trimpath` to aid with reproducibility.)

```yaml
instructions: go build -o "$BIN_PATH" -trimpath
```

---

Simplest Go 1.18+ invocation. (Additionally uses `-buildvcs=false` to aid with reproducibility.)

```yaml
instructions: go build -o "$BIN_PATH" -trimpath -buildvcs=false
```

---

More complex build, including copying a license file into the zip and `cd`ing into
a subdirectory to perform the go build.

```yaml
instructions: |
	cp LICENSE "$TARGET_DIR/"
	cd sub/directory
	go build -o "$BIN_PATH" -trimpath -buildvcs=false
```

---

An example using `make`:

```yaml
instructions: make build
```

With this Makefile:

```Makefile
build:
	go build -o "$BIN_PATH" -trimpath -buildvcs=false
```

---

See also the [example workflow](#example-workflow) above,
which injects info into the binary using `-ldflags`.

### Ensuring Reproducibility

If you are aiming to create a reproducible build, you need to at a minimum ensure that
your build is independent from the _time_ it is run, and from the _path_ that the module
is at on the filesystem.

#### Build Time

Embedding the actual 'build time' into your binary will ensure that it isn't reproducible,
because this time will be different for each build. Instead, you can use the
`PRODUCT_REVISION_TIME` which is the time of the latest commit, which will be the same
for each build of that commit.

#### Build Path

By default `go build` embeds the absolute path to the source files inside the binaries
for use in stack traces and debugging. However, this reduces reproducibility because
that path is likely to be different for different builds.

Use the `-trimpath` flag to remove the portion of the path that is dependent on the
absolute module path to aid with reproducibility.

#### VCS information

Go 1.18+ embeds information about the current checkout directory of your code, including
modified and new files. In some cases this interferes with reproducibility. You can
turn this off using the `-buildvcs=false` flag.

## Development

- This Action uses extensionless executable bash scripts in `scripts/` to perform each step.
- There are also `.bash` files in `scripts/` which define functions used in the executables.
- Both executable and library bash files have BATS tests which are defined inside files with
  the same name plus a `.bats` extension.

### Tests

**All code changes in this Action should be accompanied by new or updated tests documenting and
preserving the new behaviour.**

Run `make test` to run the BATS tests which cover the scripts.

There are also tests that exercise the action itself, see
[`.github/workflows/test.yml`](https://github.com/hashicorp/actions-go-build/blob/main/.github/workflows/test.yml).
These tests use a reusable workflow for brevity, and assert both passing and failing conditions.

The example code is also tested to ensure it really works, see
[`.github/workflows/example.yml`](https://github.com/hashicorp/actions-go-build/blob/main/.github/workflows/example.yml)
and
[`.github/workflows/example-matrix.yml`](https://github.com/hashicorp/actions-go-build/blob/main/.github/workflows/example-matrix.yml).

### Documentation

Wherever possible, the documentation in this README is generated from source code to ensure
that it is accurate and up-to-date. Run `make docs` to update it.

#### Changelog

All changes should be accompanied by a corresponding changelog entry.
Each version has a file named `dev/changes/v<VERSION>.md` which contains
the changes added during development of that version.

### Releasing

You can release a new version by running `make release`.
This uses the version string from `dev/VERSION` to add tags,
get the corresponding changelog entries, and create a new GitHub
release.

### Bash, dreaded bash.

This Action is currently written in Bash.

The primary reason is that Bash makes it trivial to call other programs and handle the
results of those calls. Relying on well-known battle-tested external programs like
`sha256sum` and `bash` itself (for executing the instructions) seems like a reasonable
first step for this Action, because they are the tools we'd use to perform this work
manually.

For the initial development phase, well-tested Bash is also useful because of the speed
and ease of deployment. It is present on all runners and doesn't require a compilation
and deployment step (or alternatively installing the toolchain to perform that
compilation).

### Future Implementation Options

Once we're happy with the basic shape of this Action, there will be options to implement
it in other ways. For example as a composite action calling Go programs to do all the work,
or calling Go programs to do the coordination of calling external programs,
or as a precompiled Docker image Action
(though that would present problems for Darwin builds which rely on macOS and CGO).

### TODO

- Add a reusable workflow for better optimisation (i.e. running in parallel jobs)
- Store build metadata for external systems to use to reproduce the build.
- See **ENGSRV-083** (internal only) for future plans.
