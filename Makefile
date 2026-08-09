# srekit — local development entrypoint.
#
# Targets here are the same ones GitHub Actions calls, so `make ci` locally and
# a green pipeline mean the same thing.
#
# COMPATIBILITY: this Makefile must keep working on GNU Make 3.81 — that is the
# system `make` on macOS and Apple will not ship a newer one (GPLv3). Do not use
# .ONESHELL / .SHELLFLAGS (3.82+), `!=` / $(file ...) (4.0+), $(intcmp) / $(let)
# (4.4+). One command per recipe line; Make aborts on the first non-zero exit.

SHELL := /usr/bin/env bash
MAKEFLAGS += --no-print-directory
.DEFAULT_GOAL := help

MAKE_MAJOR := $(firstword $(subst ., ,$(MAKE_VERSION)))
ifneq ($(MAKE_MAJOR),3)
ifndef CI
$(info note: GNU Make $(MAKE_VERSION) — this Makefile targets 3.81 (macOS system make), keep it compatible)
endif
endif

BINARY_NAME  ?= srekit
DIST_DIR     ?= $(CURDIR)/dist
SITE_DIR     ?= $(CURDIR)/site
LOCALBIN     ?= $(CURDIR)/bin
VENV         ?= $(CURDIR)/.venv
GO           ?= go
PYTHON       ?= python3
COVERPROFILE ?= cover.out
GOTEST_FLAGS ?= -v

GOLANGCI_LINT_VERSION ?= v2.12.2
GOVULNCHECK_VERSION   ?= v1.6.0

GOLANGCI_LINT := $(LOCALBIN)/golangci-lint
GOVULNCHECK   := $(LOCALBIN)/govulncheck
MKDOCS        := $(VENV)/bin/mkdocs

GOLANGCI_LINT_STAMP := $(LOCALBIN)/.golangci-lint-$(GOLANGCI_LINT_VERSION)
GOVULNCHECK_STAMP   := $(LOCALBIN)/.govulncheck-$(GOVULNCHECK_VERSION)

# --- meta -------------------------------------------------------------------

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make <target>\n\n"} \
		/^[a-zA-Z0-9_-]+:.*##/ { printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2 }' \
		$(MAKEFILE_LIST)

# --- build ------------------------------------------------------------------

run: ## Run via go run: make run ARGS="postmortem --title x"
	$(GO) run . $(ARGS)

build: ## Build the binary into ./dist
	$(GO) mod download
	CGO_ENABLED=0 $(GO) build -o $(DIST_DIR)/$(BINARY_NAME) .

install: ## go install
	$(GO) install

tidy: ## go mod tidy
	$(GO) mod tidy

fmt: ## gofmt -s -w .
	gofmt -s -w .

vet: ## go vet ./...
	$(GO) vet ./...

# --- test / lint ------------------------------------------------------------

test: ## Short tests with coverage
	$(GO) test --short -coverprofile=$(COVERPROFILE) $(GOTEST_FLAGS) ./...

test-race: ## Race tests with coverage (this is what CI runs)
	$(GO) test -race -coverprofile=$(COVERPROFILE) $(GOTEST_FLAGS) ./...

lint: $(GOLANGCI_LINT_STAMP) ## golangci-lint run
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_FLAGS)

lint-fix: $(GOLANGCI_LINT_STAMP) ## golangci-lint run --fix
	$(GOLANGCI_LINT) run --fix

govulncheck: $(GOVULNCHECK_STAMP) ## Scan for known vulnerabilities
	$(GOVULNCHECK) ./...

ci: lint test-race ## Quick pre-push check: lint + race tests

ci-full: lint test-race govulncheck docs-build ## Everything CI runs

# --- release ----------------------------------------------------------------

release-check: ## goreleaser check
	goreleaser check

release-dry: ## Snapshot build, no publish (binaries land in ./dist)
	goreleaser release --clean --snapshot --skip=publish,sign

# --- docs -------------------------------------------------------------------

$(MKDOCS): docs/requirements.txt
	$(PYTHON) -m venv $(VENV)
	$(VENV)/bin/pip install -q --upgrade pip
	$(VENV)/bin/pip install -q -r docs/requirements.txt
	@touch $@

docs-install: $(MKDOCS) ## Install MkDocs + plugins into ./.venv

docs-serve: $(MKDOCS) ## Serve the docs at http://127.0.0.1:8000
	$(MKDOCS) serve

docs-build: $(MKDOCS) ## Build the docs into ./site (strict)
	$(MKDOCS) build --strict

docs-deploy: $(MKDOCS) ## Deploy the docs to gh-pages (called from CI)
	$(MKDOCS) gh-deploy --force --no-history

# --- tools ------------------------------------------------------------------

tools: $(GOLANGCI_LINT_STAMP) $(GOVULNCHECK_STAMP) ## Install pinned tools into ./bin

$(GOLANGCI_LINT_STAMP):
	@mkdir -p $(LOCALBIN)
	@rm -f $(LOCALBIN)/.golangci-lint-*
	GOBIN=$(LOCALBIN) $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@touch $@

$(GOVULNCHECK_STAMP):
	@mkdir -p $(LOCALBIN)
	@rm -f $(LOCALBIN)/.govulncheck-*
	GOBIN=$(LOCALBIN) $(GO) install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	@touch $@

clean: ## Remove build artifacts
	rm -rf $(DIST_DIR) $(SITE_DIR) $(LOCALBIN) $(COVERPROFILE)

.PHONY: help run build install tidy fmt vet test test-race lint lint-fix \
        govulncheck ci ci-full release-check release-dry \
        docs-install docs-serve docs-build docs-deploy tools clean
