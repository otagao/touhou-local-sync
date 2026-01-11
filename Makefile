# Makefile for touhou-local-sync project

# Project variables
BINARY_NAME=thlocalsync.exe
CMD_PATH=./cmd/thlocalsync
LICENSE_TOOL_VERSION=v1.6.0

# Go environment
GOPATH ?= $(shell go env GOPATH)
GO_LICENSES := $(GOPATH)/bin/go-licenses

# Windows環境での注意:
# go-licensesはGo 1.25のツールチェーン形式に未対応のため、以下の制約があります：
# - ローカルでのNOTICE生成: Go 1.23系が必要（WSL環境でgo1.23をインストール）
# - GitHub Actions: 自動的に実行されます（推奨）
# - 依存関係を追加した場合: GitHub Actionsに任せるか、WSLでGo 1.23を使用

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

## Build targets

.PHONY: build
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BINARY_NAME)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -rf dist/
	@echo "Clean complete"

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

## License management targets

.PHONY: install-license-tool
install-license-tool: ## Install go-licenses tool
	@echo "Installing go-licenses $(LICENSE_TOOL_VERSION)..."
	@go install github.com/google/go-licenses@$(LICENSE_TOOL_VERSION)
	@echo "go-licenses installed successfully"

.PHONY: license-generate
license-generate: install-license-tool ## Generate NOTICE file with full license texts
	@echo "Generating NOTICE file with full license texts..."
	@rm -f NOTICE.tmp NOTICE.err
	@echo "Running: $(GO_LICENSES) report $(CMD_PATH) --template=scripts/notice.tmpl"
	@$(GO_LICENSES) report $(CMD_PATH) \
		--template=scripts/notice.tmpl \
		--ignore=github.com/otagao/touhou-local-sync \
		--ignore=std > NOTICE.tmp 2>NOTICE.err || true
	@if [ -s NOTICE.err ]; then \
		echo "Errors from go-licenses:"; \
		cat NOTICE.err; \
	fi
	@if [ -s NOTICE.tmp ]; then \
		if grep -q "github.com\|gopkg.in\|golang.org" NOTICE.tmp; then \
			mv NOTICE.tmp NOTICE && echo "NOTICE file updated successfully"; \
			echo "File size: $$(wc -c < NOTICE) bytes"; \
			rm -f NOTICE.err; \
		else \
			echo "Warning: go-licenses produced output but no dependencies detected"; \
			echo "Trying fallback method..."; \
			bash scripts/fallback-notice.sh $(CMD_PATH) NOTICE; \
			rm -f NOTICE.tmp NOTICE.err; \
		fi; \
	else \
		echo "Warning: go-licenses produced no output"; \
		echo "Trying fallback method..."; \
		bash scripts/fallback-notice.sh $(CMD_PATH) NOTICE; \
		rm -f NOTICE.tmp NOTICE.err; \
	fi

.PHONY: license-check
license-check: install-license-tool ## Check licenses for forbidden types
	@echo "Checking licenses for forbidden types..."
	@$(GO_LICENSES) check $(CMD_PATH) \
		--disallowed_types=forbidden,unknown \
		--ignore=github.com/otagao/touhou-local-sync \
		--ignore=std
	@echo "License check passed"

.PHONY: license-audit
license-audit: license-check license-generate ## Run complete license audit (check + generate)
	@echo "License audit completed successfully"

## Development helpers

.PHONY: deps
deps: ## Download Go module dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

.PHONY: run
run: ## Run the application
	@go run $(CMD_PATH)

.PHONY: all
all: fmt vet test build ## Run all checks and build
