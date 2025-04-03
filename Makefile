#!/usr/bin/make

# Configuration
GO111MODULE := on
APP_NAME := ncobase
CMD_PATH := ./cmd/ncobase
CLI_NAME := nco
CLI_PATH := ./cmd/cli
OUT := ./bin
PLUGIN_PATH := ./plugin
BUSINESS_PATH := ./domain

# Version with fallback
VERSION := $(shell git describe --tags --match "v*" --always 2>/dev/null || echo "unknown")
BRANCH := $(shell git symbolic-ref -q --short HEAD 2>/dev/null || echo "unknown")
REVISION := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILT_AT := $(shell date +%FT%T%z)

# Build flags
BUILD_VARS := ncobase/ncore/helper
LDFLAGS := -X $(BUILD_VARS).Version=$(VERSION) \
           -X $(BUILD_VARS).Branch=$(BRANCH) \
           -X $(BUILD_VARS).Revision=$(REVISION) \
           -X $(BUILD_VARS).BuiltAt=$(BUILT_AT) \
           -s -w -buildid=

BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

# Debug mode
ifeq ($(debug), 1)
BUILD_FLAGS += -gcflags "-N -l"
endif

# Platform specifics
SUPPORTED_OS := linux darwin windows
SUPPORTED_ARCH := amd64 arm64

# Generate all valid platform combinations
PLATFORMS := $(foreach os,$(SUPPORTED_OS),$(foreach arch,$(SUPPORTED_ARCH),$(if $(filter $(os)-$(arch),windows-arm64),,$(os)-$(arch))))

# Detect host platform
ifeq ($(OS),Windows_NT)
    HOST_OS := windows
    ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
        HOST_ARCH := amd64
    endif
    ifeq ($(PROCESSOR_ARCHITECTURE),x86)
        HOST_ARCH := 386
    endif
else
    HOST_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
    HOST_ARCH := $(shell uname -m)
    ifeq ($(HOST_ARCH),x86_64)
        HOST_ARCH := amd64
    endif
    ifeq ($(HOST_ARCH),aarch64)
        HOST_ARCH := arm64
    endif
endif

# Build binary function with enhanced error handling
define build_binary
  $(eval GOOS := $(word 1,$(subst -, ,$(1))))
  $(eval GOARCH := $(word 2,$(subst -, ,$(1))))
  $(eval EXT := $(if $(filter windows,$(GOOS)),.exe,))
  @echo "- Building for $(GOOS)/$(GOARCH)"
  @if ! command -v go >/dev/null 2>&1; then \
    echo "Error: go is not installed"; \
    exit 1; \
  fi
  @if [ "$(GOOS)" = "$(HOST_OS)" ] && [ "$(GOARCH)" = "$(HOST_ARCH)" ]; then \
    if [ "$(GOOS)" != "darwin" ] || [ "$(GOARCH)" != "arm64" ]; then \
      if ! command -v gcc >/dev/null 2>&1; then \
        echo "Error: gcc is required for native build"; \
        exit 1; \
      fi; \
    fi; \
    CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) \
      -o $(OUT)/$(APP_NAME)-$(1)$(EXT) $(CMD_PATH) || \
      (echo "Error: Failed to build for $(GOOS)/$(GOARCH)" && exit 1); \
  else \
    CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) \
      -o $(OUT)/$(APP_NAME)-$(1)$(EXT) $(CMD_PATH) || \
      (echo "Error: Failed to build for $(GOOS)/$(GOARCH)" && exit 1); \
  fi
endef

