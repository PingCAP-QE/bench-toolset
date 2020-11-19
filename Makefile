GO=GO15VENDOREXPERIMENT="1" CGO_ENABLED=0 GO111MODULE=on go

FILES_TO_FMT  := $(shell find . -path -prune -o -name '*.go' -print)

IMG ?= 5kbpers/stability_test:latest

all: format build

format: vet fmt

fmt:
	@echo "gofmt"
	@gofmt -w ${FILES_TO_FMT}
	@git diff --exit-code .

build: mod
	$(GO) build -o bin/stability_test

vet:
	go vet ./...

mod:
	@echo "go mod tidy"
	GO111MODULE=on go mod tidy
	@git diff --exit-code -- go.sum go.mod

docker-build:
	docker build . -t $(IMG)

docker-push:
	docker push $(IMG)

