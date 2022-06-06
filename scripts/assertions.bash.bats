#!/usr/bin/env bats

load assertions.bash

setup() {
	rm -rf "$BATS_TEST_TMPDIR"
	mkdir -p "$BATS_TEST_TMPDIR"
}

@test "validate assert_file_exists" {
	local FILE="$BATS_TEST_TMPDIR/a_file"
	echo "hi" > "$FILE"

	# OK
	assert_file_exists "$FILE"

	# Nonexistent.
	rm -rf "$FILE"
	if assert_file_exists "$FILE"; then
		return 1
	fi
}

@test "validate assert_executable_file_exists" {
	local FILE="$BATS_TEST_TMPDIR/a_file"
	echo "hi" > "$FILE"

	# OK
	chmod +x "$FILE"
	assert_executable_file_exists "$FILE"

	# Not executable.
	chmod -x "$FILE"
	if assert_executable_file_exists "$FILE"; then
		return 1
	fi

	# Nonexistent.
	rm "$FILE"
	if assert_executable_file_exists "$FILE"; then
		return 1
	fi
}

