# Assertion functions

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

dump_got_want() {
	echo "Got output:"
	echo "$1"
	echo "Want output:"
	echo "$2"
}

assert_success_with_output() {
	local WANT="$1"
	shift
	if ! GOT="$("$@")"; then
		echo "Command failed but was expected to pass: $*"
		return 1
	fi
	[ "$GOT" = "$WANT" ] || {
		echo "Command succeeded but gave the wrong output: $*"
		echo "Got output:"
		echo "$GOT"
		echo "Want output:"
		echo "$WANT"
		return 1
	}
}

assert_failure_with_output() {
	local WANT="$1"
	shift
	if GOT="$("$@")"; then
		echo "Command succeeded but was expected to fail: $*"
		return 1
	fi
	[ "$GOT" = "$WANT" ] || {
		echo "Command failed correctly, but gave the wrong output: $*"
		dump_got_want "$GOT" "$WANT"
		return 1
	}
}

assert_exported_in_github_env() {
	VAR_NAME="$1"
	WANT="$2"

	GOT="$(get_value_from_github_env "$VAR_NAME")"

	if ! [ "$GOT" = "$WANT" ]; then
		echo "Got $VAR_NAME='$GOT'; want $VAR_NAME='$WANT'"
		return 1
	fi
}

assert_nonempty_in_github_env() {
	VAR_NAME="$1"
	GOT="$(get_value_from_github_env "$VAR_NAME")"
	[ -n "$GOT" ] || {
		echo "$VAR_NAME is empty; wanted non-empty."
		return 1
	}
}

get_value_from_github_env() {
	VAR_NAME="$1"
	GOT="$(
		unset "$VAR_NAME"
		# shellcheck disable=SC1090 # This is a non-constant source by definition.
		source "$GITHUB_ENV.export" && echo "${!VAR_NAME}"
	)"
	echo "$GOT"
}
