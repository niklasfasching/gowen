.PHONY: default
default: test build

.PHONY: generate
generate: install
	go generate ./...

.PHONY: test
test: generate
	go test ./... -v

.PHONY: build
build: generate
	go build cmd/gowen/*

.PHONY: install
install:
	go get ./...
