# QuantAlpha HFT - Project Lifecycle Management

.PHONY: help build up down restart logs ps clean smoke-test

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build all service images
	docker-compose build

up: ## Start all services in background
	docker-compose up -d

down: ## Stop and remove containers
	docker-compose down

restart: ## Restart all services
	docker-compose restart

logs: ## Follow all service logs
	docker-compose logs -f

ps: ## List running services
	docker-compose ps

clean: ## Remove all containers, networks, and volumes
	docker-compose down -v --remove-orphans

smoke-test: ## Run basic end-to-end health check
	@echo "Checking backend health..."
	@curl -s -f http://localhost:8080/health || (echo "Backend health check failed" && exit 1)
	@echo "Checking backend readiness..."
	@curl -s -f http://localhost:8080/ready || (echo "Backend readiness check failed" && exit 1)
	@echo "Checking frontend availability..."
	@curl -s -f http://localhost:3000 || (echo "Frontend check failed" && exit 1)
	@echo "Smoke test PASSED"
