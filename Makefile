.PHONY: help dev build test lint clean migrate-up migrate-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Run the server (requires Postgres + Redis)
	go run ./cmd/server

dev-worker: ## Run the background worker
	go run ./cmd/worker

build: ## Build server binary
	go build -o bin/server ./cmd/server

build-worker: ## Build worker binary
	go build -o bin/worker ./cmd/worker

test: ## Run all tests with race detector
	go test -race ./...

test-cover: ## Run tests with coverage
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint: ## Lint Go code
	golangci-lint run ./...

lint-fix: ## Auto-fix lint issues
	golangci-lint run --fix ./...

vet: ## Run go vet
	go vet ./...

tidy: ## Run go mod tidy
	go mod tidy

migrate-up: ## Run database migrations up
	migrate -path internal/db/migrations -database "$${DATABASE_URL}" up

migrate-down: ## Rollback last database migration
	migrate -path internal/db/migrations -database "$${DATABASE_URL}" down 1

clean: ## Remove build artifacts
	rm -rf bin/
