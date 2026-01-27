#!/bin/bash

# Hysteria2 Microservices Setup Script

set -e

echo "🚀 Setting up Hysteria2 Microservices..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi

    print_success "Docker and Docker Compose are installed"
}

# Create necessary directories
create_directories() {
    print_step "Creating necessary directories..."

    mkdir -p hysteria2-microservices/deployments/docker/configs/hysteria2
    mkdir -p hysteria2-microservices/deployments/docker/configs/postgres
    mkdir -p hysteria2-microservices/deployments/docker/logs
    mkdir -p hysteria2-microservices/api-service/migrations
    mkdir -p hysteria2-microservices/web-service/src

    print_success "Directories created"
}

# Generate SSL certificates for Hysteria2
generate_ssl() {
    print_step "Generating SSL certificates for Hysteria2..."

    if [ ! -f "hysteria2-microservices/deployments/docker/configs/hysteria2/server.crt" ]; then
        openssl req -x509 -newkey rsa:4096 -keyout hysteria2-microservices/deployments/docker/configs/hysteria2/server.key -out hysteria2-microservices/deployments/docker/configs/hysteria2/server.crt -days 365 -nodes -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
        print_success "SSL certificates generated"
    else
        print_warning "SSL certificates already exist"
    fi
}

# Create .env file for API service
create_env_file() {
    print_step "Creating environment configuration..."

    cat > hysteria2-microservices/api-service/.env << EOF
# Database Configuration
DATABASE_URL=postgres://hysteria2:password123@localhost:5432/hysteria2_db?sslmode=disable

# Redis Configuration
REDIS_URL=redis://localhost:6379

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production-$(openssl rand -hex 32)
JWT_EXPIRY_HOUR=24

# Server Configuration
PORT=8080
LOG_LEVEL=info
ALLOW_ORIGINS=http://localhost:3000

# Hysteria2 Configuration
HYSTERIA_CONFIG_PATH=/etc/hysteria/server.yaml
HYSTERIA_BINARY_PATH=/usr/local/bin/hysteria
HYSTERIA_HTTP_API_URL=http://localhost:9999
HYSTERIA_ADMIN_TOKEN=hysteria_stats_secret
EOF

    print_success "Environment configuration created"
}

# Build and start services
start_services() {
    print_step "Building and starting services..."

    cd hysteria2-microservices/deployments/docker

    # Start infrastructure first
    print_step "Starting PostgreSQL and Redis..."
    docker-compose up -d postgres redis

    # Wait for services to be ready
    print_step "Waiting for services to be ready..."
    sleep 10

    # Start application services
    print_step "Starting API and Web services..."
    docker-compose up -d api-service web-service

    # Start Hysteria2 server
    print_step "Starting Hysteria2 server..."
    docker-compose up -d hysteria2-server

    cd ../../..

    print_success "All services started successfully"
}

# Show status
show_status() {
    print_step "Checking service status..."

    echo ""
    echo "📊 Service Status:"
    docker-compose -f hysteria2-microservices/deployments/docker/docker-compose.yml ps

    echo ""
    echo "🔗 Service URLs:"
    echo "  API Service:    http://localhost:8080"
    echo "  Web Panel:      http://localhost:3000"
    echo "  Hysteria2:      localhost:4433"
    echo "  API Health:     http://localhost:8080/health"
    echo ""

    echo "👤 Default Admin Credentials:"
    echo "  Username: admin"
    echo "  Password: admin123"
    echo ""

    echo "📝 Useful Commands:"
    echo "  View logs:    make logs"
    echo "  Stop all:     make stop"
    echo "  Restart:      make restart"
}

# Main setup function
main() {
    echo "🎯 Hysteria2 Microservices Setup"
    echo "=================================="

    check_docker
    create_directories
    generate_ssl
    create_env_file
    start_services
    show_status

    print_success "Setup completed successfully!"
    print_warning "Please change the default admin password after first login!"
}

# Run main function
main "$@"