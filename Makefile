SHELL := /usr/bin/env bash -euo pipefail -c

test: test/bats

test/bats:
	# Running bats tests in scripts/
	@bats scripts/
