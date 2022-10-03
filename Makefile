SHELL = /bin/bash -e -o pipefail
VERSION=0.1.1
GIT_COMMIT=$(git rev-parse --short HEAD)
LDFLAGS=-ldflags "-s -w -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT}"
BUILD_DIR ?= .build
BASE_PKG_NAME=asana-proxy

help:
	@echo "VERSION: ${VERSION}"
	@echo "GIT_COMMIT: ${GIT_COMMIT}"

build: help
	@[ -d ${BUILD_DIR} ] || mkdir -p ${BUILD_DIR}
	CGO_ENABLED=0 go build ${LDFLAGS} -o ${BUILD_DIR}/${BASE_PKG_NAME} cmd/${BASE_PKG_NAME}/main.go
	@file  ${BUILD_DIR}/${BASE_PKG_NAME}
	@du -h ${BUILD_DIR}/${BASE_PKG_NAME}

run:
	ASANA_URL=https://app.asana.com/api/1.0 \
    SERVER_ADDR=:8089 \
    .build/asana-proxy
