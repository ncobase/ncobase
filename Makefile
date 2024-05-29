#!/usr/bin/make
GO111MODULE = on
APP_NAME = stocms
CMD_PATH = ./cmd/stocms
OUT = ./bin
GIT_TAG_HEAD := `git describe --tags --match "v*" --always`
BUILD_TIME := `date +%FT%T%z`

LDFLAGS:=-ldflags "-X main.Version=$(GIT_TAG_HEAD) -X main.BuildTime=$(BUILD_TIME)"

ifeq ($(debug), 1)
LDFLAGS+= -gcflags "-N -l"
endif

generate:
	@go generate -x ./...

build-linux:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o ${OUT}/${APP_NAME} ${CMD_PATH}
	if [ ! -d "${OUT}" ]; then mkdir ${OUT}; fi
	if [ ! -f "${OUT}/config.yaml" ]; then cp -r ./configs/config.yaml ${OUT}; fi

build:
	@go build $(LDFLAGS) -o ${OUT}/${APP_NAME} ${CMD_PATH}

swag:
	@swag init --parseDependency --parseInternal --parseDepth 1 -g ${CMD_PATH}/main.go -o ./docs

run:
	@go run ${CMD_PATH}

optimize:build-linux
	@upx -9 ${OUT}/${APP_NAME}
