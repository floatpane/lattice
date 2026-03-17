.PHONY: build test run clean lint fmt vet build-full generate_screenshots

BINARY_NAME=lattice

build:
	go build -o $(BINARY_NAME) .

build-full:
	@echo "Building with version information..."
	@VERSION=$$(git describe --tags --abbrev=0 2>/dev/null || echo "dev"); \
	COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	DATE=$$(date +%Y-%m-%d); \
	echo "Version: $$VERSION"; \
	echo "Commit: $$COMMIT"; \
	echo "Date: $$DATE"; \
	go build -ldflags="-X 'main.version=$$VERSION' -X 'main.commit=$$COMMIT' -X 'main.date=$$DATE'" -o $(BINARY_NAME)-full .;

run:
	go run .

test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet

all: lint test build
