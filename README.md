# Reproducible Build Action

**This repo is WIP; not ready for use yet.**

The aim of this action is to allow defining a build in such a way that we are able
to attempt to repeat that build and compare the results to ensure that we get the
same built artifact(s) each time.

More documentation to follow as implementation progresses.

This is for internal HashiCorp use only; Internal folks please refer to RFC ENGSRV-084 for more details.

## What does it do?

Currently only supports pure Go projects.

1. **Primary Build:**
   a. Executes your build instructions in the default checkout directory.
   b. Zips and uploads the results as GitHub Actions artifacts using standard HashiCorp artifact names.
2. **Local Verification Build:**
   a. Executes your build instructions again in a different directory, at a later time.
   b. Zips and uploads the results as GitHub Actions artifacts labelled "local-verification-build".
3. **Compares Build Outputs:**
   a. Compares the SHA256 sums of your compiled binary artifacts from both builds.
   b. Compares the SHA256 sums of your zip file artifacts from both builds.
   c. Fails the build if either produce a mismatch, succeds otherwise.

## TODO

- Store build metadata for external systems to use to reproduce the build.
- Support non-Go projects.
- See ENGSRV-083 for future plans.
