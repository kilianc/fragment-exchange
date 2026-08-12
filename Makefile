.DEFAULT_GOAL := help

GSX ?= gsx

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

.PHONY: dev
dev: gen ## Run the site on http://localhost:8080
	go run ./cmd/server

.PHONY: gen
gen: ## Regenerate .gsx.go from .gsx
	$(GSX) ./web

.PHONY: test
test: gen ## Run every test, including the browser suite
	go test ./...

.PHONY: test-one
test-one: gen ## Run one browser case: make test-one FILTER=popstate
	FX_FILTER="$(FILTER)" go test -run TestBrowser -v .

.PHONY: test-headed
test-headed: gen ## Run the browser suite in a visible Chrome
	FX_HEADED=1 go test -run TestBrowser -v .

.PHONY: serve-tests
serve-tests: ## Serve the test page so you can open it yourself
	@echo "open http://localhost:8081/fx.test.html"
	@go run ./cmd/testserver

.PHONY: check
check: ## What CI runs: generated files current, formatted, vetted, tested
	$(GSX) -check ./web
	$(GSX) fmt -l ./web
	@test -z "$$(gofmt -l . | grep -v '\.gsx\.go$$')" || { echo "gofmt:"; gofmt -l . | grep -v '\.gsx\.go$$'; exit 1; }
	go vet ./...
	go test -race ./...

.PHONY: fmt
fmt: ## Format Go and GSX sources
	$(GSX) fmt -w ./web
	gofmt -w .

.PHONY: build
build: gen ## Build the site binary
	go build -o bin/server ./cmd/server

.PHONY: tools
tools: ## Install the GSX compiler
	go install github.com/kilianc/gsx/cmd/gsx@latest
