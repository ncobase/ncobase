#!/usr/bin/make
GO111MODULE = on
APP_NAME = ncobase
CMD_PATH = ./cmd/ncobase
OUT = ./bin

VERSION := $(shell git describe --tags --match "v*" --always | sed 's/-g[a-z0-9]\{7\}//')
BRANCH := $(shell git symbolic-ref -q --short HEAD)
REVISION := $(shell git rev-parse --short HEAD)
BUILT_AT := $(shell date +%FT%T%z)

BUILD_VARS := ncobase/internal/helper
LDFLAGS := -ldflags "-X ${BUILD_VARS}.Version=${VERSION} -X ${BUILD_VARS}.Branch=${BRANCH} -X ${BUILD_VARS}.Revision=${REVISION} -X ${BUILD_VARS}.BuiltAt=${BUILT_AT} -s -w"

ifeq ($(debug), 1)
LDFLAGS += -gcflags "-N -l"
endif

generate:
	@go generate -x ./...

copy-config:
	@mkdir -p ${OUT}
	@if [ ! -f "${OUT}/config.yaml" ]; then cp ./configs/config.yaml ${OUT}; fi

build-multi: build-linux build-windows build-macos-amd64 build-macos-arm64 build-arm
	@make copy-config

build-linux:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath $(LDFLAGS) -o ${OUT}/${APP_NAME}-linux-amd64 ${CMD_PATH}

build-windows:
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath $(LDFLAGS) -o ${OUT}/${APP_NAME}-windows-amd64.exe ${CMD_PATH}

build-macos-amd64:
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath $(LDFLAGS) -o ${OUT}/${APP_NAME}-darwin-amd64 ${CMD_PATH}

build-macos-arm64:
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath $(LDFLAGS) -o ${OUT}/${APP_NAME}-darwin-arm64 ${CMD_PATH}

build-arm:
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -trimpath $(LDFLAGS) -o ${OUT}/${APP_NAME}-linux-arm ${CMD_PATH}

swagger:
	@swag init --parseDependency --parseInternal --parseDepth 1 -g ${CMD_PATH}/main.go -o ./docs

build: generate
	@go build -trimpath $(LDFLAGS) -o ${OUT}/${APP_NAME} ${CMD_PATH}

run:
	@go run ${CMD_PATH}

optimize: build-multi
	@command -v upx >/dev/null 2>&1 || { echo >&2 "upx is required but not installed. Aborting."; exit 1; }
	upx -9 ${OUT}/${APP_NAME}-linux-amd64
	upx -9 ${OUT}/${APP_NAME}-windows-amd64.exe
	upx -9 ${OUT}/${APP_NAME}-darwin-amd64
	upx -9 ${OUT}/${APP_NAME}-darwin-arm64
	upx -9 ${OUT}/${APP_NAME}-linux-arm

.PHONY: generate copy-config build-multi build-linux build-windows build-macos-amd64 build-macos-arm64 build-arm swagger run optimize
