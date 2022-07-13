#!/usr/bin/env bats

set -Eeuo pipefail

load assertions.bash

setup() {
	cp -r testdata/example-app/* "$BATS_TEST_TMPDIR"
	cd "$BATS_TEST_TMPDIR"
	git init
	git add .
	git commit -m "initial commit"
	# Set the OS and arch to produce a binary that will execute on this platform.
	# You need to have Go installed locally anyway to run the test, so using go env
	# to determine the platform.
	OS="$(go env GOOS)"
	ARCH="$(go env GOARCH)"
	export OS
	export ARCH
	export GOOS="$OS"
	export GOARCH="$ARCH"
	export PRODUCT_REPOSITORY="dadgarcorp/blargles"
	export PRODUCT_NAME="blargles"
	export PRODUCT_VERSION="1.2.3"
	export PRODUCT_REVISION="cabba9e"
	export PRODUCT_REVISION_TIME="2006-02-02T22:00:01+00:00"
	export BIN_NAME="blargles"
	export ZIP_NAME="blargles.zip"
}

build() {
	actions-go-build primary
}

@test "working build instructions are executed correctly" {
	# shellcheck disable=SC2016 # We don't want to expand the vars in instructions yet.
	export INSTRUCTIONS='
		go build -o $BIN_PATH .
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
	assert_file_exists "out/blargles.zip"
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
