.PHONY: help install dev build preview sync lint format clean upgrade

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

install: ## Install dependencies
	npm install

dev: install ## Start development server
	npm run dev

build: ## Build for production
	npm run build

preview: build ## Preview production build
	npm run preview

sync: ## Sync content from Obsidian
	python3 bin/obsidian_sync.py

sync-dry-run: ## Dry run of Obsidian sync (shows what would be done)
	python3 bin/obsidian_sync.py --dry-run

lint: ## Run ESLint
	npm run lint

format: ## Format code with Prettier
	npm run format

format-check: ## Check code formatting
	npm run format:check

type-check: ## Run TypeScript type checking
	npm run type-check

clean: ## Remove build artifacts and dependencies
	rm -rf dist node_modules .astro

upgrade: ## Upgrade Astro to latest version
	npx @astrojs/upgrade

local: dev ## Alias for dev (backwards compatibility)