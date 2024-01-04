.PHONY: build test setup lint

build:
	CGO_ENABLED=0 go build -o arcade cmd/arcade/arcade.go

test:
	go generate ./...
	go test ./...

setup:
	go get -v -t -d ./...
	if [ -f go.mod ]; then \
		go mod tidy; \
		go mod verify; \
	fi; \
	if [ -f Gopkg.toml ]; then \
		curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh; \
		dep ensure; \
	fi; \
	if ! command -v golangci-lint > /dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest; \
	else \
		echo "golangci-lint is already installed."; \
	fi

lint:
	golangci-lint run --skip-files .*_test.go --enable wsl --enable misspell --timeout 180s