# Build extensions function with enhanced platform check
define build_extensions
    @echo "- Building $(1) for $(HOST_OS)/$(HOST_ARCH)"
    @if [ "$(HOST_OS)" = "windows" ]; then \
        echo "Warning: Plugin building is not supported on Windows"; \
        exit 0; \
    fi
    @if ! command -v gcc >/dev/null 2>&1; then \
        echo "Error: gcc is required for plugin builds"; \
        exit 1; \
    fi
    @if [ "$(1)" = "plugins" ]; then \
        mkdir -p $(OUT)/extension/plugins; \
        OUTPUT_DIR="$(OUT)/extension/plugins"; \
    else \
        mkdir -p $(OUT)/extension; \
        OUTPUT_DIR="$(OUT)/extension"; \
    fi; \
    for dir in $($2)/*; do \
        if [ -d "$$dir" ] && [ -f "$$dir/cmd/main.go" ]; then \
            NAME=$$(basename $$dir); \
            case $(HOST_OS) in \
                windows) EXT="dll" ;; \
                darwin) EXT="dylib" ;; \
                *) EXT="so" ;; \
            esac; \
            echo "  Building $(1) $$NAME..."; \
            if ! CGO_ENABLED=1 GOOS=$(HOST_OS) GOARCH=$(HOST_ARCH) go build -buildmode=plugin \
                $(BUILD_FLAGS) \
                -o $$OUTPUT_DIR/$$NAME.$$EXT \
                $$dir/cmd/main.go; then \
                echo "Error: Failed to build $$NAME for $(HOST_OS)/$(HOST_ARCH)"; \
                exit 1; \
            fi; \
            echo "  Successfully built $$NAME.$$EXT"; \
        fi; \
    done
endef

.PHONY: install generate copy-config check-tools build-multi build-business build-plugins build-all swagger build build-cli run optimize clean version help

# Install required tools
install:
	@echo "Installing Required Tools"
	@if ! command -v go >/dev/null 2>&1; then \
		echo "Error: go is required but not installed"; \
		exit 1; \
	fi
	@go install github.com/swaggo/swag/cmd/swag@latest || \
		(echo "Error: Failed to install swag" && exit 1)
	@echo "✓ Successfully installed required tools"

# Generate code
generate:
	@echo "Generating Code"
	@go generate -x ./... || (echo "Error: Code generation failed" && exit 1)
	@echo "✓ Code generation completed"

# Copy config file with checks
copy-config:
	@mkdir -p $(OUT)
	@if [ ! -f "$(OUT)/config.yaml" ]; then \
		if [ -f "./setup/config/config.yaml" ]; then \
			echo "- Copying config file..."; \
			cp ./setup/config/config.yaml $(OUT)/config.yaml || \
				(echo "Error: Failed to copy config file" && exit 1); \
			echo "✓ Config file copied"; \
		else \
			echo "Error: Source config file not found"; \
			exit 1; \
		fi; \
	else \
		echo "• Config file already exists"; \
	fi

# Check required tools
check-tools:
	@echo "Checking Required Tools"
	@echo "• Checking Go installation..."
	@if ! command -v go >/dev/null 2>&1; then \
		echo "Error: go is required but not installed"; \
		exit 1; \
	fi
	@if [ "$(HOST_OS)" != "darwin" ] || [ "$(HOST_ARCH)" != "arm64" ]; then \
		echo "• Checking GCC installation..."; \
		if ! command -v gcc >/dev/null 2>&1; then \
			echo "Error: gcc is required but not installed"; \
			exit 1; \
		fi; \
	fi
	@echo "✓ All required tools are available"

# Build for all supported platforms
build-multi: generate
	@echo "Starting Multi-platform Build"
	@mkdir -p $(OUT)
	$(foreach platform,$(PLATFORMS),$(call build_binary,$(platform)))
	@$(MAKE) copy-config
	@echo "✓ Multi-platform build completed"

# Build business extensions
build-business: check-tools
	@echo "Building Business Extensions"
	$(call build_extensions,business,BUSINESS_PATH)
	@echo "✓ Business extensions build completed"

# Build plugins
build-plugins: check-tools
	@echo "Building Plugins"
	$(call build_extensions,plugins,PLUGIN_PATH)
	@echo "✓ Plugins build completed"

# Build all components
build-all: check-tools build-multi build-business build-plugins
	@echo "✓ All components built successfully"

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger Documentation"
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Error: swag is required but not installed. Run 'make install' first."; \
		exit 1; \
	fi
	@swag init --parseDependency --parseInternal --parseDepth 1 -g $(CMD_PATH)/main.go -o ./docs/swagger || \
		(echo "Error: Failed to generate Swagger documentation" && exit 1)
	@echo "✓ Swagger documentation generated"

# Build for current platform
build: generate
	@echo "Building for Current Platform"
	@mkdir -p $(OUT)
	@CGO_ENABLED=1 go build $(BUILD_FLAGS) -o $(OUT)/$(APP_NAME) $(CMD_PATH) || \
		(echo "Error: Build failed" && exit 1)
	@echo "✓ Build completed"

# Build CLI tool
build-cli: check-tools
	@echo "Building CLI Tool"
	@mkdir -p $(OUT)
	@CGO_ENABLED=1 go build $(BUILD_FLAGS) -o $(OUT)/$(CLI_NAME) $(CLI_PATH) || \
		(echo "Error: CLI build failed" && exit 1)
	@echo "✓ CLI tool built"

# Run application
run:
	@echo "Running Application"
	@go run $(CMD_PATH)

# Optimize binaries using UPX
optimize: build-multi
	@echo "Optimizing Binaries"
	@if ! command -v upx >/dev/null 2>&1; then \
		echo "Error: upx is required but not installed"; \
		exit 1; \
	fi
	@for file in $(OUT)/$(APP_NAME)-*; do \
		if [ ! -f "$$file" ]; then \
			continue; \
		fi; \
		if [ "$(HOST_ARCH)" = "arm64" ] && [[ "$$file" == *"-darwin-arm64"* ]]; then \
			echo "• Skipping compression for ARM64 binary: $$file"; \
		elif [[ "$$file" == *.exe ]] && [ "$(HOST_OS)" != "windows" ]; then \
			echo "• Skipping compression for Windows binary on non-Windows host: $$file"; \
		else \
			echo "- Compressing $$file..."; \
			upx -9 "$$file" || echo "Warning: Failed to compress $$file"; \
		fi; \
	done
	@echo "✓ Binary optimization completed"

# Clean build artifacts with safety checks
clean:
	@echo "Cleaning Build Artifacts"
	@if [ -d "$(OUT)" ]; then \
		rm -rf $(OUT) && \
		echo "✓ Build artifacts cleaned" || \
		(echo "Error: Failed to clean build artifacts" && exit 1); \
	else \
		echo "• Nothing to clean"; \
	fi

# Show version information
version:
	@echo "Version Information"
	@echo "• Version:   $(VERSION)"
	@echo "• Branch:    $(BRANCH)"
	@echo "• Revision:  $(REVISION)"
	@echo "• Built at:  $(BUILT_AT)"

# Show help
help:
	@echo "Build System Help"
	@echo ""
	@echo "Build Targets:"
	@echo "  • Main Builds:"
	@echo "    make build          - Build for current platform"
	@echo "    make build-multi    - Build for all supported platforms"
	@echo "    make build-all      - Build all components"
	@echo ""
	@echo "  • Component Builds:"
	@echo "    make build-business - Build business extensions"
	@echo "    make build-plugins  - Build plugins"
	@echo "    make build-cli      - Build CLI tool"
	@echo ""
	@echo "  • Development:"
	@echo "    make install        - Install required tools"
	@echo "    make generate       - Generate code"
	@echo "    make run            - Run application"
	@echo "    make swagger        - Generate Swagger docs"
	@echo ""
	@echo "  • Maintenance:"
	@echo "    make clean          - Remove build artifacts"
	@echo "    make optimize       - Compress binaries (requires upx)"
	@echo "    make version        - Show version info"
	@echo ""
	@echo "Build Options:"
	@echo "  debug=1              - Enable debug symbols"
	@echo ""
	@echo "Environment Info:"
	@echo "  • OS:           $(HOST_OS)"
	@echo "  • Architecture: $(HOST_ARCH)"
	@echo "  • Supported platforms: $(PLATFORMS)"
