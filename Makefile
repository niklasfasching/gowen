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

.PHONY: generate-gh-pages
generate-gh-pages: install
	cp -r etc/gh-pages .
	cp $(shell go env GOROOT)/misc/wasm/wasm_exec.js gh-pages/wasm_exec.js
	mv gh-pages/_js.go gh-pages/js.go
	go get github.com/kr/pretty
	GOOS=js GOARCH=wasm go build -o gh-pages/main.wasm gh-pages/js.go
	rm gh-pages/js.go
