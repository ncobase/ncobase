#!/usr/bin/make

GO111MODULE := on
APP_NAME := ncobase
CMD_PATH := ./cmd/ncobase
CLI_NAME := nco
CLI_PATH := ./cmd/cli
OUT := ./bin
PLUGIN_PATH := ./plugin
BUSINESS_PATH := ./domain

VERSION := $(shell git describe --tags --match "v*" --always | sed 's/-g[a-z0-9]\{7\}//')
BRANCH := $(shell git symbolic-ref -q --short HEAD)
REVISION := $(shell git rev-parse --short HEAD)
BUILT_AT := $(shell date +%FT%T%z)

BUILD_VARS := ncobase/common/helper
LDFLAGS := -X $(BUILD_VARS).Version=$(VERSION) \
           -X $(BUILD_VARS).Branch=$(BRANCH) \
           -X $(BUILD_VARS).Revision=$(REVISION) \
           -X $(BUILD_VARS).BuiltAt=$(BUILT_AT) \
           -s -w -buildid=

BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

ifeq ($(debug), 1)
BUILD_FLAGS += -gcflags "-N -l"
endif

PLATFORMS := linux-amd64 linux-arm64 darwin-amd64

# Detect if running on an Apple Silicon chip
ifeq ($(shell uname -m), arm64)
APPLE_CHIP := 1
else
APPLE_CHIP := 0
endif

define build_binary
  $(eval GOOS := $(word 1,$(subst -, ,$(1))))
  $(eval GOARCH := $(word 2,$(subst -, ,$(1))))
  $(eval EXT := $(if $(filter windows,$(GOOS)),.exe,))
  @echo "Building for $(1)"
  @CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -o $(OUT)/$(APP_NAME)-$(1)$(EXT) $(CMD_PATH)
endef

.PHONY: install generate copy-config build-multi build-plugins build-plugin build-business build-all swagger build run optimize clean version help

install:
	@go install github.com/swaggo/swag/cmd/swag@latest

generate:
	@go generate -x ./...

copy-config:
	@mkdir -p $(OUT)
	@if [ ! -f "$(OUT)/config.yaml" ]; then cp ./setup/config/config.yaml $(OUT)/config.yaml; fi

check-tools:
	@which go >/dev/null 2>&1 || (echo "go is required but not installed" && exit 1)
	@if [ $(APPLE_CHIP) -eq 0 ]; then \
		which gcc >/dev/null 2>&1 || (echo "gcc is required but not installed" && exit 1); \
	fi

build-multi: generate
	@mkdir -p $(OUT)
	$(foreach platform,$(PLATFORMS),$(call build_binary,$(platform)))
	@$(MAKE) copy-config

