SHELL := /bin/bash
PATH := $(shell go env GOPATH)/bin:$(PATH)
export PATH
.PHONY: help fmt lint vet fix tidy build test clean install-hooks bench

# ── Default ──
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ── Formatting ──
fmt:
	golangci-lint fmt ./...

# ── Linting ──
fix: ## Run golangci-lint with auto-fix
	golangci-lint run --fix --timeout=5m || true

lint: ## Run golangci-lint
	golangci-lint run ./... --timeout=5m

vet: ## Run go vet
	go vet ./...

# ── Maintenance ──
tidy: ## Run go mod tidy + verify
	go mod tidy
	go mod verify

# ── Build & Test ──
build: ## Build all packages
	go build ./...

test: ## Run tests
	go test -race -count=1 ./...

bench: ## Run benchmarks
	go test -bench=. ./internal/tests -count=5

# ── All-in-one ──
check: fmt fix tidy lint vet build ## Format, fix, tidy, lint, vet, build

# ── Git hooks ──
install-hooks: ## Install .githooks as git hooks
	cp .githooks/* .git/hooks/ 2>/dev/null || true
	chmod +x .git/hooks/* 2>/dev/null || true

