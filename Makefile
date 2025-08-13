
.PHONY: build test run clean docker-build docker-up swagger

BINARY_NAME=server
BUILD_DIR=bin
GO_VERSION=1.24.3
LINTER_CONFIG=.golangci.yml

build:
	@echo "Building application..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

test:
	@echo "Running tests..."
	@go test -v ./...

run: build
	@echo "Starting application..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)

docker-build:
	@echo "Building Docker image..."
	@docker-compose build --no-cache

docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up -d

swagger:
	@echo "Generating Swagger documentation..."
	@swag init --parseDependency --parseInternal -g ./cmd/server/main.go
