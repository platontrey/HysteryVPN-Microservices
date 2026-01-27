#!/bin/bash

# Development startup script for Hysteria2 distributed VPN system

set -e

echo "🚀 Starting Hysteria2 Distributed VPN System..."

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Function to check if port is in use
check_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "❌ Port $1 is in use"
        return 1
    else
        echo -e "${GREEN}✓${NC} Port $1 is available"
        return 0
    fi
}

# Function to check if service is healthy
wait_for_service() {
    local service_name=$1
    local health_check=$2
    local max_attempts=30
    local attempt=1

    echo "Waiting for $service_name to be healthy..."
    while [ $attempt -le $max_attempts ]; do
        if eval "$health_check" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ $service_name is healthy!${NC}"
            return 0
        fi
        echo -e "⏳ Attempt $attempt/$max_attempts: $service_name is not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo -e "❌ $service_name failed to become healthy after $max_attempts attempts"
    return 1
}

# Check required ports
echo "Checking required ports..."
check_port 5432 || exit 1  # PostgreSQL
check_port 6379 || exit 1  # Redis
check_port 8080 || exit 1  # API Service
check_port 8081 || exit 1  # Orchestrator Service
check_port 50052 || exit 1 # Orchestrator gRPC
check_port 3000 || exit 1  # Web Service

# Change to deployment directory
cd "$(dirname "$0")/../deployments/docker"

# Start infrastructure services
echo -e "${BLUE}📦 Starting infrastructure services...${NC}"
docker-compose up -d postgres redis

# Wait for database
wait_for_service "PostgreSQL" "docker exec hysteria2-postgres pg_isready -U hysteria2 -d hysteria2_db"

# Wait for Redis
wait_for_service "Redis" "docker exec hysteria2-redis redis-cli ping"

# Start application services
echo -e "${BLUE}🏗️ Starting application services...${NC}"
docker-compose up -d orchestrator-service

# Wait for orchestrator
wait_for_service "Orchestrator" "curl -f http://localhost:8081/health"

# Start API service
echo -e "${BLUE}🔧 Starting API service...${NC}"
cd ../../api-service
go run cmd/server/main.go &
API_PID=$!

# Wait for API
sleep 5
wait_for_service "API Service" "curl -f http://localhost:8080/health"

# Start web service
echo -e "${BLUE}🌐 Starting web service...${NC}"
cd ../web-service
npm run dev &
WEB_PID=$!

# Wait for web service
sleep 5
wait_for_service "Web Service" "curl -f http://localhost:3000"

# Start VPS agents
echo -e "${BLUE}🌍 Starting VPS agents...${NC}"
cd ../deployments/docker
docker-compose up -d agent-us-east agent-europe agent-asia

# Wait for agents to register
sleep 10

echo ""
echo -e "${GREEN}🎉 Hysteria2 Distributed VPN System is running!${NC}"
echo ""
echo -e "${YELLOW}📊 Services:${NC}"
echo "   • Web Interface:    http://localhost:3000"
echo "   • API Service:      http://localhost:8080"
echo "   • Orchestrator:     http://localhost:8081"
echo "   • gRPC (Orchestrator): localhost:50052"
echo ""
echo -e "${YELLOW}🌐 VPS Agents:${NC}"
echo "   • US East Agent:    grpc://localhost:50061"
echo "   • Europe Agent:     grpc://localhost:50062"
echo "   • Asia Agent:       grpc://localhost:50063"
echo ""
echo -e "${YELLOW}💾 Databases:${NC}"
echo "   • PostgreSQL:       localhost:5432"
echo "   • Redis:           localhost:6379"
echo ""
echo -e "${YELLOW}🔍 Monitoring:${NC}"
echo "   View logs:         docker-compose logs -f [service-name]"
echo "   Check status:      docker-compose ps"
echo "   Stop system:       docker-compose down"
echo ""
echo -e "${YELLOW}📖 For development use:${NC}"
echo "   Orchestrator:      cd orchestrator-service && go run cmd/server/main.go"
echo "   Agent (US East):    cd agent-service && NODE_ID=node-us-east-1 go run cmd/agent/main.go"
echo ""
echo -e "${YELLOW}🔧 Process IDs:${NC}"
echo "   API:  $API_PID"
echo "   Web:  $WEB_PID"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for interrupt
trap "echo -e '${BLUE}Stopping services...${NC}'; kill $API_PID $WEB_PID 2>/dev/null; cd deployments/docker && docker-compose down; exit 0" INT

# Keep running
wait