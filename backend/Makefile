.PHONY: build run test test-race fmt vet clean docker-build docker-run help

BINARY_NAME=syncraft-api
BIN_DIR=bin
MAIN_PACKAGE=./cmd/api

help:
	@echo "Available targets:"
	@echo "  build        Compile API server binary"
	@echo "  run          Run API server locally"
	@echo "  test         Run all unit and integration tests"
	@echo "  test-race    Run tests with race detector enabled"
	@echo "  fmt          Format source code using gofmt"
	@echo "  vet          Run go vet static analysis"
	@echo "  clean        Remove compiled binaries"
	@echo "  docker-build Build Docker container image"
	@echo "  docker-run   Run Docker container on port 8080"

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

run:
	go run $(MAIN_PACKAGE)

test:
	go test -v ./...

test-race:
	go test -race -v ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)

docker-build:
	docker build -t syncraft-backend:latest .

docker-run:
	docker run -p 8080:8080 --env-file .env.example syncraft-backend:latest
