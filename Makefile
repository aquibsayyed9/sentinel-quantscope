# Makefile

.PHONY: build run-api run-marketdata run-ruleengine migrate test lint docker-build clean

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOLINT=golangci-lint

# Binary names
API_BINARY=sentinel-api
MARKETDATA_BINARY=sentinel-marketdata
RULEENGINE_BINARY=sentinel-ruleengine

# Build all binaries
build:
	$(GOBUILD) -o bin/$(API_BINARY) ./cmd/api
	$(GOBUILD) -o bin/$(MARKETDATA_BINARY) ./cmd/marketdata
	$(GOBUILD) -o bin/$(RULEENGINE_BINARY) ./cmd/ruleengine

# Run API server
run-api:
	$(GOCMD) run ./cmd/api

# Run Market Data service
run-marketdata:
	$(GOCMD) run ./cmd/marketdata

# Run Rule Engine service
run-ruleengine:
	$(GOCMD) run ./cmd/ruleengine

# Run database migrations
migrate:
	@echo "Running database migrations..."
	./scripts/migrations.sh up

# Create a new migration file
migrate-create:
	@read -p "Enter migration name: " name; \
	./scripts/migrations.sh create $$name

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -cover ./...

# Lint code
lint:
	$(GOLINT) run ./...

# Tidy and vendor Go modules
deps:
	$(GOMOD) tidy
	$(GOMOD) vendor

# Build Docker images
docker-build:
	docker build -f deployments/docker/Dockerfile.api -t sentinel-api .
	docker build -f deployments/docker/Dockerfile.marketdata -t sentinel-marketdata .
	docker build -f deployments/docker/Dockerfile.ruleengine -t sentinel-ruleengine .

# Run services with Docker Compose
docker-up:
	docker-compose -f deployments/docker-compose.yml up -d

# Stop Docker Compose services
docker-down:
	docker-compose -f deployments/docker-compose.yml down

# Clean build artifacts
clean:
	rm -rf bin/*
	rm -rf vendor/