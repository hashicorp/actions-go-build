# Reproducible Build Action

**This repo is WIP; not ready for use yet.**

The aim of this action is to allow defining a build in such a way that we are able
to attempt to repeat that build and compare the results to ensure that we get the
same built artifact(s) each time.

More documentation to follow as implementation progresses.

This is for internal HashiCorp use only; Internal folks please refer to RFC ENGSRV-084 for more details.

## What does it do?

Currently only supports pure Go projects.

1. **Installs Specified Go version**
1. **Primary Build:**
	1. Executes your build instructions in the default checkout directory.
	1. Zips and uploads the results as GitHub Actions artifacts using standard HashiCorp artifact names.
1. **Local Verification Build:**
	1. Executes your build instructions again in a different directory, at a later time.
	1. Zips and uploads the results as GitHub Actions artifacts labelled "local-verification-build".
1. **Compares Build Outputs:**
	1. Compares the SHA256 sums of your compiled binary artifacts from both builds.
	1. Compares the SHA256 sums of your zip file artifacts from both builds.
	1. Fails the build if either produce a mismatch, succeds otherwise.

## Usage

This Action can run on both Ubuntu and macOS runners.

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
          go_version: 1.18
          product_name: example-app
          product_version: 1.2.3
          os: linux
          arch: amd64
          instructions: |-
            cd ./testdata/example-app
            go build \n              -trimpath \n              -buildvcs=false \n              -o "$BIN_PATH" \n              -ldflags "
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
|  **`instructions`**&nbsp;_(required)_     |  Build instructions to generate the binary. See [Build Instructions](#build-instructions) for more info.  |
|  **`go_version`**&nbsp;_(required)_       |  Version of Go to use for this build.                                                                     |
|  **`product_version`**&nbsp;_(required)_  |  Version of the product being built.                                                                      |
|  **`os`**&nbsp;_(required)_               |  Target product operating system.                                                                         |
|  **`arch`**&nbsp;_(required)_             |  Target product architecture.                                                                             |
|  **`product_name`**&nbsp;_(required)_     |  Name of the product to build. Used to calculate default `bin_name` and `zip_name`.                       |
|  `bin_name`&nbsp;_(optional)_             |  Name of the product binary generated. Defaults to `product_name` minus any `-enterprise` suffix.         |
|  `zip_name`&nbsp;_(optional)_             |  Name of the product zip file. Defaults to `<product_name>_<product_version>_<os>_<arch>.zip`.            |
<!-- end:insert:scripts/codegen/inputs_doc -->

### Build Instuctions

The main input is [`build_instructions`](#build-instructions) which defines the build.
Build instructions is a bash script. It should be kept as simple as possible.

When the `build_instructions` are executed, there are a set of environment variables
already exported that you can make use of in your instructions,
(see [Build Environment](#build-environment), below).

At a minimum, the script must use the environment variable `$BIN_PATH`
because the minimal thing it can do is to write the compiled binary to `$BIN_PATH`.
In order to add other files like licenses etc to the zip file, you need to
write them into `$TARGET_DIR` in your build instructions.

### Build Environment

The following variables are exported when `build_instructions` are executed,
so you can use them in your script and they're available to programs your
script calls.

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

## TODO

- Store build metadata for external systems to use to reproduce the build.
- Support non-Go projects.
- See ENGSRV-083 for future plans.
