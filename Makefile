.PHONY: help up down build logs clean test backend-dev frontend-dev

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

up: ## Start all services with docker-compose
	docker-compose up -d

down: ## Stop all services
	docker-compose down

build: ## Build all services
	docker-compose build

logs: ## Show logs from all services
	docker-compose logs -f

clean: ## Remove all containers, volumes, and images
	docker-compose down -v --rmi all

restart: ## Restart all services
	docker-compose restart

ps: ## Show running containers
	docker-compose ps

backend-dev: ## Run backend in development mode (requires local Go installation)
	cd backend && go run cmd/main.go

backend-deps: ## Install backend dependencies
	cd backend && go mod download

backend-test: ## Run backend tests
	cd backend && go test -v ./...

frontend-dev: ## Serve frontend locally (requires Python)
	cd frontend && python3 -m http.server 8080

# Database operations
db-migrate: ## Run database migrations
	@echo "Migrations are handled automatically by GORM AutoMigrate"

db-reset: ## Reset database (WARNING: deletes all data)
	docker-compose down -v postgres
	docker-compose up -d postgres

# Monitoring
health: ## Check health of all services
	@echo "Backend health:"
	@curl -s http://localhost:3000/health | jq . || echo "Backend not responding"
	@echo "\nRedis health:"
	@docker-compose exec redis redis-cli ping || echo "Redis not responding"
	@echo "\nPostgreSQL health:"
	@docker-compose exec postgres pg_isready -U testbox || echo "PostgreSQL not responding"
	@echo "\nRabbitMQ health:"
	@curl -s -u guest:guest http://localhost:15672/api/health/checks/alarms || echo "RabbitMQ not responding"

# Utilities
install-tools: ## Install required development tools
	@echo "Installing Go dependencies..."
	cd backend && go mod download
	@echo "Done!"
