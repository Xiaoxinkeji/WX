.PHONY: all build test clean run install-deps

all: test build

install-deps:
	go mod download
	go mod tidy

build:
	go build -o bin/wechat-assistant ./cmd/app

test:
	go test ./... -v -coverprofile=coverage.out -covermode=atomic

test-coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

run:
	go run ./cmd/app

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	gofmt -s -w .

vet:
	go vet ./...
