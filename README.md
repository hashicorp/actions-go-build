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

### Inputs

<!-- insert:scripts/codegen/inputs_doc -->
<!-- end:insert:scripts/codegen/inputs_doc -->

## TODO

- Store build metadata for external systems to use to reproduce the build.
- Support non-Go projects.
- See ENGSRV-083 for future plans.
