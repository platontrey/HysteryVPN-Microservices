#!/bin/bash

# Docker Utilities Library
# Functions for Docker setup and management

# Check if Docker is installed and running
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_warning "Docker not found. Installing Docker..."
        install_docker
    fi

    if ! systemctl is-active --quiet docker; then
        print_warning "Docker service not running. Starting..."
        systemctl start docker
        systemctl enable docker
    fi

    print_success "Docker is ready"
}

# Install Docker on various distributions
install_docker() {
    if [[ -f /etc/debian_version ]]; then
        # Debian/Ubuntu
        apt-get update
        apt-get install -y \
            apt-transport-https \
            ca-certificates \
            curl \
            gnupg \
            lsb-release

        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

        echo \
          "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
          $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

        apt-get update
        apt-get install -y docker-ce docker-ce-cli containerd.io

    elif [[ -f /etc/redhat-release ]]; then
        # CentOS/RHEL
        yum install -y yum-utils
        yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
        yum install -y docker-ce docker-ce-cli containerd.io
    fi

    # Start and enable docker service
    systemctl start docker
    systemctl enable docker

    # Add current user to docker group if not root
    if [[ $EUID -ne 0 ]]; then
        usermod -aG docker "$(whoami)"
    fi
}

# Install Docker Compose
install_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        print_warning "Installing Docker Compose..."
        curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        chmod +x /usr/local/bin/docker-compose
        print_success "Docker Compose installed"
    else
        print_success "Docker Compose already installed"
    fi
}

# Build all Docker images
build_docker_images() {
    print_step "Building Docker images"

    cd "$SCRIPT_DIR/deployments/docker" || {
        print_error "Docker directory not found"
        exit 1
    }

    docker-compose build --parallel
    print_success "Docker images built"
}

# Start all services
start_services() {
    print_step "Starting all services"

    cd "$SCRIPT_DIR/deployments/docker" || {
        print_error "Docker directory not found"
        exit 1
    }

    # Start infrastructure first
    docker-compose up -d postgres redis

    # Wait for databases
    print_info "Waiting for databases to initialize..."
    sleep 15

    # Start application services
    docker-compose up -d orchestrator-service api-service web-service

    # Start nodes if configured
    if [[ "$NODE_COUNT" -gt 0 ]]; then
        for ((i=1; i<=NODE_COUNT; i++)); do
            docker-compose up -d "agent-node-$i"
        done
    fi

    # Start monitoring if enabled
    if [[ "$MONITORING_ENABLED" == true ]]; then
        docker-compose up -d prometheus grafana node-exporter alertmanager
    fi

    print_success "All services started"
}

# Health checks for all services
run_docker_health_checks() {
    print_step "Running Docker health checks"

    cd "$SCRIPT_DIR/deployments/docker" || return 1

    local services=("postgres" "redis" "orchestrator-service" "api-service" "web-service")

    for service in "${services[@]}"; do
        if docker-compose ps "$service" | grep -q "Up"; then
            print_success "$service is running"
        else
            print_error "$service failed to start"
            return 1
        fi
    done

    # Check API endpoints
    if ! curl -f -s http://localhost:8080/health &>/dev/null; then
        print_error "API service health check failed"
        return 1
    fi

    if ! curl -f -s http://localhost:8081/health &>/dev/null; then
        print_error "Orchestrator service health check failed"
        return 1
    fi

    print_success "All health checks passed"
}

# Stop all services
stop_services() {
    print_step "Stopping all services"

    cd "$SCRIPT_DIR/deployments/docker" || return 1

    docker-compose down
    print_success "Services stopped"
}

# Show Docker service status
show_docker_status() {
    print_header "🐳 DOCKER SERVICE STATUS"

    echo -e "${YELLOW}Running containers:${NC}"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

    echo
    echo -e "${YELLOW}Images:${NC}"
    docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | grep hysteria

    echo
    echo -e "${YELLOW}Networks:${NC}"
    docker network ls | grep hysteria
}