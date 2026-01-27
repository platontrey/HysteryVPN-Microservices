#!/bin/bash

# HysteriaVPN Auto-Installation Script
# This script sets up the complete VPN orchestrator and example nodes with interactive configuration

set -e

echo "🚀 Starting HysteriaVPN auto-installation..."
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    echo "   Visit: https://docs.docker.com/get-docker/"
    echo "   On Windows, install Docker Desktop."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    echo "   Visit: https://docs.docker.com/compose/install/"
    exit 1
fi

echo "✅ Docker and Docker Compose are installed"
echo ""

# Check if we're in the right directory
if [ ! -f "docker-compose.yml" ]; then
    echo "❌ docker-compose.yml not found. Please run this script from the deployments/docker directory"
    exit 1
fi

# Interactive configuration
echo "🔧 Configuration Setup"
echo "======================"

# Get server configuration
read -p "Enter your public domain/IP for the server (default: localhost): " SERVER_HOST
SERVER_HOST=${SERVER_HOST:-localhost}

read -p "Enter database password (default: random): " DB_PASSWORD
if [ -z "$DB_PASSWORD" ]; then
    DB_PASSWORD=$(openssl rand -hex 16)
fi

read -p "Enter JWT secret (default: random): " JWT_SECRET
if [ -z "$JWT_SECRET" ]; then
    JWT_SECRET=$(openssl rand -hex 32)
fi

read -p "Enter node auth token (default: random): " NODE_AUTH_TOKEN
if [ -z "$NODE_AUTH_TOKEN" ]; then
    NODE_AUTH_TOKEN=$(openssl rand -hex 16)
fi

echo ""
echo "Node Configuration:"
read -p "Create example nodes? (y/n, default: y): " CREATE_NODES
CREATE_NODES=${CREATE_NODES:-y}

if [[ $CREATE_NODES =~ ^[Yy]$ ]]; then
    read -p "How many example nodes to create? (1-3, default: 3): " NUM_NODES
    NUM_NODES=${NUM_NODES:-3}
else
    NUM_NODES=0
fi

# Create logs directory
mkdir -p logs

# Create environment files with user input
echo "🔧 Creating environment configuration..."

# Orchestrator .env
cat > orchestrator.env << EOF
DB_HOST=postgres
DB_PORT=5432
DB_USER=hysteria2
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=hysteria2_db
REDIS_HOST=redis
REDIS_PORT=6379
GRPC_HOST=0.0.0.0
GRPC_PORT=50052
SERVER_HOST=0.0.0.0
SERVER_PORT=8081
JWT_SECRET=${JWT_SECRET}
NODE_AUTH_TOKEN=${NODE_AUTH_TOKEN}
LOG_LEVEL=info
LOG_FORMAT=json
EOF

# API .env
cat > api.env << EOF
DATABASE_URL=postgres://hysteria2:${DB_PASSWORD}@postgres:5432/hysteria2_db?sslmode=disable
REDIS_URL=redis://redis:6379
JWT_SECRET=${JWT_SECRET}
LOG_LEVEL=info
ALLOW_ORIGINS=http://${SERVER_HOST}:3000
JWT_EXPIRY_HOUR=24
ORCHESTRATOR_URL=orchestrator-service:50052
EOF

# Web .env
cat > web.env << EOF
REACT_APP_API_URL=http://${SERVER_HOST}:8080
REACT_APP_WS_URL=ws://${SERVER_HOST}:8080
REACT_APP_ORCHESTRATOR_URL=http://${SERVER_HOST}:8081
EOF

echo "🔄 Pulling Docker images..."
docker-compose pull

echo "🏗️ Building services..."
docker-compose build --parallel

echo "🚀 Starting core services..."
docker-compose up -d postgres redis orchestrator-service api-service web-service

echo "⏳ Waiting for core services to be healthy..."
sleep 30

# Start nodes if requested
if [ "$NUM_NODES" -gt 0 ]; then
    echo "🚀 Starting $NUM_NODES example nodes..."
    if [ "$NUM_NODES" -ge 1 ]; then
        docker-compose up -d agent-us-east
    fi
    if [ "$NUM_NODES" -ge 2 ]; then
        docker-compose up -d agent-europe
    fi
    if [ "$NUM_NODES" -ge 3 ]; then
        docker-compose up -d agent-asia
    fi
    sleep 10
fi

# Check if services are running
if docker-compose ps | grep -q "Up"; then
    echo "✅ Installation completed successfully!"
    echo ""
    echo "🌐 Access URLs:"
    echo "   Web Interface: http://${SERVER_HOST}:3000"
    echo "   API: http://${SERVER_HOST}:8081"
    echo "   REST API: http://${SERVER_HOST}:8080"
    echo ""
    if [ "$NUM_NODES" -gt 0 ]; then
        echo "📍 Example Nodes:"
        echo "   US East: hysteria-agent-us-east (port 8081)"
        echo "   Europe: hysteria-agent-europe (port 8082)"
        echo "   Asia: hysteria-agent-asia (port 8083)"
        echo ""
    fi
    echo "📋 Next steps:"
    echo "   1. Access the web interface and create your first user"
    echo "   2. Configure domain names and SSL certificates for production"
    echo "   3. Deploy additional agents on VPS servers"
    echo "   4. Test VPN connections with the HysteryVPN client"
    echo ""
    echo "🔍 Check logs: docker-compose logs -f [service-name]"
    echo "🛑 Stop: docker-compose down"
else
    echo "❌ Some services failed to start. Check logs with: docker-compose logs"
    exit 1
fi