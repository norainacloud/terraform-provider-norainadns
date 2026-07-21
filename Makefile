# Copyright (c) Noraina
# SPDX-License-Identifier: MPL-2.0

HOSTNAME=registry.terraform.io
NAMESPACE=norainacloud
NAME=norainadns
BINARY=terraform-provider-${NAME}
VERSION=0.1.0
OS_ARCH=linux_amd64

default: install

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

docs:
	go generate ./...

clean:
	rm -f ${BINARY}
	rm -rf ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}

.PHONY: build install test testacc fmt lint docs clean
