#!/usr/bin/env bats

set -Eeuo pipefail

load assertions.bats

setup() {
	source scripts/build.bash
	cp -r testdata/example-app/* "$BATS_TEST_TMPDIR"
	cd "$BATS_TEST_TMPDIR"
	# Set the OS and arch to produce a binary that will execute on this platform.
	# You need to have Go installed locally anyway to run the test, so using go env
	# to determine the platform.
	export OS="$(go env GOOS)"
	export ARCH="$(go env GOARCH)"
	export GOOS="$OS"
	export GOARCH="$ARCH"
	export PRODUCT_NAME="blargles"
	export PRODUCT_VERSION="1.2.3"
	export PRODUCT_REVISION="cabba9e"
	export TARGET_DIR="dist"
	export BIN_NAME="blargles"
	export BIN_PATH="$TARGET_DIR/$BIN_NAME"
	export ZIP_DIR="zip"
	export ZIP_NAME="blargles.zip"
	export ZIP_PATH="$ZIP_DIR/$ZIP_NAME"
	export META_DIR="meta"

	export PRODUCT_REVISION_TIME="2006-02-02T22:00:01+00:00"
}

@test "working build instructions are executed correctly" {
	export INSTRUCTIONS='
		go build -o $TARGET_DIR/$BIN_NAME .
	'
	# Run the build function.
	build || {
		echo "Build failed but it shouldn't have."
		return 1
	}

	# Assert the thing got built.
	assert_executable_file_exists "dist/blargles"

	# Run the file.
	./dist/blargles

	# Assert the zip was created.
	assert_file_exists "zip/blargles.zip"
}

@test "failing build instructions result in failure" {
	export INSTRUCTIONS='
		echo "On no!"
		exit 1
		echo "WAT!"
	'

	# Run the build function.
	if build; then
		echo "Build succeeded but it should have failed."
		return 1
	fi
}
