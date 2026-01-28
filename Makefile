.PHONY: help build run test clean docker-build docker-run proto

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Orchestrator Service (Master Server)
orchestrator-build: ## Build orchestrator service
	cd orchestrator-service && go mod tidy && go build -o bin/orchestrator cmd/server/main.go

orchestrator-run: ## Run orchestrator service
	cd orchestrator-service && go run cmd/server/main.go

orchestrator-test: ## Test orchestrator service
	cd orchestrator-service && go test ./...

# Agent Service (VPS Nodes)
agent-build: ## Build agent service
	cd agent-service && go mod tidy && go build -o bin/agent cmd/agent/main_full.go

agent-build-simple: ## Build agent service (simple version)
	cd agent-service && go mod tidy && go build -o bin/agent cmd/agent/main_simple.go

agent-run: ## Run agent service
	cd agent-service && go run cmd/agent/main_full.go

agent-test: ## Test agent service
	cd agent-service && go test ./...

# API Service (Existing)
api-build: ## Build API service
	cd api-service && go mod tidy && go build -o bin/server cmd/server/main.go

api-run: ## Run API service
	cd api-service && go run cmd/server/main.go

api-test: ## Test API service
	cd api-service && go test ./...

# Web Service (React UI)
web-install: ## Install web service dependencies
	cd web-service && npm install

web-build: ## Build web service
	cd web-service && npm run build

web-run: ## Run web service in development mode
	cd web-service && npm run dev

web-test: ## Test web service
	cd web-service && npm run test

# Protobuf generation
proto: ## Generate protobuf files
	@echo "Generating protobuf files..."
	@which protoc >/dev/null || (echo "protoc is not installed" && exit 1)
	@mkdir -p orchestrator-service/pkg/proto
	@protoc --go_out=orchestrator-service/pkg/proto --go-grpc_out=orchestrator-service/pkg/proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/node_management.proto
	@echo "✅ Protobuf files generated"

# Docker commands
docker-build: ## Build all Docker images
	docker compose -f deployments/docker/docker-compose.yml build

docker-run: ## Run all services with Docker Compose
	docker-compose -f deployments/docker/docker-compose.yml up -d

docker-stop: ## Stop all Docker services
	docker-compose -f deployments/docker/docker-compose.yml down

docker-logs: ## Show Docker logs
	docker-compose -f deployments/docker/docker-compose.yml logs -f

docker-restart: ## Restart all services
	docker-compose -f deployments/docker/docker-compose.yml restart

docker-status: ## Show Docker service status
	docker-compose -f deployments/docker/docker-compose.yml ps

# Development commands
dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	@make web-install
	@which protoc >/dev/null || (echo "Installing protoc..." && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest)
	@make proto
	@echo "✅ Development environment ready!"

dev-start: ## Start distributed system in development mode
	@echo "🚀 Starting Hysteria2 Distributed VPN System..."
	@./scripts/dev-start.sh

dev-stop: ## Stop development environment
	@echo "🛑 Stopping development environment..."
	@docker-compose -f deployments/docker/docker-compose.yml down

# Database commands
db-up: ## Start only database services
	docker-compose -f deployments/docker/docker-compose.yml up -d postgres redis

db-migrate: ## Run database migrations
	docker-compose -f deployments/docker/docker-compose.yml run --rm orchestrator-service /app/migrate

db-backup: ## Create database backup
	docker exec hysteria2-postgres pg_dump -U hysteria2 hysteria2_db > backup_$(date +%Y%m%d_%H%M%S).sql
	@echo "✅ Database backed up"

db-restore: ## Restore database from backup (usage: make db-restore BACKUP=file.sql)
	@if [ -z "$(BACKUP)" ]; then echo "Usage: make db-restore BACKUP=backup_file.sql"; exit 1; fi
	docker exec -i hysteria2-postgres psql -U hysteria2 hysteria2_db < $(BACKUP)
	@echo "✅ Database restored from $(BACKUP)"

# Testing commands
test-all: ## Run all tests
	@echo "Running all tests..."
	@make orchestrator-test
	@make agent-test
	@make api-test
	@make web-test
	@echo "✅ All tests completed"

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@make orchestrator-test
	@make agent-test
	@make api-test

# Cleanup
clean: ## Clean build artifacts and containers
	@echo "Cleaning up..."
	@docker-compose -f deployments/docker/docker-compose.yml down -v --remove-orphans
	@docker system prune -f
	@rm -rf orchestrator-service/bin/
	@rm -rf agent-service/bin/
	@rm -rf api-service/bin/
	@rm -rf web-service/dist/
	@rm -rf web-service/node_modules/
	@echo "✅ Cleanup completed"

