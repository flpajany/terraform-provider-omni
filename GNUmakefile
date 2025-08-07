NAME=omni
VERSION=0.1.8
OS_ARCH=$(shell go version | cut -d' ' -f 4 | sed -e 's/\//_/')
ifeq ($(OS_ARCH),windows_amd64)
	BINARY=terraform-provider-$(NAME).exe
	TF_PLUGIN_DIR=${APPDATA}/terraform.d
else
	BINARY=terraform-provider-$(NAME)
	TF_PLUGIN_DIR=~/.terraform.d
endif

#default: fmt install generate
default: install-local

build:
# go build -v ./...
	go build -o ${BINARY}

install: build
	go install -v ./...

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: fmt lint test testacc build install generate

