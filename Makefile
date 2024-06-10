#!/usr/bin/make
GO111MODULE = on
APP_NAME = stocms
CMD_PATH = ./cmd/stocms
OUT = ./bin

VERSION := $(shell git describe --tags --match "v*" --always | sed 's/-g[a-z0-9]\{7\}//')
BRANCH := $(shell git symbolic-ref -q --short HEAD)
REVISION := $(shell git rev-parse --short HEAD)
BUILT_AT := $(shell date +%FT%T%z)

BUILD_VARS := stocms/internal/helper
LDFLAGS := -ldflags "-X ${BUILD_VARS}.Version=${VERSION} -X ${BUILD_VARS}.Branch=${BRANCH} -X ${BUILD_VARS}.Revision=${REVISION} -X ${BUILD_VARS}.BuiltAt=${BUILT_AT} -s -w"


ifeq ($(debug), 1)
LDFLAGS += -gcflags "-N -l"
endif

generate:
	@go generate -x ./...

build-linux:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath $(LDFLAGS) -o ${OUT}/${APP_NAME} ${CMD_PATH}
	if [ ! -d "${OUT}" ]; then mkdir ${OUT}; fi
	if [ ! -f "${OUT}/config.yaml" ]; then cp -r ./configs/config.yaml ${OUT}; fi

build:generate
	@go build -trimpath $(LDFLAGS) -o ${OUT}/${APP_NAME} ${CMD_PATH}

swagger:
	@swag init --parseDependency --parseInternal --parseDepth 1 -g ${CMD_PATH}/main.go -o ./docs

run:
	@go run ${CMD_PATH}

optimize:build-linux
	@upx -9 ${OUT}/${APP_NAME}
