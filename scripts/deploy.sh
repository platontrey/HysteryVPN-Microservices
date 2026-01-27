#!/bin/bash

# Production deployment script for Hysteria2 distributed VPN system

set -e

echo "🚀 Deploying Hysteria2 Distributed VPN System to Production..."

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
ENVIRONMENT=${1:-production}
COMPOSE_FILE="docker-compose.yml"
OVERLAY_FILE="docker-compose.prod.yml"

# Check if docker compose files exist
if [ ! -f "$COMPOSE_FILE" ]; then
    echo -e "${RED}❌ $COMPOSE_FILE not found!${NC}"
    exit 1
fi

if [ ! -f "$OVERLAY_FILE" ]; then
    echo -e "${YELLOW}⚠️ $OVERLAY_FILE not found, using only $COMPOSE_FILE${NC}"
    OVERLAY=""
else
    OVERLAY="-f $OVERLAY_FILE"
fi

# Validate environment
echo -e "${BLUE}🔍 Validating deployment environment...${NC}"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    exit 1
fi

# Check required environment variables
required_vars=("DB_PASSWORD" "JWT_SECRET" "NODE_AUTH_TOKEN")
missing_vars=()

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        missing_vars+=("$var")
    fi
done

if [ ${#missing_vars[@]} -gt 0 ]; then
    echo -e "${RED}❌ Missing required environment variables: ${missing_vars[*]}${NC}"
    echo "Please set these variables in your .env file or environment"
    exit 1
fi

# Pull latest images
echo -e "${BLUE}📦 Pulling latest Docker images...${NC}"
docker-compose -f $COMPOSE_FILE $OVERLAY pull

# Build custom images
echo -e "${BLUE}🔨 Building custom images...${NC}"
docker-compose -f $COMPOSE_FILE $OVERLAY build --no-cache

# Stop existing services
echo -e "${BLUE}🛑 Stopping existing services...${NC}"
docker-compose -f $COMPOSE_FILE $OVERLAY down

# Backup database
echo -e "${BLUE}💾 Creating database backup...${NC}"
BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"
docker exec hysteria2-postgres pg_dump -U hysteria2 hysteria2_db > $BACKUP_FILE
echo -e "${GREEN}✅ Database backed up to $BACKUP_FILE${NC}"

# Run database migrations
echo -e "${BLUE}🔄 Running database migrations...${NC}"
docker-compose -f $COMPOSE_FILE $OVERLAY run --rm orchestrator-service /app/migrate

# Start services
echo -e "${BLUE}🚀 Starting services...${NC}"
docker-compose -f $COMPOSE_FILE $OVERLAY up -d

# Health checks
echo -e "${BLUE}🏥 Performing health checks...${NC}"

services=("postgres:5432" "redis:6379" "orchestrator-service:8081" "api-service:8080" "web-service:80")

for service in "${services[@]}"; do
    service_name=$(echo $service | cut -d: -f1)
    service_port=$(echo $service | cut -d: -f2)
    
    echo "Checking $service_name..."
    max_attempts=30
    attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if docker exec hysteria2-$service_name wget --quiet --tries=1 --spider http://localhost:$service_port/health 2>/dev/null; then
            echo -e "${GREEN}✅ $service_name is healthy${NC}"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            echo -e "${RED}❌ $service_name failed health check${NC}"
            echo -e "${RED}Check logs: docker-compose logs $service_name${NC}"
            exit 1
        fi
        
        echo -e "⏳ $service_name not ready, attempt $attempt/$max_attempts"
        sleep 2
        attempt=$((attempt + 1))
    done
done

# Verify VPS agents
echo -e "${BLUE}🌍 Checking VPS agents...${NC}"
sleep 10

agent_count=$(docker-compose ps | grep -c "agent-")
echo -e "${GREEN}✅ $agent_count VPS agents are running${NC}"

# Show status
echo ""
echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
echo ""
echo -e "${YELLOW}📊 Service Status:${NC}"
docker-compose ps
echo ""
echo -e "${YELLOW}📊 Resource Usage:${NC}"
docker stats --no-stream
echo ""
echo -e "${YELLOW}📊 Recent Logs:${NC}"
docker-compose logs --tail=5
echo ""
echo -e "${YELLOW}🔧 Management Commands:${NC}"
echo "   View logs:         docker-compose logs -f [service-name]"
echo "   Scale services:    docker-compose up -d --scale agent-us-east=3"
echo "   Update service:    docker-compose up -d --no-deps [service-name]"
echo "   Stop system:       docker-compose down"
echo "   Full restart:      docker-compose down && docker-compose up -d"
echo ""
echo -e "${YELLOW}🔍 Monitoring URLs:${NC}"
echo "   Web Interface:     https://your-domain.com"
echo "   API Documentation: https://your-domain.com/docs"
echo "   Health Status:     https://your-domain.com/health"
echo ""
echo -e "${YELLOW}📊 Metrics & Monitoring:${NC}"
echo "   Application logs:  docker-compose logs -f --tail=100"
echo "   System metrics:    docker stats"
echo "   Database status:   docker exec hysteria2-postgres pg_isready"
echo ""