# Production deployment
deploy: ## Deploy to production environment
	@echo "🚀 Deploying to production..."
	@./scripts/deploy.sh production

deploy-staging: ## Deploy to staging environment
	@echo "🚀 Deploying to staging..."
	@./scripts/deploy.sh staging

# Monitoring and logs
logs-orchestrator: ## Show orchestrator logs
	docker-compose -f deployments/docker/docker-compose.yml logs -f orchestrator-service

logs-api: ## Show API logs
	docker-compose -f deployments/docker/docker-compose.yml logs -f api-service

logs-web: ## Show web service logs
	docker-compose -f deployments/docker/docker-compose.yml logs -f web-service

logs-agents: ## Show all agent logs
	docker-compose -f deployments/docker/docker-compose.yml logs -f agent-us-east agent-europe agent-asia

logs-all: ## Show all logs
	docker-compose -f deployments/docker/docker-compose.yml logs -f

# Health checks
health: ## Check health of all services
	@echo "🏥 Checking service health..."
	@docker-compose -f deployments/docker/docker-compose.yml ps
	@echo ""
	@echo "Web Interface: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000 || echo "❌ Down")"
	@echo "API Service: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health || echo "❌ Down")"
	@echo "Orchestrator: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/health || echo "❌ Down")"
	@echo "Database: $(docker exec hysteria2-postgres pg_isready -U hysteria2 >/dev/null 2>&1 && echo "✅ Healthy" || echo "❌ Down")"
	@echo "Redis: $(docker exec hysteria2-redis redis-cli ping >/dev/null 2>&1 && echo "✅ Healthy" || echo "❌ Down")"

# Full development workflow
dev-full: clean proto docker-build docker-run ## Full development setup and start
	@echo "🎉 Full development environment started!"
	@make health

# Production commands
prod-backup: ## Production backup
	@echo "📦 Creating production backup..."
	@./scripts/backup.sh

prod-restore: ## Production restore
	@echo "📦 Restoring production backup..."
	@./scripts/restore.sh

scale-agents: ## Scale VPS agents (usage: make scale-agents COUNT=3)
	@if [ -z "$(COUNT)" ]; then echo "Usage: make scale-agents COUNT=3"; exit 1; fi
	docker-compose -f deployments/docker/docker-compose.yml up -d --scale agent-us-east=$(COUNT) --scale agent-europe=$(COUNT) --scale agent-asia=$(COUNT)

# Obfuscation commands
enable-obfuscation: ## Enable advanced obfuscation on all nodes
	@echo "Enabling advanced obfuscation for DPI bypass..."
	@docker-compose -f deployments/docker/docker-compose.yml exec orchestrator-service /app/enable-obfuscation.sh
	@echo "✅ Advanced obfuscation enabled"

configure-dpi-bypass: ## Configure all settings for Russian DPI bypass
	@echo "Configuring DPI bypass settings..."
	@make enable-obfuscation
	@docker-compose -f deployments/docker/docker-compose.yml exec orchestrator-service /app/configure-dpi-bypass.sh
	@echo "✅ DPI bypass configuration completed"

# Installer generation
generate-installers: ## Generate universal installer scripts
	@echo "Generating installer scripts..."
	@chmod +x scripts/install-central.sh
	@chmod +x scripts/install-node.sh
	@echo "✅ Installer scripts generated and made executable"

# Quick install commands
install-central: generate-installers ## Quick install central server (requires DOMAIN and EMAIL env vars)
	@if [ -z "$(DOMAIN)" ]; then echo "Error: DOMAIN environment variable required"; echo "Usage: make install-central DOMAIN=vpn.example.com EMAIL=admin@example.com"; exit 1; fi
	@if [ -z "$(EMAIL)" ]; then echo "Error: EMAIL environment variable required"; echo "Usage: make install-central DOMAIN=vpn.example.com EMAIL=admin@example.com"; exit 1; fi
	@echo "Installing Hysteria2 Central Server..."
	@DOMAIN=$(DOMAIN) EMAIL=$(EMAIL) bash scripts/install-central.sh

install-node: generate-installers ## Quick install node (requires MASTER_SERVER env var)
	@if [ -z "$(MASTER_SERVER)" ]; then echo "Error: MASTER_SERVER environment variable required"; echo "Usage: make install-node MASTER_SERVER=central.example.com:50052"; exit 1; fi
	@echo "Installing Hysteria2 Node..."
	@MASTER_SERVER=$(MASTER_SERVER) bash scripts/install-node.sh
	@echo "✅ Scaled agents to $(COUNT) instances each"