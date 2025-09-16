.PHONY: help build run test clean swagger docker-build docker-run

# Default target
help:
	@echo "Available commands:"
	@echo "  build          - Build the application binary"
	@echo "  run            - Run the application"
	@echo "  test           - Run all tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  swagger        - Generate Swagger documentation"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker Compose"
	@echo ""
	@echo "  docker-logs-app- Show app logs (including auto-migration logs)"

# Build the application
build:
	@echo "Building item-sync application..."
	go build -o bin/item-sync .

# Run the application
run:
	@echo "Running item-sync application..."
	go run .

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf docs/

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	swag init --parseDependency
	@echo "Swagger documentation generated in docs/ directory"
	@echo "API docs will be available at http://localhost:8080/swagger/index.html"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t item-sync:latest .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	docker compose up -d
	@echo "Services started. API available at http://localhost:8080"
	@echo "Swagger UI available at http://localhost:8080/swagger/index.html"

# Stop Docker Compose services
docker-stop:
	@echo "Stopping Docker Compose services..."
	docker compose down

# Show Docker Compose logs
docker-logs:
	docker compose logs -f

# Docker commands (migrations run automatically on app startup)
docker-logs-app:
	@echo "Showing application logs (including migration logs)..."
	docker-compose logs -f app

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Development tools installed"