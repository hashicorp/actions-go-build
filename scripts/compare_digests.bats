#!/usr/bin/env bats

set -Eeuo pipefail

load assertions.bash

setup() {
	export PRIMARY_BUILD_ROOT="$BATS_TEST_TMPDIR/primary"
	mkdir -p "$PRIMARY_BUILD_ROOT"
	export VERIFICATION_BUILD_ROOT="$BATS_TEST_TMPDIR/verification"
	mkdir -p "$VERIFICATION_BUILD_ROOT"

	# TEST_FILE_CONTENTS and TEST_FILE_SHA256 need to be modified at the same time if at all.
	export TEST_FILE_CONTENTS="Test file"
	export TEST_FILE_SHA256="114811b0b8998cb9853a5379598021410feddf69bb2ee7b7145d052a7e9b5d45"

	export META_DIR="meta"

	setup_build_root_dir "$PRIMARY_BUILD_ROOT"
	setup_build_root_dir "$VERIFICATION_BUILD_ROOT"

	# These vars need to be set for logging purposes in compare_digests.
	export PRODUCT_NAME="some-product"
	export PRODUCT_VERSION="1.2.3"
	export OS="darwin"
	export ARCH="arm64"

	export GITHUB_STEP_SUMMARY="$BATS_TEST_TMPDIR/github-step-summary.md"
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

@test "success if all digests match" {
	(
		enter_primary_root
		echo "blah" > meta/bin_digest
		echo "blah" > meta/zip_digest
	)

	(
		enter_verification_root
		echo "blah" > meta/bin_digest
		echo "blah" > meta/zip_digest
	)

	assert_success_with_output "" ./scripts/compare_digests
}

@test "failure if bin digest no match" {
	(
		enter_primary_root
		echo "blah" > meta/bin_digest
		echo "blah" > meta/zip_digest
	)

	(
		enter_verification_root
		echo "no" > meta/bin_digest
		echo "blah" > meta/zip_digest
	)

	assert_failure_with_output "" ./scripts/compare_digests
}

@test "failure if zip digest no match" {
	(
		enter_primary_root
		echo "blah" > meta/bin_digest
		echo "blah" > meta/zip_digest
	)

	(
		enter_verification_root
		echo "blah" > meta/bin_digest
		echo "no" > meta/zip_digest
	)

	assert_failure_with_output "" ./scripts/compare_digests
}

@test "failure if both digests no match" {
	(
		enter_primary_root
		echo "blah" > meta/bin_digest
		echo "blah" > meta/zip_digest
	)

	(
		enter_verification_root
		echo "no" > meta/bin_digest
		echo "no" > meta/zip_digest
	)

	assert_failure_with_output "" ./scripts/compare_digests
}

@test "failure if primary bin digest missing" {
	(
		enter_primary_root
		echo "blah" > meta/zip_digest
	)

	(
		enter_verification_root
		echo "blah" > meta/bin_digest
		echo "blah" > meta/zip_digest
	)

	assert_failure_with_output "" ./scripts/compare_digests
}

@test "failure if verification bin digest missing" {
	(
		enter_primary_root
		echo "blah" > meta/bin_digest
		echo "blah" > meta/zip_digest
	)

	(
		enter_verification_root
		echo "blah" > meta/zip_digest
	)

	assert_failure_with_output "" ./scripts/compare_digests
}

@test "failure if primary zip digest missing" {
	(
		enter_primary_root
		echo "blah" > meta/bin_digest
	)

	(
		enter_verification_root
		echo "blah" > meta/bin_digest
		echo "blah" > meta/zip_digest
	)

	assert_failure_with_output "" ./scripts/compare_digests
}

@test "failure if verification zip digest missing" {
	(
		enter_primary_root
		echo "blah" > meta/bin_digest
		echo "blah" > meta/zip_digest
	)

	(
		enter_verification_root
		echo "blah" > meta/bin_digest
	)

	assert_failure_with_output "" ./scripts/compare_digests
}
