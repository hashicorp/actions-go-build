set -Eeuo pipefail
log() { echo "==> $*" 1>&2; }
err() { log "ERROR: $*"; return 1; }
die() { log "FATAL: $*"; exit 1; }

# We rely on the GNU date program as it can convert the format of arbitrary dates.
# Replace 'date' with a function that routes to GNU date if needed.
date() {
	local PROG
	PROG="$(gnu_date_prog)"
	"$PROG" "$@"
}

# This function echoes either 'date' or 'gdate' if it's installed as one of those.
# It exits with an error if GNU date is not found.
gnu_date_prog() {
	local ERROR="GNU date not installed."
	[ "$(uname)" != "Darwin" ] || ERROR+=" On mac? Try 'brew install coreutils'"
	is_gnu_date date || is_gnu_date gdate || err "$ERROR"
}

# is_gnu_date fails with no output if the named program in the path is not GNU date.
# Otherwise it succeeds and prints the name of the program passed in.
is_gnu_date() {
	"$1" -d yesterday > /dev/null 2>&1 || return 1
	# Echo the binary path to avoid ambiguity with the shim function.
	which "$1"
}

# If GNU touch is installed as gtouch, use that rather than
# the standard touch, because on some systems the standard
# touch doesn't support the -d flag.
touch() {
	if command -v gtouch > /dev/null 2>&1; then
		gtouch "$@"
	else
		# Don't invoke the function again.
		"$(which touch)" "$@"
	fi
}
