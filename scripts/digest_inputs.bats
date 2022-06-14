#!/usr/bin/env bats

# shellcheck disable=SC2030,SC2031 # We want the isolation these are warning about.

set -Eeuo pipefail

load assertions.bash

# setup ensures that there's a fresh .tmp directory, gitignored,
# and sets the GITHUB_ENV variable to a file path inside that directory.
setup() {
	export GITHUB_ENV="$BATS_TEST_TMPDIR/github.env"
	rm -rf "$(dirname "$GITHUB_ENV")"
	mkdir -p "$(dirname "$GITHUB_ENV")"
}

set_required_env_vars() {
	export PRODUCT_REPOSITORY="dadgarcorp/blargle"
	export OS="darwin"
	export ARCH="amd64"
	export PRODUCT_VERSION="1.2.3"
	export REPRODUCIBLE="assert"
	export INSTRUCTIONS="
		Some
		multi-line
		build instructions
	"

	# Export non-required env vars empty. This means
	# we can set them in tests without having to
	# remember to export them (causing shellcheck
	# to complain.
	export PRODUCT_NAME BIN_NAME ZIP_NAME
}

@test "required vars passed through unchanged" {
	# Setup.
	set_required_env_vars

	# Run the script under test.
	./scripts/digest_inputs

	# Assert required vars passed through unchanged.
	assert_exported_in_github_env PRODUCT_NAME    "blargle"
	assert_exported_in_github_env OS              "darwin"
	assert_exported_in_github_env ARCH            "amd64"
	assert_exported_in_github_env PRODUCT_VERSION "1.2.3"
	assert_exported_in_github_env REPRODUCIBLE    "assert"
	assert_exported_in_github_env INSTRUCTIONS    "
		Some
		multi-line
		build instructions
	"
}

@test "non-required vars handled correctly" {
	# Setup.
	set_required_env_vars

	export BIN_NAME="somethingelse"
	export ZIP_NAME="somethingelse.zip"

	# Run the script under test.
	./scripts/digest_inputs

	# Assert non-required env vars handled correctly.
	assert_exported_in_github_env BIN_NAME "somethingelse"
	assert_exported_in_github_env ZIP_NAME "somethingelse.zip"
	assert_exported_in_github_env ZIP_PATH "out/somethingelse.zip"
	assert_exported_in_github_env BIN_PATH "dist/somethingelse"
}

# Assert on the whole environment.
# Set 'WANT_<NAME>' before calling this to override the expected value
# for a given variable.
assert_exported_vars() {
	assert_exported_in_github_env PRODUCT_NAME            "${WANT_PRODUCT_NAME:-blargle}"
	assert_exported_in_github_env PRODUCT_VERSION         "${WANT_PRODUCT_VERSION:-1.2.3}"
	assert_exported_in_github_env REPRODUCIBLE            "${WANT_REPRODUCIBLE:-assert}"
	assert_exported_in_github_env OS                      "${WANT_OS:-darwin}"
	assert_exported_in_github_env ARCH                    "${WANT_ARCH:-amd64}"
	assert_exported_in_github_env GOOS                    "${WANT_GOOS:-darwin}"
	assert_exported_in_github_env GOARCH                  "${WANT_GOARCH:-amd64}"
	assert_exported_in_github_env TARGET_DIR              "${WANT_TARGET_DIR:-dist}"
	assert_exported_in_github_env ZIP_DIR                 "${WANT_ZIP_DIR:-out}"
	assert_exported_in_github_env META_DIR                "${WANT_META_DIR:-.meta}"
	assert_exported_in_github_env BIN_NAME                "${WANT_BIN_NAME:-blargle}"
	assert_exported_in_github_env ZIP_NAME                "${WANT_ZIP_NAME:-blargle_1.2.3_darwin_amd64.zip}"
	assert_exported_in_github_env BIN_PATH                "${WANT_BIN_PATH:-dist/blargle}"
	assert_exported_in_github_env ZIP_PATH                "${WANT_ZIP_PATH:-out/blargle_1.2.3_darwin_amd64.zip}"

	assert_exported_in_github_env PRIMARY_BUILD_ROOT      "$(pwd)"
	assert_exported_in_github_env VERIFICATION_BUILD_ROOT "$(dirname "$PWD")/verification"
	assert_exported_in_github_env PRODUCT_REVISION        "$(git rev-parse HEAD)"
	assert_nonempty_in_github_env PRODUCT_REVISION_TIME
}

