group() {
	echo "::group::$*" 1>&2
}

endgroup() {
	echo "::endgroup::" 1>&2
}

group_cmd() {
	group "$*"
	trap endgroup RETURN
	"$@"
}
