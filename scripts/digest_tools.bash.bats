#!/usr/bin/env bats

set -Eeuo pipefail

load assertions.bats

setup() {
	source scripts/digest_tools.bash
	cd "$BATS_TEST_TMPDIR"
	export PRIMARY_ROOT_DIR="$(pwd)/primary"
	mkdir -p "$PRIMARY_ROOT_DIR"
	export LOCAL_VERIFICATION_ROOT_DIR="$(pwd)/verification"
	mkdir -p "$LOCAL_VERIFICATION_ROOT_DIR"

	# TEST_FILE_CONTENTS and TEST_FILE_SHA256 need to be modified at the same time if at all.
	TEST_FILE_CONTENTS="Test file"
	TEST_FILE_SHA256="114811b0b8998cb9853a5379598021410feddf69bb2ee7b7145d052a7e9b5d45"

	export META_DIR="meta"

	setup_build_root_dir "$PRIMARY_ROOT_DIR"
	setup_build_root_dir "$LOCAL_VERIFICATION_ROOT_DIR"
}

setup_build_root_dir() {(
	cd "$1"
	echo "$TEST_FILE_CONTENTS" > "testfile"
	mkdir -p "$META_DIR"
)}

enter_primary_root() {
	cd "$PRIMARY_ROOT_DIR"
}

enter_verification_root() {
	cd "$LOCAL_VERIFICATION_ROOT_DIR"
}

@test "write digest writes correct digest to expected path" {	
	enter_primary_root
	write_digest bin testfile
	assert_file_has_contents "meta/bin_digest" "$TEST_FILE_SHA256"
}

@test "read digest reads correct bin digest from primary build path" {	
	echo "deadbeef" > "$PRIMARY_ROOT_DIR/meta/bin_digest"
	GOT="$(read_digest primary bin)"
	[ "$GOT" = "deadbeef" ] || {
		echo "Read '$GOT'; want 'deadbeef'"
	}
}

@test "read digest reads correct bin digest from verification build path" {	
	echo "cabba9e" > "$LOCAL_VERIFICATION_ROOT_DIR/meta/bin_digest"
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
	echo "thesame" > "$PRIMARY_ROOT_DIR/meta/bin_digest"
	echo "thesame" > "$LOCAL_VERIFICATION_ROOT_DIR/meta/bin_digest"

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
	echo "thesame" > "$PRIMARY_ROOT_DIR/meta/zip_digest"
	echo "thesame" > "$LOCAL_VERIFICATION_ROOT_DIR/meta/zip_digest"

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
	echo "thesame" > "$PRIMARY_ROOT_DIR/meta/bin_digest"
	echo "different" > "$LOCAL_VERIFICATION_ROOT_DIR/meta/bin_digest"

	WANT="$(printf "thesame\ndifferent")"

	assert_failure_with_output "$WANT" compare_digest bin

}

@test "compare zip digest fails when digests are different" {
	echo "thesame" > "$PRIMARY_ROOT_DIR/meta/zip_digest"
	echo "different" > "$LOCAL_VERIFICATION_ROOT_DIR/meta/zip_digest"

	WANT="$(printf "thesame\ndifferent")"

	assert_failure_with_output "$WANT" compare_digest zip

}
