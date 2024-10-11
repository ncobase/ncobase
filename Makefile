#!/usr/bin/make

GO111MODULE := on
APP_NAME := ncobase
CMD_PATH := ./cmd/ncobase
OUT := ./bin
PLUGIN_PATH := ./plugin

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

.PHONY: install generate copy-config build-multi build-plugins build-plugin build-all swagger build run optimize clean version help

install:
	@go install github.com/swaggo/swag/cmd/swag@latest

generate:
	@go generate -x ./...

copy-config:
	@mkdir -p $(OUT)
	@if [ ! -f "$(OUT)/config.yaml" ]; then cp ./infra/config/config.yaml $(OUT)/config.yaml; fi

build-multi: generate
	@mkdir -p $(OUT)
	$(foreach platform,$(PLATFORMS),$(call build_binary,$(platform)))
	@$(MAKE) copy-config

build-plugins:
	@mkdir -p $(OUT)/plugins
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
						*) EXT="so" ;; \
					esac; \
					echo "Building plugin $$PLUGIN_NAME for $$platform..."; \
					CGO_ENABLED=1 GOOS=$$GOOS GOARCH=$$GOARCH go build -buildmode=plugin $(BUILD_FLAGS) -o $(OUT)/plugins/$$PLUGIN_NAME-$$platform.$$EXT $$dir/cmd/plugin.go || exit 1; \
				done \
			fi \
		done \
	fi

build-all: build-multi build-plugins

swagger:
	@swag init --parseDependency --parseInternal --parseDepth 1 -g $(CMD_PATH)/main.go -o ./docs

build: generate
	@go build $(BUILD_FLAGS) -o $(OUT)/$(APP_NAME) $(CMD_PATH)

build-plugin: build
	@mkdir -p $(OUT)/plugins
	@for dir in $(PLUGIN_PATH)/*; do \
		if [ -d "$$dir" ] && [ -f "$$dir/cmd/plugin.go" ]; then \
			echo "Building plugin $$dir for current platform..."; \
			PLUGIN_NAME=$$(basename $$dir); \
			GOOS=$$(go env GOOS); \
			GOARCH=$$(go env GOARCH); \
			case $$GOOS in \
				windows) EXT="dll" ;; \
				*) EXT="so" ;; \
			esac; \
			echo "Building plugin $$PLUGIN_NAME for $$GOOS/$$GOARCH..."; \
			go build -buildmode=plugin $(BUILD_FLAGS) -o $(OUT)/plugins/$$PLUGIN_NAME.$$EXT $$dir/cmd/plugin.go || exit 1; \
		fi \
	done

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
	@echo "  build-plugin    - Build plugins and main application for the current platform"
	@echo "  build-multi     - Build for multiple platforms"
	@echo "  build-plugins   - Build plugins for multiple platforms"
	@echo "  build-all       - Build main application and plugins"
	@echo "  clean           - Remove build artifacts"
	@echo "  optimize        - Compress binaries using UPX"
	@echo "  run             - Run the application"
	@echo "  swagger         - Generate Swagger documentation"
	@echo "  version         - Display version information"
	@echo "  help            - Display this help message"
