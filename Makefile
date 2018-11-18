.PHONY: default
default: test build

.PHONY: run
run: generate
	go run cmd/gowen/*

.PHONY: generate
generate: install
	go generate ./...

.PHONY: test
test: generate
	go test ./... -v

.PHONY: build
build: install
	go build cmd/gowen/*

.PHONY: install
install:
	go get ./...

.PHONY: setup
setup: install
	git config core.hooksPath etc/githooks
