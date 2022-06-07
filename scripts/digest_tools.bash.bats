#!/usr/bin/env bats

set -Eeuo pipefail

load assertions.bash

setup() {
	source scripts/digest_tools.bash
	cd "$BATS_TEST_TMPDIR"
	PRIMARY_BUILD_ROOT="$(pwd)/primary"
	export PRIMARY_BUILD_ROOT
	mkdir -p "$PRIMARY_BUILD_ROOT"
	VERIFICATION_BUILD_ROOT="$(pwd)/verification"
	export VERIFICATION_BUILD_ROOT
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
	echo "$TEST_FILE_CONTENTS" > "testfile"
	mkdir -p "$META_DIR"
)}

enter_primary_root() {
	cd "$PRIMARY_BUILD_ROOT"
}

enter_verification_root() {
	cd "$VERIFICATION_BUILD_ROOT"
}

@test "write digest writes correct digest to expected path" {
	enter_primary_root
	write_digest bin testfile
	assert_file_has_contents "meta/bin_digest" "$TEST_FILE_SHA256"
}

@test "read digest reads correct bin digest from primary build path" {
	echo "deadbeef" > "$PRIMARY_BUILD_ROOT/meta/bin_digest"
	assert_success_with_output "deadbeef" read_digest primary bin
}

@test "read digest reads correct bin digest from verification build path" {
	echo "cabba9e" > "$VERIFICATION_BUILD_ROOT/meta/bin_digest"
	GOT="$(read_digest verification bin)"
	[ "$GOT" = "cabba9e" ] || {
		echo "Read '$GOT'; want 'cabba9e'"
	}
}


@test "compare bin digest fails when digests don't exist" {
	if compare_digest bin; then
		echo "compare_digest succeeded but it should have failed"
		return 1
	fi
}

@test "compare zip digest fails when digests don't exist" {
	if compare_digest zip; then
		echo "compare_digest succeeded but it should have failed"
		return 1
	fi
}

@test "compare bin digest succeeds when digests the same" {
	echo "thesame" > "$PRIMARY_BUILD_ROOT/meta/bin_digest"
	echo "thesame" > "$VERIFICATION_BUILD_ROOT/meta/bin_digest"

	GOT="$(compare_digest bin)" || {
		echo "compare_digest failed but it should have succeeded"
		return 1
	}

	WANT="thesame"

	[ "$GOT" = "$WANT" ] || {
		echo "Got digest '$GOT'; want '$WANT'"
		return 1
	}

}

@test "compare zip digest succeeds when digests the same" {
	echo "thesame" > "$PRIMARY_BUILD_ROOT/meta/zip_digest"
	echo "thesame" > "$VERIFICATION_BUILD_ROOT/meta/zip_digest"

	GOT="$(compare_digest zip)" || {
		echo "compare_digest failed but it should have succeeded"
		return 1
	}

	WANT="thesame"

	[ "$GOT" = "$WANT" ] || {
		echo "Got digest '$GOT'; want '$WANT'"
		return 1
	}

}

@test "compare bin digest fails when digests are different" {
	echo "thesame" > "$PRIMARY_BUILD_ROOT/meta/bin_digest"
	echo "different" > "$VERIFICATION_BUILD_ROOT/meta/bin_digest"

	WANT="$(printf "thesame\ndifferent")"

	assert_failure_with_output "$WANT" compare_digest bin

}

@test "compare zip digest fails when digests are different" {
	echo "thesame" > "$PRIMARY_BUILD_ROOT/meta/zip_digest"
	echo "different" > "$VERIFICATION_BUILD_ROOT/meta/zip_digest"

	WANT="$(printf "thesame\ndifferent")"

	assert_failure_with_output "$WANT" compare_digest zip

}
