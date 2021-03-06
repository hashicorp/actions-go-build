<!--
DO NOT EDIT THIS FILE MANUALLY; IT IS GENERATED BY 'make docs'
Instead, edit the files in dev/changes/, then run 'make docs' to update this file.
-->
# Changelog - Go Build Action

## Unreleased Changes (targeting v0.1.4)

### Fixed

### Added

### Changed

---
Internal Changes
  - Add `make version/set VERSION=<X.Y.Z>` to set the development version.
  - Changelog generation more robust.
  - When pre-push hook fails you now see the output so it's easier to debug.
  - Readme tidied up.
  - Converted digest inputs to Go.

## [v0.1.3](https://github.com/hashicorp/actions-go-build/releases/tag/v0.1.3) - June 15, 2022

- Adds .exe extension for windows binaries.
- Clearer logging when calculated default values are used.
- Internal:
  - Better handling of inputs with default values that also need to be manipulated or validated.
  - Better test coverage for input handling.
  - Target `make changelog/add` reminds about recent commits to help remember what's been done
    recently, regenerates the main `CHANGELOG.md`, and commits the result.
  - Release workflow to release via GitHub Actions.

## [v0.1.2](https://github.com/hashicorp/actions-go-build/releases/tag/v0.1.2) - June 13, 2022

- Default `product_name` to `repo_name`
- Automatically append `+ent` suffix for `-enteprise` products unless there's already
  any version metadata present.
- Fix broken tests.

### Development

- Added convenience script to set the current development version: `./dev/release/set_version`
- Added git pre-push hook to check that tests pass, all docs are up to date and more.

## [v0.1.1](https://github.com/hashicorp/actions-go-build/releases/tag/v0.1.1) - June 10, 2022

More conventional default zip name.
This means that e.g. actions-docker-build will guess the right
zip file name without explicit config.

## [v0.1.0](https://github.com/hashicorp/actions-go-build/releases/tag/v0.1.0) - June 10, 2022

### Fixed

- Uses the`-X` flag when zipping to exclude UID and GID info from the zip.
  This seems to make the zip file more likely to reproduce correctly.

### Improved

- Test cases now moved to their own reusable workflow which parameterises
  the runner. This means there are half as many test cases defined,
  and we just run the entire suite twice, once for linux and once for mac.
- Test cases are now ready to be run on our own self-hosted runners as well
  so we can exercise them in that environment.
- When there is a zip mismatch, we now dump detailed info about the zip file
  using `zipinfo` and we stat the product binary to aid with debugging.
- Logging now uses bold and coloured text to highlight major passages
  in the logs (bold blue) errors (bold red) and other important info
  (just bold).

## [v0.0.3](https://github.com/hashicorp/actions-go-build/releases/tag/v0.0.3) - June 08, 2022

- More graceful handling of installing coreutils on macOS. 
  Doesn't attempt to install again if GNU date already present.

## [v0.0.2](https://github.com/hashicorp/actions-go-build/releases/tag/v0.0.2) - June 08, 2022

- Better error message when env vars missing.

## [v0.0.1](https://github.com/hashicorp/actions-go-build/releases/tag/v0.0.1) - June 08, 2022

Initial version.

See the [README at this tag](https://github.com/hashicorp/actions-go-build/blob/v0.0.1/README.md) for details.

