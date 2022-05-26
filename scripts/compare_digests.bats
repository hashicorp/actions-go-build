#!/usr/bin/env bats

set -Eeuo pipefail

load assertions.bats

setup() {
	cd "$BATS_TEST_TMPDIR"
	export PRIMARY_BUILD_ROOT="$(pwd)/primary"
	mkdir -p "$PRIMARY_BUILD_ROOT"
	export VERIFICATION_BUILD_ROOT="$(pwd)/verification"
	mkdir -p "$VERIFICATION_BUILD_ROOT"

	# TEST_FILE_CONTENTS and TEST_FILE_SHA256 need to be modified at the same time if at all.
	TEST_FILE_CONTENTS="Test file"
	TEST_FILE_SHA256="114811b0b8998cb9853a5379598021410feddf69bb2ee7b7145d052a7e9b5d45"

	export META_DIR="meta"

	setup_build_root_dir "$PRIMARY_BUILD_ROOT"
	setup_build_root_dir "$VERIFICATION_BUILD_ROOT"
}

setup_build_root_dir() {(
	cd "$1"
	mkdir -p "$META_DIR"
)}

enter_primary_root() {
	cd "$PRIMARY_BUILD_ROOT"
}

enter_verification_root() {
	cd "$VERIFICATION_BUILD_ROOT"
}

