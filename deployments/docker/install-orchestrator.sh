#!/bin/bash

# HysteriaVPN Orchestrator Installation Script
# Installs the central orchestrator server with databases and web interface

set -e

echo "🚀 Installing HysteriaVPN Orchestrator..."
echo ""

# Check and install dependencies
echo "📦 Checking and installing dependencies..."

# Install Docker if not present
if ! command -v docker &> /dev/null; then
    echo "🐳 Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh

    # Start and enable Docker service
    if command -v systemctl &> /dev/null; then
        sudo systemctl start docker
        sudo systemctl enable docker
    fi

    # Add current user to docker group (may require logout/login)
    sudo usermod -aG docker $USER
    echo "⚠️  Added user to docker group. You may need to logout/login or run 'newgrp docker'"
fi

# Install Docker Compose if not present
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "🐳 Installing Docker Compose..."
    if command -v apt-get &> /dev/null; then
        sudo apt-get update
        sudo apt-get install -y docker-compose
    elif command -v yum &> /dev/null; then
        sudo yum install -y docker-compose
    elif command -v dnf &> /dev/null; then
        sudo dnf install -y docker-compose
    else
        echo "❌ Could not install Docker Compose automatically. Please install manually."
        exit 1
    fi
fi

echo "✅ Dependencies OK"
echo ""

# Interactive configuration
echo "🔧 Orchestrator Configuration"
echo "============================="

read -p "Enter server domain/IP (default: localhost): " SERVER_HOST
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

# Create logs directory
mkdir -p logs

# Create .env files
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

cat > api.env << EOF
DATABASE_URL=postgres://hysteria2:${DB_PASSWORD}@postgres:5432/hysteria2_db?sslmode=disable
REDIS_URL=redis://redis:6379
JWT_SECRET=${JWT_SECRET}
LOG_LEVEL=info
ALLOW_ORIGINS=http://${SERVER_HOST}:3000
JWT_EXPIRY_HOUR=24
ORCHESTRATOR_URL=orchestrator-service:50052
EOF

cat > web.env << EOF
REACT_APP_API_URL=http://${SERVER_HOST}:8080
REACT_APP_WS_URL=ws://${SERVER_HOST}:8080
REACT_APP_ORCHESTRATOR_URL=http://${SERVER_HOST}:8081
EOF

echo ""
echo "🚀 Starting orchestrator services..."
docker-compose up -d postgres redis orchestrator-service api-service web-service

echo "⏳ Waiting for services..."
sleep 30

if docker-compose ps postgres redis orchestrator-service api-service web-service | grep -q "Up"; then
    echo ""
    echo "✅ Orchestrator installed successfully!"
    echo "🌐 Web Interface: http://${SERVER_HOST}:3000"
    echo "🔗 API: http://${SERVER_HOST}:8081"
    echo "📊 REST API: http://${SERVER_HOST}:8080"
    echo ""
    echo "📝 Save these credentials:"
    echo "   DB Password: ${DB_PASSWORD}"
    echo "   JWT Secret: ${JWT_SECRET}"
    echo "   Node Auth Token: ${NODE_AUTH_TOKEN}"
    echo ""
    echo "🔍 Logs: docker-compose logs -f"
    echo "🛑 Stop: docker-compose down"
else
    echo "❌ Installation failed. Check logs: docker-compose logs"
    exit 1
fi