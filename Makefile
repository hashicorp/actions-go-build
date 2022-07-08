SHELL := /usr/bin/env bash -euo pipefail -c

default: test

# Always just install the git hooks.
_ := $(shell cd .git/hooks && ln -fs ../../dev/git_hooks/* .)

CURR_VERSION := $(shell cat dev/VERSION)
CURR_VERSION_CL := dev/changes/v$(CURR_VERSION).md

BATS := bats -j 10 -T

test: test/bats test/go

test/update: test/go/update

CLINAME := $(notdir $(CURDIR))
CLI     := bin/$(CLINAME)
RUNCLI  := @./$(CLI)

cli:
	@go build -trimpath -o "$(CLI)"

ifneq ($(GITHUB_PATH),)
install: cli
	@echo "$(dir $(CURDIR)/$(CLI))" >> "$$GITHUB_PATH"
	@echo "Command '$(CLINAME)' installed to GITHUB_PATH"
else
install: cli
	@go install "$(CLIPKG)"
	@echo "Command '$(CLINAME)' installed to GOBIN"
endif


# The run/cli/... targets build and then run the CLI itself
# which is usful for quickly seeing its output whilst developing.

run/cli/%: export PRODUCT_REPOSITORY := hashicorp/actions-go-build
run/cli/%: export PRODUCT_VERSION    := 1.2.3
run/cli/%: export OS                 := $(shell go env GOOS)
run/cli/%: export ARCH               := $(shell go env GOARCH)
run/cli/%: export REPRODUCIBLE       := assert
run/cli/%: export INSTRUCTIONS       := echo "Running build in bash"; go build -o "$$BIN_PATH"

run/cli/config: cli
	$(RUNCLI) config

run/cli/config/github: cli
	$(RUNCLI) config -github

run/cli/env: cli
	$(RUNCLI) env

# run/cli/env/describe is called by dev/docs/environment_doc
run/cli/env/describe: cli
	$(RUNCLI) env describe

run/cli/env/dump: cli
	$(RUNCLI) env dump

run/cli/primary: cli
	$(RUNCLI) primary

run/cli/verification: cli
	$(RUNCLI) verification

test/bats:
	# Running bats tests in scripts/
	@$(BATS) scripts/

test/go/update: export UPDATE_TESTDATA := true
test/go/update: test/go

test/go: 
	go test ./...

.PHONY: docs
docs: readme changelog

readme:
	@./dev/docs/readme_update

changelog:
	@./dev/docs/changelog_update

changelog/view:
	@echo "Current development version: $(CURR_VERSION)"
	@echo
	@[[ -s "$(CURR_VERSION_CL)" ]] && cat "$(CURR_VERSION_CL)" || echo '    - changelog empty -'
	@echo
	@echo "Use 'make changelog/add' to edit this version's changelog."

CL_REMINDERS_COMMENT := RECENT COMMITS TO JOG YOUR MEMORY (DELETE THIS SECTION WHEN DONE)...

# changelog/add appends recent commit logs (since the file was last updated)
# to the changelog, and opens it in the editor.
changelog/add:
	@echo "<!-- $(CL_REMINDERS_COMMENT)" >> "$(CURR_VERSION_CL)"
	@git log $$(git rev-list -1 HEAD "$(CURR_VERSION_CL)")..HEAD >> "$(CURR_VERSION_CL)"
	@echo " END $(CL_REMINDERS_COMMENT) -->" >> "$(CURR_VERSION_CL)"
	@$(EDITOR) "$(CURR_VERSION_CL)"
	@$(MAKE) changelog
	@git add CHANGELOG.md "$(CURR_VERSION_CL)" && git commit -m "update changelog for v$(CURR_VERSION)" && \
		echo "==> Changelog updated and committed, thanks for keeping it up-to-date!"

.PHONY: debug/docs
debug/docs: export DEBUG := 1
debug/docs: docs

LDFLAGS += -X 'main.Version=1.2.3'
LDFLAGS += -X 'main.Revision=cabba9e'
LDFLAGS += -X 'main.RevisionTime=2022-05-30T14:45:00+00:00'

.PHONY: example-app
example-app:
	@cd testdata/example-app && go build -ldflags "$(LDFLAGS)" . && ./example-app

GO_BUILD := go build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" -o "$$BIN_PATH"

# 'make tools' will use the brew target if on Darwin.
# Otherwise it just prints a message about dependencies.
ifeq ($(shell uname),Darwin)
tools: tools/mac/brew
else
tools:
	@echo "Please ensure that BATS, coreutils, util-linux, github-markdown-toc, and GNU parallel are installed."
endif

# tools/mac/brew tries to install dependencies on mac using homebrew.
tools/mac/brew:
	brew bundle --no-upgrade	

.PHONY: release
release:
	@./dev/release/create

version:
	@echo "$(CURR_VERSION)"

version/check:
	@./dev/release/version_check

version/set:
	@[[ -z "$(VERSION)" ]] && { \
		echo "Usage:" && \
		echo "    make VERSION=<version> version/set" && \
		echo "Current Version:" && \
		echo "    $(CURR_VERSION)" && \
		exit 1; \
	}; \
	./dev/release/set_version "$(VERSION)" && \
	git add dev/VERSION dev/changes/v$(VERSION).md && \
	git commit -m "set development version to v$(VERSION)"

EXAMPLE1         := .github/workflows/example.yml
EXAMPLE1_CURRENT := .github/workflows/example_currentbranch.yml
EXAMPLE2         := .github/workflows/example-matrix.yml
EXAMPLE2_CURRENT := .github/workflows/example-matrix_currentbranch.yml

REPLACEMENTS := -e 's|hashicorp/actions-go-build@main|./|g'
REPLACEMENTS += -e 's|(main)|(current branch)|g'

define UPDATE_CURRENT_BRANCH_EXAMPLE
TARGET="$(1).currentbranch.yml" && \
echo "# GENERATED FILE, DO NOT MODIFY; INSTEAD EDIT $(1) AND RUN 'make examples'" > "$$TARGET" && \
sed $(REPLACEMENTS)  "$(1)" >> "$$TARGET"
endef

examples:
	$(call UPDATE_CURRENT_BRANCH_EXAMPLE,$(EXAMPLE1))
	$(call UPDATE_CURRENT_BRANCH_EXAMPLE,$(EXAMPLE2))
