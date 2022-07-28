SHELL := /usr/bin/env bash -euo pipefail -c

default: test

# Always just install the git hooks.
_ := $(shell cd .git/hooks && ln -fs ../../dev/git_hooks/* .)

ifneq ($(PRODUCT_VERSION),)
CURR_VERSION := $(PRODUCT_VERSION)
else
CURR_VERSION := $(shell cat dev/VERSION)
endif

PRODUCT_VERSION ?= $(CURR_VERSION)
export PRODUCT_VERSION

CURR_VERSION_CL := dev/changes/v$(CURR_VERSION).md

DIRTY := $(shell git diff --exit-code > /dev/null 2>&1 || echo -n "dirty-")

ifneq ($(PRODUCT_REVISION),)
CURR_REVISION := $(PRODUCT_REVISION)
else
CURR_REVISION := $(shell git rev-parse HEAD)
PRODUCT_REVISION := $(CURR_REVISION)
endif

PRODUCT_REVISION_TIME := $(shell git show -s --format=%ci "$(CURR_REVISION)" || echo "")

export PRODUCT_REVISION
export PRODUCT_REVISION_TIME

CURR_REVISION    := $(DIRTY)$(CURR_REVISION)
PRODUCT_REVISION ?= $(CURR_REVISION)

test: test/go

cover: GO_TEST_FLAGS := -coverprofile=coverage.profile
cover: test/go
	@go tool cover -html=coverage.profile && rm coverage.profile

test/update: test/go/update

CLINAME := $(notdir $(CURDIR))
CLI     := bin/$(CLINAME)
RUNCLI  := @./$(CLI)

BIN_PATH ?= $(CLI)
export BIN_PATH

LDFLAGS      := -X 'main.FullVersion=$$PRODUCT_VERSION'
LDFLAGS      += -X 'main.Revision=$$PRODUCT_REVISION'
LDFLAGS      += -X 'main.RevisionTime=$$PRODUCT_REVISION_TIME'
BUILD_FLAGS  := -v -trimpath -buildvcs=false -ldflags="$(LDFLAGS)"
INSTRUCTIONS := go build -o "$$BIN_PATH" $(BUILD_FLAGS)
INSTALL      := go install $(BUILD_FLAGS)

export INSTRUCTIONS

FORMAT_INSTRUCTIONS := sed -E -e 's/-([^-]+)/\n  -\1/g' -e 's/-X/  -X/g' -e 's/\n/\\\n/g' -e 's/  \\/ \\/g'

instructions:
	@printf "%s\n" "$$INSTRUCTIONS" | $(FORMAT_INSTRUCTIONS)

dogfood:
	actions-go-build

.PHONY: dev
dev:
	@$(MAKE) instructions
	@$(MAKE) env
	@$(MAKE) cli

env:
	@echo "ENV:"
	@echo "  PRODUCT_VERSION=$$PRODUCT_VERSION"
	@echo "  PRODUCT_REVISION=$$PRODUCT_REVISION"
	@echo "  PRODUCT_REVISION_TIME=$$PRODUCT_REVISION_TIME"

.PHONY: $(BIN_PATH)
$(BIN_PATH):
	@$(INSTRUCTIONS)

cli: $(BIN_PATH)
	@echo "Build successful."
	$(BIN_PATH) --version

ifneq ($(GITHUB_PATH),)
install: $(BIN_PATH)
	@echo "$(dir $(CURDIR)/$(CLI))" >> "$$GITHUB_PATH"
	@echo "Command '$(CLINAME)' installed to GITHUB_PATH"
	$(CLINAME) --version
else
install:
	@$(INSTALL)
	@echo "Command '$(CLINAME)' installed to GOBIN"
	$(CLINAME) --version
endif

mod/framework/update:
	@REF="$$(cd ../composite-action-framework-go && make module/ref/head)" && \
		go get "$$REF"

# The run/cli/... targets build and then run the CLI itself
# which is usful for quickly seeing its output whilst developing.

run/cli/%: $(CLI)

run/cli/config:
	$(RUNCLI) config

run/cli/config/github:
	$(RUNCLI) config -github

run/cli/env:
	$(RUNCLI) env

# run/cli/env/describe is called by dev/docs/environment_doc
run/cli/env/describe:
	$(RUNCLI) config env describe

run/cli/env/dump:
	$(RUNCLI) env dump

run/cli/primary:
	$(RUNCLI) primary

run/cli/verification:
	$(RUNCLI) verification

run/cli/compare:
	$(RUNCLI) compare

test/go/update: export UPDATE_TESTDATA := true
test/go/update: test/go

test/go: 
	go test $(GO_TEST_FLAGS) ./...

.PHONY: docs
docs: readme changelog

.PHONY: readme
readme:
	@./dev/docs/readme_update

.PHONY: changelog
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

#LDFLAGS += -X 'main.Version=1.2.3'
#LDFLAGS += -X 'main.Revision=cabba9e'
#LDFLAGS += -X 'main.RevisionTime=2022-05-30T14:45:00+00:00'

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
REPLACEMENTS += -e 's|on: \{ push: \{ branches: main \} \}|on: push|g'
REPLACEMENTS += -e 's|main|current branch|g'

define UPDATE_CURRENT_BRANCH_EXAMPLE
@TARGET="$(1).currentbranch.yml" && \
echo "# GENERATED FILE, DO NOT MODIFY; INSTEAD EDIT $(1) AND RUN 'make examples'" > "$$TARGET" && \
sed -E $(REPLACEMENTS)  "$(1)" >> "$$TARGET" && \
echo "Example file updated: $$TARGET"
endef

examples:
	$(call UPDATE_CURRENT_BRANCH_EXAMPLE,$(EXAMPLE1))
	$(call UPDATE_CURRENT_BRANCH_EXAMPLE,$(EXAMPLE2))
