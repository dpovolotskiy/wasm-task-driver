PLUGIN_BINARY=build/wasm-task-driver
export GO111MODULE=on

default: clean go-mod-tidy lint build

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf ${PLUGIN_BINARY}

build: go-mod-tidy
	go build -o ${PLUGIN_BINARY} .

lint:
	@golangci-lint run --timeout 15m

go-mod-tidy:
	@go mod tidy -v
