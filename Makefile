.PHONY: build test run clean docker-build docker-run help

# Variables
BINARY_NAME=ussc
CMD_PATH=./cmd/ussc
DOCKER_IMAGE=udp-receiver
DOCKER_TAG=latest

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: bin/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Test coverage report: coverage.html"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	@go run $(CMD_PATH)/main.go -config ./cmd/ussc/config.yml.example

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./... || echo "Install golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run --rm -p 8080:8080 -p 8830:8830/udp -p 8831:8831/udp \
		-v $(PWD)/cmd/ussc/config.yml:/app/config.yml \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Build and run Docker container
docker: docker-build docker-run

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  test           - Run tests with coverage report"
	@echo "  test-coverage  - Run tests and show coverage summary"
	@echo "  run            - Run the application"
	@echo "  clean          - Clean build artifacts"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  deps           - Install dependencies"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker         - Build and run Docker container"
	@echo "  help           - Show this help message"

