assert_file_exists() {
	[ -f "$1" ] || {
		echo "File '$1' does not exist, but it should."
		return 1
	}
}

assert_executable_file_exists() {
	assert_file_exists "$1"
	[ -x "$1" ] || {
		echo "File '$1' is not executable, but it should."
		return 1
	}
}

assert_file_has_contents() {
	local FILE="$1"
	local CONTENTS="$2"
	assert_file_exists "$FILE"
	local GOT
	GOT="$(cat "$FILE")"
	[ "$GOT" = "$CONTENTS" ] || {
		echo "File '$FILE' has unexpected contents."
		echo "Got:"
		cat "$FILE"
		echo "Want:"
		echo "$CONTENTS"
		return 1
	}
}
