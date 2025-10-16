# FDOT Makefile
.PHONY: help build-credmgr

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build-credmgr: ## Install credmgr locally (go install) and build Windows version
	@echo "Installing credmgr locally..."
	@go install ./cmd/credmgr
	@echo "✅ credmgr installed to $$(go env GOPATH)/bin/ (or $$GOBIN if set)"
	@echo "   Make sure $$(go env GOPATH)/bin is in your PATH"
	@echo ""
	@echo "Building credmgr for Windows..."
	@mkdir -p /mnt/c/Users/KN018NZ/bin
	@GOOS=windows GOARCH=amd64 go build -o /mnt/c/Users/KN018NZ/bin/credmgr.exe ./cmd/credmgr
	@cp pkg/fdh/credmgr/test-credmgr.bat /mnt/c/Users/KN018NZ/bin/
	@echo "✅ Windows credmgr.exe and test-credmgr.bat copied to /mnt/c/Users/KN018NZ/bin/"
	@echo ""
	@echo "🎉 Build complete!"