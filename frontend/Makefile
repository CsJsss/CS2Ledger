# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: dev
dev: install ## Start Vite dev server with HMR.
	npm run dev

.PHONY: install
install: ## Install npm dependencies.
	npm ci

##@ Quality

.PHONY: lint
lint: ## Run ESLint on all source files.
	npm run lint

.PHONY: fmt
format: ## Format source files with Prettier.
	npm run format

.PHONY: typecheck
typecheck: ## Run TypeScript type checking.
	npm run typecheck

.PHONY: check
check: lint typecheck ## Run all quality checks.

##@ Build

.PHONY: build
build: install ## Build production bundle.
	npm run build

.PHONY: preview
preview: ## Preview production build locally.
	npm run preview

##@ Maintenance

.PHONY: hooks
hooks: ## Install husky git hooks.
	npx husky install

.PHONY: clean
clean: ## Remove build output and node_modules.
	rm -rf dist/ node_modules/

.PHONY: outdated
outdated: ## Check for outdated npm packages.
	npm outdated
