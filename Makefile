.PHONY: build test clean install help

# Default target
all: build

# Build the binary
build:
	go build -o bin/i18nedt cmd/i18nedt/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

# Install the binary to /usr/local/bin
install: build
	cp bin/i18nedt /usr/local/bin/

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the i18nedt binary"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  install        - Install i18nedt to /usr/local/bin"
	@echo "  clean          - Clean build artifacts"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Usage examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make test-coverage"
	@echo "  make install"
	@echo "  i18nedt -k home.welcome src/locales/{zh-CN,en-US}.json"