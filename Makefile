.PHONY: test integration help all

# Default target
all: test

# Run unit tests (mocked, environment-agnostic)
test:
	@mkdir -p logs
	go test -v ./... > logs/test.log 2>&1

# Run integration tests (requires Windows + OPC Server)
integration:
	@mkdir -p logs
	go test -v -tags integration ./... > logs/integration.log 2>&1

# Show help
help:
	@echo "Available targets:"
	@echo "  make test        - Run unit tests (mocked, environment-agnostic)"
	@echo "  make integration - Run integration tests (requires Windows + OPC Server)"
	@echo "  make help        - Show this help message"
