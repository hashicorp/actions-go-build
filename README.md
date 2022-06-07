# Go Build Action

_**Build and package a (reproducible) Go binary.**_

- **Define** the build.
- **Assert** that it is reproducible (optionally).
- **Use** the resultant artifacts in your workflow.

_This is intended for internal HashiCorp use only; Internal folks please refer to RFC ENGSRV-084 for more details._

<!-- insert:scripts/codegen/table_of_contents -->
## Table of Contents
* [Table of Contents](#table-of-contents)
* [Features](#features)
* [Usage](#usage)
  * [Example Workflow](#example-workflow)
  * [Inputs](#inputs)
  * [Build Environment](#build-environment)
    * [Environment Variables](#environment-variables)
  * [Build Instructions](#build-instructions)
    * [Example Build Instructions](#example-build-instructions)
* [TODO](#todo)
<!-- end:insert:scripts/codegen/table_of_contents -->

## Features

- **Results are zipped** using standard HashiCorp naming conventions.
- **You can include additional files** in the zip like licenses etc.
- **Convention over configuration** means minimal config required.
- **Reproducibility** is checked at build time.
- **Fast feedback** if accidental nondeterminism is introduced.

## Usage

This Action can run on both Ubuntu and macOS runners.

### Example Workflow

Example usage ([see this workflow running here](https://github.com/hashicorp/actions-reproducible-build/actions/workflows/example.yml)).

<!-- insert:scripts/codegen/print_example_workflow -->
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
            go build \
              -trimpath \
              -buildvcs=false \
              -o "$BIN_PATH" \
              -ldflags "
                -X 'main.Version=$PRODUCT_VERSION'
                -X 'main.Revision=$PRODUCT_REVISION'
                -X 'main.RevisionTime=$PRODUCT_REVISION_TIME'
              "
```
<!-- end:insert:scripts/codegen/print_example_workflow -->

### Inputs

<!-- insert:scripts/codegen/inputs_doc -->
|  Name                                     |  Description                                                                                              |
|  -----                                    |  -----                                                                                                    |
|  **`product_name`**&nbsp;_(required)_     |  Name of the product to build. Used to calculate default `bin_name` and `zip_name`.                       |
|  **`product_version`**&nbsp;_(required)_  |  Version of the product being built.                                                                      |
|  **`go_version`**&nbsp;_(required)_       |  Version of Go to use for this build.                                                                     |
|  **`os`**&nbsp;_(required)_               |  Target product operating system.                                                                         |
|  **`arch`**&nbsp;_(required)_             |  Target product architecture.                                                                             |
|  `bin_name`&nbsp;_(optional)_             |  Name of the product binary generated. Defaults to `product_name` minus any `-enterprise` suffix.         |
|  `zip_name`&nbsp;_(optional)_             |  Name of the product zip file. Defaults to `<product_name>_<product_version>_<os>_<arch>.zip`.            |
|  **`instructions`**&nbsp;_(required)_     |  Build instructions to generate the binary. See [Build Instructions](#build-instructions) for more info.  |
<!-- end:insert:scripts/codegen/inputs_doc -->

### Build Environment

When the `instructions` are executed, there are a set of environment variables
already exported that you can make use of
(see [Environment Variables](#environment-variables below).

#### Environment Variables

<!-- insert:scripts/codegen/environment_doc -->
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
<!-- end:insert:scripts/codegen/environment_doc -->

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

See also the [example workflow](#example-workflow) above, which injects info into the binary using `-ldflags`.

## TODO

- Add a reusable workflow for better optimisation (i.e. running in parallel jobs)
- Store build metadata for external systems to use to reproduce the build.
- See ENGSRV-083 for future plans.
