PLUGIN_BINARY=build/wasm-task-driver
export GO111MODULE=on

default: build

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf ${PLUGIN_BINARY}

build: go-mod-tidy
	go build -o ${PLUGIN_BINARY} .

lint:
	@golangci-lint run --timeout 15m

go-mod-tidy:
	@report=`go mod tidy -v 2>&1` ; if [ -n "$$report" ]; then echo "$$report"; exit 1; fi
