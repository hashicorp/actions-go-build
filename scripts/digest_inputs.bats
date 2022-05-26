#!/usr/bin/env bats

set -Eeuo pipefail

load assertions.bats

# setup ensures that there's a fresh .tmp directory, gitignored,
# and sets the GITHUB_ENV variable to a file path inside that directory.
setup() {
	rm -rf ./.tmp
	export GITHUB_ENV=./.tmp/github.env
	mkdir -p ./.tmp
	echo "*" > ./.tmp/.gitignore
}

set_required_env_vars() {
	export PACKAGE_NAME="blargle"
	export INSTRUCTIONS="
		Some
		multi-line
		build instructions
	"
	export OS="darwin"
	export ARCH="amd64"
	export PRODUCT_VERSION="1.2.3"
}

@test "required vars passed through unchanged" {
	# Setup.
	set_required_env_vars

	# Run the script under test.
	./scripts/digest_inputs

	# Assert required vars passed through unchanged.
	assert_exported_in_github_env PACKAGE_NAME    "blargle"
	assert_exported_in_github_env OS              "darwin"
	assert_exported_in_github_env ARCH            "amd64"
	assert_exported_in_github_env PRODUCT_VERSION "1.2.3"
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

@test "default vars calculated correctly - non-enterprise" {
	# Setup.
	set_required_env_vars

	# Run the script under test.
	./scripts/digest_inputs

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
	assert_nonempty_in_github_env PRODUCT_REVISION_TIME_LOCAL
}

@test "default vars calculated correctly - enterprise" {
	# Setup.
	set_required_env_vars
	export PACKAGE_NAME="blargle-enterprise"

	# Run the script under test.
	./scripts/digest_inputs

	# Assert default vars generated correctly.
	assert_exported_in_github_env GOOS                    "darwin"
	assert_exported_in_github_env GOARCH                  "amd64"
	assert_exported_in_github_env TARGET_DIR              "dist"
	assert_exported_in_github_env ZIP_DIR                 "out"
	assert_exported_in_github_env META_DIR                ".meta"
	assert_exported_in_github_env PRIMARY_BUILD_ROOT      "$(pwd)"
	assert_exported_in_github_env VERIFICATION_BUILD_ROOT "$(dirname "$PWD")/verification"
	assert_exported_in_github_env BIN_NAME                "blargle"
	assert_exported_in_github_env ZIP_NAME                "blargle-enterprise_1.2.3_darwin_amd64.zip"
	assert_exported_in_github_env PRODUCT_REVISION        "$(git rev-parse HEAD)"
	assert_exported_in_github_env BIN_PATH                "dist/blargle"
	assert_exported_in_github_env ZIP_PATH                "out/blargle-enterprise_1.2.3_darwin_amd64.zip"
}
