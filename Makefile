SHELL := /usr/bin/env bash -euo pipefail -c

default: test

# tools/mac tries to install dependencies on mac using homebrew.
tools/mac:
	brew install coreutils util-linux

BATS := bats -j 10 -T

test: test/bats

test/bats:
	# Running bats tests in scripts/
	@$(BATS) scripts/

.PHONY: docs
docs:
	@./scripts/codegen/update_docs

LDFLAGS += -X 'main.Version=1.2.3'
LDFLAGS += -X 'main.Revision=cabba9e'
LDFLAGS += -X 'main.RevisionTime=2022-05-30T14:45:00+00:00'

.PHONY: example-app
example-app:
	@cd testdata/example-app && go build -ldflags "$(LDFLAGS)" . && ./example-app
