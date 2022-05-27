SHELL := /usr/bin/env bash -euo pipefail -c

default: test

# tools/mac tries to install dependencies on mac using homebrew.
tools/mac:
	brew install coreutils util-linux

test: test/bats

test/bats:
	# Running bats tests in scripts/
	@bats scripts/

docs:
	@./scripts/codegen/update_docs
