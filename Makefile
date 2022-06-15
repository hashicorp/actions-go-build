SHELL := /usr/bin/env bash -euo pipefail -c

default: test

# Always just install the git hooks.
_ := $(shell cd .git/hooks && ln -fs ../../dev/git_hooks/* .)

CURR_VERSION := $(shell cat dev/VERSION)
CURR_VERSION_CL := dev/changes/v$(CURR_VERSION).md

BATS := bats -j 10 -T

test: test/bats

test/bats:
	# Running bats tests in scripts/
	@$(BATS) scripts/

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

example: export REPRODUCIBLE             := assert
example: export INSTRUCTIONS             := cd testdata/example-app && $(GO_BUILD)
example: export OS                       := $(shell go env GOOS)
example: export ARCH                     := $(shell go env GOARCH)
example: export PRODUCT_REPOSITORY       := example-app
example: export PRODUCT_NAME             := example-app
example: export PRODUCT_VERSION          := 1.2.3
example: export EXAMPLE_TMP              := $(shell mktemp -d)
example: export GITHUB_ENV               := $(EXAMPLE_TMP)/github_env
example: export GITHUB_STEP_SUMMARY      := $(EXAMPLE_TMP)/github_step_summary
example: export PRIMARY_BUILD_ROOT       := $(EXAMPLE_TMP)/primary
example: export VERIFICATION_BUILD_ROOT  := $(EXAMPLE_TMP)/verification
example:
	rm -rf "$(EXAMPLE_TMP)" && mkdir -p "$(EXAMPLE_TMP)"
	cp -Rf . "$(PRIMARY_BUILD_ROOT)"
	cd $(PRIMARY_BUILD_ROOT) && \
		source scripts/inputs.bash && \
		digest_inputs && \
		./scripts/primary_build && \
		./scripts/local_verification_build && \
		trap 'cat $(GITHUB_STEP_SUMMARY)' EXIT && \
		./scripts/compare_digests


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
