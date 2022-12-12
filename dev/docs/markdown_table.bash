# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Bash library

markdown_table_rows() {
	column -ts '|' -o ' | ' | sed -E -e 's/^ //g' -e 's/ $//g'
}

markdown_table() {

	HEADER_FUNC="$1"
	BODY_FUNC="$2"
	{ "$HEADER_FUNC"; "$BODY_FUNC"; } | markdown_table_rows
}

# Get GNU column program.
COLUMN="column"
if [ "$(uname)" = "Darwin" ]; then
	COLUMN="/opt/homebrew/opt/util-linux/bin/column"
	if [ ! -x "$COLUMN" ]; then
		COLUMN="/usr/local/opt/util-linux/bin/column"
		if [ ! -x "$COLUMN" ]; then
			die "Missing dependency; please install util-linux, e.g.: 'brew install util-linux'"
		fi
	fi
fi

column() {
	$COLUMN "$@" || return 1
}

write_header() {
	write_row "${HEADERS[@]}"
	for _ in "${HEADERS[@]}"; do
		printf "| ----- "
	done
	printf "|\n"
}

write_row() {
	for V in "$@"; do
		printf "| %s " "$V"
	done
	printf "|\n"
}

read_output_to_array() {
	local NAME="$1"; shift
	IFS=$'\n' read -r -d '' -a "$NAME" < <("$@" && printf '\0')
}
