set -Eeuo pipefail

trap 'exit 1' ERR

log() { echo "==> $*" 1>&2; }
err() { log "$(bold_red "ERROR:") $(bold "$*")"; return 1; }
die() { log "$(bold_red "FATAL:") $(bold "$*")"; exit 1; }

styled_text() { ATTR="$1"; shift; echo -en '\033['"${ATTR}m$*"'\033[0m'; }

bold()      { styled_text "1"    "$*"; }
blue()      { styled_text "94"   "$*"; }
bold_blue() { styled_text "1;94" "$*"; }
red()       { styled_text "91"   "$*"; }
bold_red()  { styled_text "1;91" "$*"; }

log_bold() { log "$(bold_blue "$*")"; }

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
	local BINARY_PATH
	BINARY_PATH="$(which "$1")" || return 1
	"$BINARY_PATH" -d yesterday > /dev/null 2>&1 || return 1
	which "$BINARY_PATH"
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
