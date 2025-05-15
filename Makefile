# Frida Injector Makefile
.PHONY: build clean test all

# Binary name
BINARY_NAME=frida-injector
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

# Directories
CMD_DIR=.
BUILD_DIR=build

# Build flags
LDFLAGS=-ldflags "-s -w"
CGO_FLAGS=CGO_CFLAGS="-Wno-error=incompatible-pointer-types"

all: clean build

build:
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@$(CGO_FLAGS) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

# Cross-compile for different platforms
build-linux:
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) for linux/amd64..."
	@$(CGO_FLAGS) GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 $(CMD_DIR)

build-darwin:
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) for darwin/amd64..."
	@$(CGO_FLAGS) GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 $(CMD_DIR)

# Run the program
run:
	@$(CGO_FLAGS) go run $(CMD_DIR)

# Install the program
install:
	@$(CGO_FLAGS) go install $(LDFLAGS)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean

# Run tests
test:
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Help target
help:
	@echo "Available targets:"
	@echo "  all        - Clean and build the project"
	@echo "  build      - Build for current platform"
	@echo "  build-linux - Build for Linux/amd64"
	@echo "  build-darwin - Build for macOS/amd64"
	@echo "  run        - Run the project"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  lint       - Run linter"
	@echo "  install    - Install binary"