# Assert on the whole environment using default "want" values for enterprise.
assert_exported_vars_ent() {
	WANT_ZIP_NAME="${WANT_ZIP_NAME:-blargle_1.2.3+ent_darwin_amd64.zip}"
	WANT_ZIP_PATH="${WANT_ZIP_PATH:-out/blargle_1.2.3+ent_darwin_amd64.zip}"
	WANT_PRODUCT_NAME="${WANT_PRODUCT_NAME:-blargle-enterprise}"
	WANT_PRODUCT_VERSION="${WANT_PRODUCT_VERSION:-1.2.3+ent}"

	assert_exported_vars
}

@test "default vars calculated correctly - non-enterprise" {
	# Setup.
	set_required_env_vars

	# Run the script under test.
	./scripts/digest_inputs

	assert_exported_vars
}

@test "default vars calculated correctly - non-enterprise - windows" {
	# Setup.
	set_required_env_vars
	OS=windows

	# Run the script under test.
	./scripts/digest_inputs

	WANT_OS="windows"
	WANT_GOOS="windows"
	WANT_BIN_NAME="blargle.exe"
	WANT_BIN_PATH="dist/blargle.exe"
	WANT_ZIP_NAME="blargle_1.2.3_windows_amd64.zip"
	WANT_ZIP_PATH="out/blargle_1.2.3_windows_amd64.zip"

	assert_exported_vars
}

@test "default vars calculated correctly - non-enterprise - no product name" {
	# Setup.
	set_required_env_vars

	# Run the script under test.
	./scripts/digest_inputs

	assert_exported_vars
}

@test "default vars calculated correctly - enterprise - with product name" {
	# Setup.
	set_required_env_vars
	PRODUCT_REPOSITORY="blargle-enterprise"
	PRODUCT_NAME="blargle-enterprise"

	# Run the script under test.
	./scripts/digest_inputs

	assert_exported_vars_ent
}

@test "default vars calculated correctly - enterprise - no product name" {
	# Setup.
	set_required_env_vars
	PRODUCT_REPOSITORY="someorg/blargle-enterprise"

	# Run the script under test.
	./scripts/digest_inputs

	assert_exported_vars_ent
}

@test "default vars calculated correctly - enterprise - no product name - repo no org" {
	# Setup.
	set_required_env_vars
	PRODUCT_REPOSITORY="blargle-enterprise"

	# Run the script under test.
	./scripts/digest_inputs

	assert_exported_vars_ent
}

@test "default vars calculated correctly - enterprise - windows - default bin name" {
	# Setup.
	set_required_env_vars
	PRODUCT_REPOSITORY="blargle-enterprise"
	OS=windows

	# Run the script under test.
	./scripts/digest_inputs

	WANT_OS="windows"
	WANT_GOOS="windows"
	WANT_BIN_NAME="blargle.exe"
	WANT_BIN_PATH="dist/blargle.exe"
	WANT_ZIP_NAME="blargle_1.2.3+ent_windows_amd64.zip"
	WANT_ZIP_PATH="out/blargle_1.2.3+ent_windows_amd64.zip"

	assert_exported_vars_ent
}

@test "default vars calculated correctly - enterprise - windows - overridden bin name" {
	# Setup.
	set_required_env_vars
	PRODUCT_REPOSITORY="blargle-enterprise"
	OS=windows
	BIN_NAME=bugler

	# Run the script under test.
	./scripts/digest_inputs

	WANT_OS="windows"
	WANT_GOOS="windows"
	WANT_BIN_NAME="bugler.exe"
	WANT_BIN_PATH="dist/bugler.exe"
	WANT_ZIP_NAME="blargle_1.2.3+ent_windows_amd64.zip"
	WANT_ZIP_PATH="out/blargle_1.2.3+ent_windows_amd64.zip"

	assert_exported_vars_ent
}

@test "default vars calculated correctly - enterprise - windows - with .exe already" {
	# Setup.
	set_required_env_vars
	PRODUCT_REPOSITORY="blargle-enterprise"
	OS=windows
	BIN_NAME=bugler.exe

	# Run the script under test.
	./scripts/digest_inputs

	WANT_OS="windows"
	WANT_GOOS="windows"
	WANT_BIN_NAME="bugler.exe"
	WANT_BIN_PATH="dist/bugler.exe"
	WANT_ZIP_NAME="blargle_1.2.3+ent_windows_amd64.zip"
	WANT_ZIP_PATH="out/blargle_1.2.3+ent_windows_amd64.zip"

	assert_exported_vars_ent
}