build-business: check-tools
	@mkdir -p $(OUT)/extension
	@if [ $(APPLE_CHIP) -eq 1 ]; then \
		echo "Skipping multi-platform extension build on Apple Silicon"; \
	else \
		for dir in $(BUSINESS_PATH)/*; do \
			if [ -d "$$dir" ] && [ -f "$$dir/cmd/main.go" ]; then \
				BUSINESS_NAME=$$(basename $$dir); \
				for platform in $(PLATFORMS); do \
					GOOS=$$(echo $$platform | cut -d'-' -f1); \
					GOARCH=$$(echo $$platform | cut -d'-' -f2); \
					case $$GOOS in \
						windows) EXT="dll" ;; \
						darwin) EXT="dylib" ;; \
						*) EXT="so" ;; \
					esac; \
					echo "Building extension $$BUSINESS_NAME for $$platform..."; \
					CGO_ENABLED=1 GOOS=$$GOOS GOARCH=$$GOARCH go build -buildmode=plugin $(BUILD_FLAGS) \
						-o $(OUT)/extension/$$BUSINESS_NAME-$$platform.$$EXT $$dir/cmd/main.go || exit 1; \
				done \
			fi \
		done \
	fi

build-plugins: check-tools
	@mkdir -p $(OUT)/extension/plugins
	@if [ $(APPLE_CHIP) -eq 1 ]; then \
		echo "Skipping multi-platform plugin build on Apple Silicon"; \
	else \
		for dir in $(PLUGIN_PATH)/*; do \
			if [ -d "$$dir" ] && [ -f "$$dir/cmd/plugin.go" ]; then \
				PLUGIN_NAME=$$(basename $$dir); \
				for platform in $(PLATFORMS); do \
					GOOS=$$(echo $$platform | cut -d'-' -f1); \
					GOARCH=$$(echo $$platform | cut -d'-' -f2); \
					case $$GOOS in \
						windows) EXT="dll" ;; \
						darwin) EXT="dylib" ;; \
						*) EXT="so" ;; \
					esac; \
					echo "Building plugin $$PLUGIN_NAME for $$platform..."; \
					CGO_ENABLED=1 GOOS=$$GOOS GOARCH=$$GOARCH go build -buildmode=plugin $(BUILD_FLAGS) \
						-o $(OUT)/extension/plugins/$$PLUGIN_NAME-$$platform.$$EXT $$dir/cmd/plugin.go || exit 1; \
				done \
			fi \
		done \
	fi

build-plugin: check-tools
	@mkdir -p $(OUT)/extension/plugins
	@for dir in $(PLUGIN_PATH)/*; do \
		if [ -d "$$dir" ] && [ -f "$$dir/cmd/plugin.go" ]; then \
			echo "Building plugin $$dir for current platform..."; \
			PLUGIN_NAME=$$(basename $$dir); \
			GOOS=$$(go env GOOS); \
			GOARCH=$$(go env GOARCH); \
			case $$GOOS in \
				windows) EXT="dll" ;; \
				darwin) EXT="dylib" ;; \
				*) EXT="so" ;; \
			esac; \
			echo "Building plugin $$PLUGIN_NAME for $$GOOS/$$GOARCH..."; \
			CGO_ENABLED=1 go build -buildmode=plugin $(BUILD_FLAGS) \
				-o $(OUT)/extension/plugins/$$PLUGIN_NAME.$$EXT $$dir/cmd/plugin.go || exit 1; \
		fi \
	done

build-all: check-tools build-multi build-plugins build-business

swagger:
	@swag init --parseDependency --parseInternal --parseDepth 1 -g $(CMD_PATH)/main.go -o ./docs/swagger

build: generate
	@go build $(BUILD_FLAGS) -o $(OUT)/$(APP_NAME) $(CMD_PATH)

build-cli:
	@go build $(BUILD_FLAGS) -o $(OUT)/$(CLI_NAME) $(CLI_PATH)

run:
	@go run $(CMD_PATH)

optimize: build-multi
	@command -v upx >/dev/null 2>&1 || { echo >&2 "upx is required but not installed. Aborting."; exit 1; }
	@for file in $(OUT)/$(APP_NAME)-*; do \
		if [ "$$(uname -m)" != "arm64" ] || [ ! "$${file}" = *"-darwin-arm64"* ]; then \
			upx -9 "$$file"; \
		else \
			echo "Skipping UPX compression for ARM64 binary on ARM64 Mac: $$file"; \
		fi; \
	done

clean:
	@rm -rf $(OUT)
	@echo "Cleaned build artifacts"

version:
	@echo "Version: $(VERSION)"
	@echo "Branch: $(BRANCH)"
	@echo "Revision: $(REVISION)"
	@echo "Built at: $(BUILT_AT)"

help:
	@echo "Available targets:"
	@echo "  build           - Build for the current platform"
	@echo "  build-plugin    - Build plugins for current platform"
	@echo "  build-multi     - Build for multiple platforms"
	@echo "  build-plugins   - Build plugins for multiple platforms"
	@echo "  build-business  - Build business extensions for multiple platforms"
	@echo "  build-all       - Build main application and plugins"
	@echo "  clean           - Remove build artifacts"
	@echo "  optimize        - Compress binaries using UPX"
	@echo "  run             - Run the application"
	@echo "  swagger         - Generate Swagger documentation"
	@echo "  version         - Display version information"
	@echo "  help            - Display this help message"
