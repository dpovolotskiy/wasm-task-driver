PLUGIN_BINARY=build/wasm-task-driver
PKGS = $(shell go list ./... | grep -v vendor)

default: clean go-mod-tidy lint test build

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf ${PLUGIN_BINARY}

build: go-mod-tidy
	go build -o ${PLUGIN_BINARY} .

lint:
	@golangci-lint run --timeout 15m

go-mod-tidy:
	@go mod tidy -v

test:
	@go test -race -coverprofile=coverage.txt -covermode=atomic $(PKGS)
	@go tool cover -html=coverage.txt -o coverage.html
