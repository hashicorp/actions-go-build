#!/usr/bin/env bats

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

non_ent_exported_assertions() {
	# Assert default vars generated correctly.
	assert_exported_in_github_env GOOS                    "darwin"
	assert_exported_in_github_env GOARCH                  "amd64"
	assert_exported_in_github_env TARGET_DIR              "dist"
	assert_exported_in_github_env ZIP_DIR                 "out"
	assert_exported_in_github_env META_DIR                ".meta"
	assert_exported_in_github_env PRIMARY_BUILD_ROOT      "$(pwd)"
	assert_exported_in_github_env VERIFICATION_BUILD_ROOT "$(dirname "$PWD")/verification"
	assert_exported_in_github_env BIN_NAME                "blargle"
	assert_exported_in_github_env ZIP_NAME                "blargle_1.2.3_darwin_amd64.zip"
	assert_exported_in_github_env PRODUCT_REVISION        "$(git rev-parse HEAD)"
	assert_exported_in_github_env BIN_PATH                "dist/blargle"
	assert_exported_in_github_env ZIP_PATH                "out/blargle_1.2.3_darwin_amd64.zip"

	assert_nonempty_in_github_env PRODUCT_REVISION_TIME
}

@test "default vars calculated correctly - non-enterprise" {
	# Setup.
	set_required_env_vars

	# Run the script under test.
	./scripts/digest_inputs

	non_ent_exported_assertions
}

@test "default vars calculated correctly - non-enterprise - windows" {
	# Setup.
	set_required_env_vars
	OS=windows

	# Run the script under test.
	./scripts/digest_inputs

	assert_exported_in_github_env BIN_NAME "blargle.exe"
}

@test "default vars calculated correctly - non-enterprise - no product name" {
	# Setup.
	set_required_env_vars

	# Run the script under test.
	./scripts/digest_inputs

	non_ent_exported_assertions
}

ent_exported_assertions() {
	# Assert default vars generated correctly.
	assert_exported_in_github_env GOOS                    "darwin"
	assert_exported_in_github_env GOARCH                  "amd64"
	assert_exported_in_github_env TARGET_DIR              "dist"
	assert_exported_in_github_env ZIP_DIR                 "out"
	assert_exported_in_github_env META_DIR                ".meta"
	assert_exported_in_github_env PRIMARY_BUILD_ROOT      "$(pwd)"
	assert_exported_in_github_env VERIFICATION_BUILD_ROOT "$(dirname "$PWD")/verification"
	assert_exported_in_github_env BIN_NAME                "blargle"
	assert_exported_in_github_env ZIP_NAME                "blargle_1.2.3+ent_darwin_amd64.zip"
	assert_exported_in_github_env PRODUCT_REVISION        "$(git rev-parse HEAD)"
	assert_exported_in_github_env BIN_PATH                "dist/blargle"
	assert_exported_in_github_env ZIP_PATH                "out/blargle_1.2.3+ent_darwin_amd64.zip"
	assert_nonempty_in_github_env PRODUCT_REVISION_TIME
}

@test "default vars calculated correctly - enterprise - with product name" {
	# Setup.
	set_required_env_vars
	PRODUCT_REPOSITORY="blargle-enterprise"
	PRODUCT_NAME="blargle-enterprise"

	# Run the script under test.
	./scripts/digest_inputs

	ent_exported_assertions
}

@test "default vars calculated correctly - enterprise - no product name" {
	# Setup.
	set_required_env_vars
	PRODUCT_REPOSITORY="someorg/blargle-enterprise"

	# Run the script under test.
	./scripts/digest_inputs

	ent_exported_assertions
}

@test "default vars calculated correctly - enterprise - no product name - repo no org" {
	# Setup.
	set_required_env_vars
	PRODUCT_REPOSITORY="blargle-enterprise"

	# Run the script under test.
	./scripts/digest_inputs

	ent_exported_assertions
}

@test "default vars calculated correctly - enterprise - windows - default bin name" {
	# Setup.
	set_required_env_vars
	PRODUCT_REPOSITORY="blargle-enterprise"
	OS=windows

	# Run the script under test.
	./scripts/digest_inputs

	assert_exported_in_github_env BIN_NAME "blargle.exe"
	assert_exported_in_github_env ZIP_NAME "blargle_1.2.3+ent_windows_amd64.zip"
}

@test "default vars calculated correctly - enterprise - windows - overridden bin name" {
	# Setup.
	set_required_env_vars
	PRODUCT_REPOSITORY="blargle-enterprise"
	OS=windows
	BIN_NAME=bugler

	# Run the script under test.
	./scripts/digest_inputs

	assert_exported_in_github_env BIN_NAME "bugler.exe"
	assert_exported_in_github_env ZIP_NAME "blargle_1.2.3+ent_windows_amd64.zip"
}

@test "default vars calculated correctly - enterprise - windows - with .exe already" {
	# Setup.
	set_required_env_vars
	PRODUCT_REPOSITORY="blargle-enterprise"
	OS=windows
	BIN_NAME=bugler.exe

	# Run the script under test.
	./scripts/digest_inputs

	assert_exported_in_github_env BIN_NAME "bugler.exe"
	assert_exported_in_github_env ZIP_NAME "blargle_1.2.3+ent_windows_amd64.zip"
}
