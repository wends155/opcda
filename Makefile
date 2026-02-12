.PHONY: test integration help all

# Default target
all: test

# Run unit tests (mocked, environment-agnostic)
test:
	go test -v ./...

# Run integration tests (requires Windows + OPC Server)
integration:
	go test -v -tags integration ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  make test        - Run unit tests (mocked, environment-agnostic)"
	@echo "  make integration - Run integration tests (requires Windows + OPC Server)"
	@echo "  make help        - Show this help message"
