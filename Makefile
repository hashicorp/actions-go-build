SHELL := /usr/bin/env bash -euo pipefail -c

MAKEFLAGS := --jobs=10

PRODUCT_NAME := actions-go-build
DESTDIR ?= /usr/local/bin

# Set AUTOCLEAR=1 to have the terminal cleared before running builds,
# tests, and installs.
CLEAR := $(AUTOCLEAR)
ifeq ($(CLEAR),1)
	CLEAR := clear
else
	CLEAR :=
endif

SOURCE_ID       := .git/source-id
SOURCE_ID_VALUE := $(shell SOURCE_ID=$(SOURCE_ID) ./dev/update-source-id)

# Uncomment to show the source ID.
# $(info SOURCE_ID=$(SOURCE_ID_VALUE))

default: run

ifeq ($(TMPDIR),)
TMPDIR := $(RUNNER_TEMP)
export TMPDIR
endif
ifeq ($(TMPDIR),)
$(error Neither TMPDIR nor RUNNER_TEMP are set.)
endif

TEST_LOG := $(TMPDIR)/go_tests.log
RUN_TESTS_QUIET := @$(MAKE) test > "$(TEST_LOG)" 2>&1 || { cat "$(TEST_LOG)" ; exit 1; }

# Always just install the git hooks unless in CI (GHA sets CI=true as do many CI providers).
ifeq ($(CI),true)
_ := $(shell mkdir -p .git/hooks && cd .git/hooks && ln -fs ../../dev/git_hooks/* .)
endif

ifneq ($(PRODUCT_VERSION),)
CURR_VERSION := $(PRODUCT_VERSION)
else
CURR_VERSION := $(shell cat dev/VERSION)
endif

CURR_VERSION_CL := dev/changes/v$(CURR_VERSION).md

DIRTY := $(shell git diff --exit-code > /dev/null 2>&1 || echo -n "dirty-")

ifneq ($(PRODUCT_REVISION),)
CURR_REVISION := $(PRODUCT_REVISION)
else
CURR_REVISION := $(shell git rev-parse HEAD)
PRODUCT_REVISION := $(CURR_REVISION)
endif

CURR_REVISION    := $(DIRTY)$(CURR_REVISION)
PRODUCT_REVISION ?= $(CURR_REVISION)
CLINAME          := $(PRODUCT_NAME)

# Release versions of the CLI are built in three phases:
#
#    1) INITIAL_BUILD       - No build metadata.
#    2) INTERMEDIATE_BUILD    - Some build metadata.
#    3) BOOTSTRAPPED_BUILD           - All build metadata.
#
# See comments below for more explanation.

TMP_BASE := $(TMPDIR)/actions-go-build.builds/$(SOURCE_ID_VALUE)

# INITIAL_BUILD is a build of the CLI done using just `go build ...`. This is used to bootstrap
# compiling the CLI using itself, for dogfooding purposes. The INITIAL_BUILD contains none of the
# automatically generated metadata like the version or revision. It is used to build the
# intermediate build...
INITIAL_BUILD := $(TMP_BASE)/initial/$(CLINAME)

# INTERMEDIATE_BUILD is a build of the CLI done using the INITIAL_BUILD build. Because it used
# INITIAL_BUILD (i.e. the code in this repo) to build itself, it contains automatically generated
# metadata like the version and revision. However, it does not contain the metadata about the
# version of actions-go-build that built it because INITIAL_BUILD doesn't have that metadata
# available to inject.
INTERMEDIATE_BUILD := $(TMP_BASE)/intermediate/$(CLINAME)

# BOOTSTRAPPED_BUILD is the final build of the CLI, done using the INTERMEDIATE_BUILD. Because
# INTERMEDIATE_BUILD contains build metadata (e.g. version and revision), it is able to inject
# that information, into this final build as "tool metadata". Thus we can track the provanance of
# this binary  just like we are able to with any product binaries also built using this tool.
BOOTSTRAPPED_BUILD := $(TMP_BASE)/bootstrapped/$(CLINAME)

# HOST_PLATFORM_TARGETS are targets that must always produce output compatible with
# the current host platform. We therefore unset the GOOS and GOARCH variable to allow
# the defaults to shine through.
HOST_PLATFORM_TARGETS := $(INITIAL_BUILD) $(INTERMEDIATE_BUILD) test/go
$(HOST_PLATFORM_TARGETS): export GOOS :=
$(HOST_PLATFORM_TARGETS): export GOARCH :=

HOST_PLATFORM_TARGET_ENV := GOOS= GOARCH= OS= ARCH=

#
# Targets
#

test: test/go

.PHONY: test/go
test/go: compile
	@$(HOST_PLATFORM_TARGET_ENV) go test $(GO_TEST_FLAGS) ./...

cover: GO_TEST_FLAGS := -coverprofile=coverage.profile
cover: test/go
	@go tool cover -html=coverage.profile && rm coverage.profile

test/update: test/go/update

.PHONY: test/go/update
test/go/update: export UPDATE_TESTDATA := true
test/go/update: test/go
	@echo "Test data updated."

.PHONY: compile
compile:
	@$(CLEAR)
	@go build ./...

.PHONY: env
env:
	@echo "ENV:"
	@echo "  PRODUCT_VERSION=$$PRODUCT_VERSION"
	@echo "  PRODUCT_REVISION=$$PRODUCT_REVISION"
	@echo "  PRODUCT_REVISION_TIME=$$PRODUCT_REVISION_TIME"

# When building the binary, we first do a plain 'go build' to build a bootstrap
# binary that contains no version info. Then we use that version of the binary
# to build this product with all its own version info added automatically from the
# build context.
#
# We then use _that_ binary to build yet another binary, this time with the
# correct tool version injected into the build. So it now contains the version of
# actions-go-build that built this verison of actions-go-build as well (these
# versions are the same).
#
# By performing all three builds, we are fully dogfooding the build process, and
# ensuring that the version we are releasing has been used to produce itself.

$(INITIAL_BUILD): $(SOURCE_ID)
	@echo "# Running tests..." 1>&2
	@$(RUN_TESTS_QUIET)
	@BIN_PATH="$@" ./dev/build initial > /dev/null

RUN_QUIET = OUT="$$($(1) 2>&1)" || { \
				echo "Command Failed: $(notdir $(1))"; echo "$$OUT"; exit 1; \
			} 

$(INTERMEDIATE_BUILD): $(INITIAL_BUILD)
	@BIN_PATH="$@" ./dev/build intermediate "$<" > /dev/null

$(BOOTSTRAPPED_BUILD): $(INTERMEDIATE_BUILD)
	@BIN_PATH="$@" ./dev/build bootstrapped "$<" > /dev/null

cli: $(BOOTSTRAPPED_BUILD)
	@echo "Build successful."
	$(BOOTSTRAPPED_BUILD) --version

.PHONY: install
# Ensure install always targets the host platform.
install: export GOOS :=
install: export GOARCH :=

ifneq ($(GITHUB_PATH),)
# install for GitHub Actions.
install: $(BOOTSTRAPPED_BUILD)
	@echo "$(dir $(CURDIR)/$(BOOTSTRAPPED_BUILD))" >> "$(GITHUB_PATH)"
	@echo "Command '$(CLINAME)' installed to GITHUB_PATH ($(GITHUB_PATH))"
	@export PATH="$$(cat $(GITHUB_PATH))" && $(CLINAME) --version
else
# install for local use.
install: $(BOOTSTRAPPED_BUILD)
	@mv "$<" "$(DESTDIR)"
	@V="$$($(CLINAME) version -short)" && \
		echo "# $(CLINAME) v$$V installed to $(DESTDIR)"
endif

.PHONY: mod/framework/update
mod/framework/update:
	@REF="$$(cd ../composite-action-framework-go && make module/ref/head)" && \
		go get "$$REF"

# The run/... targets build and then run the CLI itself
# which is usful for quickly seeing its output whilst developing.

.PHONY: run
run: $(INITIAL_BUILD)
	@$${QUIET:-false} || $(CLEAR)
	@$${QUIET:-false} || echo "\$$ $(notdir $<) $(RUN)"
	@$(INITIAL_BUILD) $(RUN)

.PHONY: docs
docs: readme changelog

.PHONY: readme
readme:
	@./dev/docs/readme_update

.PHONY: changelog
changelog:
	@./dev/docs/changelog_update

.PHONY: changelog/view
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

GH := $(shell command -v gh 2> /dev/null)
ifeq ($(GH),)
	GH := echo "Please install the [GitHub CLI](https://github.com/cli/cli\#installation)"; \#
else
	GH_AUTHED := $(shell gh auth status > /dev/null 2>&1 && echo true)
ifneq ($(GH_AUTHED),true)
	GH := echo "Please ensure 'gh auth status' succeeds and try again."; \#
endif
endif

#
# Release build targets
#
define FINAL_BUILD_TARGETS

# build/<platform> does not require a clean worktree and results in a "Development" build.

build/$(1): $$(BOOTSTRAPPED_BUILD)
	@./dev/build dev "$$<" "$1" > /dev/null

DEV_BUILDS := $$(DEV_BUILDS) build/$(1)

dist/$1/actions-go-build \
out/actions-go-build_$(CURR_VERSION)_$(subst /,_,$1).zip \
zip/$(1) \
release/build/$(1): $$(BOOTSTRAPPED_BUILD)
	@./dev/build release "$$<" "$1" > /dev/null
RELEASE_BUILDS := $$(RELEASE_BUILDS) release/build/$(1)
RELEASE_ZIPS   := $$(RELEASE_ZIPS) out/actions-go-build_$(CURR_VERSION)_$(subst /,_,$1).zip

endef

TARGET_PLATFORMS := $(shell ./dev/build platforms)
$(eval $(foreach P,$(TARGET_PLATFORMS),$(call FINAL_BUILD_TARGETS,$(P))))

.PHONY: $(RELEASE_BUILDS)
.PHONY: $(DEV_BUILDS)

build: $(DEV_BUILDS)

release/build: $(RELEASE_BUILDS)

release: $(RELEASE_ZIPS)
	@for Z in $^; do echo $$Z; done

version: version/check
	@LATEST="$(shell $(GH) release list -L 1 --exclude-drafts | grep Latest | cut -f1)"; \
		echo "Working on v$(CURR_VERSION) (Latest public release: $$LATEST)"
.PHONY: version

version/check:
	@./dev/release/version_check || { \
		echo "Tip: run 'make version/set VERSION=<next version>'"; \
		exit 1; \
	}
.PHONY: version/check

version/set:
	@[[ -z "$(VERSION)" ]] && { \
		echo "Usage:" && \
		echo "    make VERSION=<version> version/set" && \
		echo "Current Version:" && \
		echo "    $(CURR_VERSION)" && \
		exit 1; \
	}; \
	./dev/release/set_version "$(VERSION)" && \
	make changelog && \
	git add dev/VERSION dev/changes/v$(VERSION).md CHANGELOG.md && \
	git commit -m "set development version to v$(VERSION)"
.PHONY: version/set

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
