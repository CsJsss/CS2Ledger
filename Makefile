SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

export CGO_ENABLED := 1

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Setup

.PHONY: setup
setup: install-wails ## Install all dev tools, dependencies, and git hooks.
	go mod download
	go install tool
	cd frontend && npm ci

.PHONY: install-wails
install-wails: ## Install Wails CLI (optional, for wails dev / wails build).
	go install github.com/wailsapp/wails/v2/cmd/wails@latest

##@ Development

.PHONY: dev
dev: ## Start Wails dev server with hot reload.
	wails dev

.PHONY: fmt
fmt: ## Run go fmt against all code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against all code.
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint against all code.
	go tool golangci-lint run ./...

.PHONY: test
test: ## Run Go tests.
	go test ./... -v -count=1

##@ Build

.PHONY: build
build: lint test ## Build production binary → build/bin/.
	wails build

.PHONY: build-dev
build-dev: ## Build development binary (debug symbols).
	wails build -debug

.PHONY: build-windows
build-windows: lint test ## Build Windows production binary.
	wails build -platform windows/amd64

.PHONY: build-dev-windows
build-dev-windows: ## Build Windows development binary.
	CC=x86_64-w64-mingw32-gcc wails build -platform windows/amd64 -debug

.PHONY: frontend
frontend: ## Build frontend assets only.
	cd frontend && npm ci && npm run build

.PHONY: clean
clean: ## Remove build artifacts.
	rm -rf build/ dist/ frontend/dist/

##@ Database

.PHONY: db-new
db-new: ## Create a new migration file. Usage: make db-new NAME=add_xxx
	@test -n "$(NAME)" || { echo "Usage: make db-new NAME=<migration_name>"; exit 1; }
	@touch migrations/$(shell date +%s)_$(NAME).sql
	@echo "Created: migrations/$(shell ls -t migrations/ | head -1)"

##@ Docker

.PHONY: docker-build-linux
docker-build-linux: ## Build Linux binary via Docker → build/bin/.
	DOCKER_BUILDKIT=1 docker build -f install/docker/Dockerfile.linux --output build/bin/ .

.PHONY: docker-build-windows
docker-build-windows: ## Build Windows .exe via Docker → build/bin/.
	DOCKER_BUILDKIT=1 docker build -f install/docker/Dockerfile.windows --output build/bin/ .

.PHONY: docker-build
docker-build: docker-build-linux docker-build-windows ## Build Linux + Windows binaries via Docker.
