# Makefile for touhou-local-sync project

# Project variables
BINARY_NAME=thlocalsync.exe
CMD_PATH=./cmd/thlocalsync
LICENSE_TOOL_VERSION=latest

# Windows環境での注意:
# go-licensesはLinux/macOS向けツールのため、以下の方法で実行してください：
# 1. WSL2環境で実行: wsl make license-generate
# 2. GitHub Actions上で実行（推奨）
# 3. Git BashまたはMinGW環境で実行（動作不安定の可能性あり）

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
	@go-licenses report $(CMD_PATH) \
		--template=scripts/notice.tmpl \
		--ignore=github.com/otagao/touhou-local-sync > NOTICE
	@echo "NOTICE file updated successfully"

.PHONY: license-check
license-check: install-license-tool ## Check licenses for forbidden types
	@echo "Checking licenses for forbidden types..."
	@go-licenses check $(CMD_PATH) \
		--disallowed_types=forbidden,unknown \
		--ignore=github.com/otagao/touhou-local-sync
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
