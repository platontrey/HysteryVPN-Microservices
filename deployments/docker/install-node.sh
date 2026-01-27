#!/bin/bash

# HysteriaVPN Node Installation Script
# Installs a VPN node agent on a VPS server

set -e

echo "🚀 Installing HysteriaVPN Node Agent..."
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

    # Add current user to docker group
    sudo usermod -aG docker $USER
    echo "⚠️  Added user to docker group. You may need to logout/login or run 'newgrp docker'"
fi

echo "✅ Dependencies OK"
echo ""

# Interactive configuration
echo "🔧 Node Configuration"
echo "===================="

read -p "Enter orchestrator server IP/domain:port (e.g., 192.168.1.100:50052): " MASTER_SERVER
if [ -z "$MASTER_SERVER" ]; then
    echo "❌ Master server is required"
    exit 1
fi

read -p "Enter node ID (unique identifier, e.g., node-us-east-1): " NODE_ID
if [ -z "$NODE_ID" ]; then
    echo "❌ Node ID is required"
    exit 1
fi

read -p "Enter node name (e.g., US East): " NODE_NAME
NODE_NAME=${NODE_NAME:-${NODE_ID}}

read -p "Enter node location (e.g., New York): " NODE_LOCATION
NODE_LOCATION=${NODE_LOCATION:-Unknown}

read -p "Enter node country code (e.g., US): " NODE_COUNTRY
NODE_COUNTRY=${NODE_COUNTRY:-XX}

read -p "Enter node public IP address: " NODE_IP_ADDRESS
if [ -z "$NODE_IP_ADDRESS" ]; then
    # Try to detect public IP
    NODE_IP_ADDRESS=$(curl -s ifconfig.me 2>/dev/null || echo "127.0.0.1")
    echo "Auto-detected IP: ${NODE_IP_ADDRESS}"
fi

read -p "Enter Hysteria2 port (default: 8080): " HYSTERIA_PORT
HYSTERIA_PORT=${HYSTERIA_PORT:-8080}

echo ""
echo "📋 Configuration Summary:"
echo "   Master Server: ${MASTER_SERVER}"
echo "   Node ID: ${NODE_ID}"
echo "   Node Name: ${NODE_NAME}"
echo "   Location: ${NODE_LOCATION}"
echo "   Country: ${NODE_COUNTRY}"
echo "   IP Address: ${NODE_IP_ADDRESS}"
echo "   Hysteria Port: ${HYSTERIA_PORT}"
echo ""

read -p "Continue with installation? (y/n): " CONFIRM
if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
    echo "Installation cancelled."
    exit 0
fi

# Check if agent-service directory exists
if [ ! -d "../agent-service" ]; then
    echo "❌ agent-service directory not found. Run this script from deployments/docker/"
    exit 1
fi

# Build the agent image
echo "🏗️ Building agent image..."
docker build -t hysteria-agent ../agent-service

# Create container name
CONTAINER_NAME="hysteria-agent-${NODE_ID}"

# Remove existing container if exists
if docker ps -a --format 'table {{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "🛑 Removing existing container..."
    docker rm -f ${CONTAINER_NAME}
fi

# Run the agent
echo "🚀 Starting agent..."
docker run -d \
    --name ${CONTAINER_NAME} \
    --restart unless-stopped \
    -p ${HYSTERIA_PORT}:${HYSTERIA_PORT}/udp \
    -p 50051:50051 \
    -e MASTER_SERVER=${MASTER_SERVER} \
    -e NODE_ID=${NODE_ID} \
    -e NODE_NAME="${NODE_NAME}" \
    -e NODE_HOSTNAME=${NODE_ID}.vpn.local \
    -e NODE_IP_ADDRESS=${NODE_IP_ADDRESS} \
    -e NODE_LOCATION="${NODE_LOCATION}" \
    -e NODE_COUNTRY=${NODE_COUNTRY} \
    -e NODE_GRPC_PORT=50051 \
    -e LOG_LEVEL=info \
    -e LOG_FORMAT=json \
    hysteria-agent

# Wait a bit
sleep 5

# Check if running
if docker ps --format 'table {{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo ""
    echo "✅ Node agent installed successfully!"
    echo "📍 Container: ${CONTAINER_NAME}"
    echo "🌐 Hysteria2 Port: ${HYSTERIA_PORT}"
    echo "🔗 gRPC Port: 50051"
    echo ""
    echo "🔍 Logs: docker logs -f ${CONTAINER_NAME}"
    echo "🛑 Stop: docker stop ${CONTAINER_NAME} && docker rm ${CONTAINER_NAME}"
    echo ""
    echo "The node should automatically register with the orchestrator."
else
    echo "❌ Agent failed to start. Check logs: docker logs ${CONTAINER_NAME}"
    exit 1
fi