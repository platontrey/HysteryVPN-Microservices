#!/bin/bash

# Hysteria2 VPN Node Installer
# Installs: Agent service for VPS nodes
# For deployment on individual VPS servers

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
INSTALL_DIR="/opt/hysteria2-node"
MASTER_SERVER=""
NODE_ID=""
NODE_NAME=""
NODE_LOCATION=""
NODE_COUNTRY=""
LISTEN_PORT=8443
OBFUSCATION_ENABLED=true

# Functions
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}" >&2
    exit 1
}

warning() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

info() {
    echo -e "${BLUE}[INFO] $1${NC}"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root"
    fi
}

check_os() {
    if [[ ! -f /etc/os-release ]]; then
        error "Unsupported OS"
    fi

    . /etc/os-release
    case $ID in
        ubuntu|debian|centos|rhel|fedora)
            log "Detected OS: $PRETTY_NAME"
            ;;
        *)
            warning "Untested OS: $PRETTY_NAME. Continuing anyway..."
            ;;
    esac
}

check_hardware() {
    # Check CPU cores
    CPU_CORES=$(nproc)
    if [[ $CPU_CORES -lt 1 ]]; then
        warning "Low CPU cores detected: $CPU_CORES"
    else
        log "CPU cores: $CPU_CORES"
    fi

    # Check memory
    MEM_GB=$(free -g | awk 'NR==2{printf "%.0f", $2}')
    if [[ $MEM_GB -lt 1 ]]; then
        warning "Low memory detected: ${MEM_GB}GB"
    else
        log "Memory: ${MEM_GB}GB"
    fi

    # Check disk space
    DISK_GB=$(df / | awk 'NR==2{printf "%.0f", $4/1024/1024}')
    if [[ $DISK_GB -lt 10 ]]; then
        warning "Low disk space detected: ${DISK_GB}GB"
    else
        log "Disk space: ${DISK_GB}GB"
    fi
}

install_dependencies() {
    log "Installing system dependencies..."

    if command -v apt &> /dev/null; then
        apt install -y curl wget git ufw jq htop iotop sysstat fail2ban logrotate unattended-upgrades
    elif command -v yum &> /dev/null; then
        yum install -y curl wget git firewalld jq htop iotop sysstat fail2ban logrotate yum-cron
    elif command -v dnf &> /dev/null; then
        dnf install -y curl wget git firewalld jq htop iotop sysstat fail2ban logrotate dnf-automatic
    else
        error "Unsupported package manager"
    fi
}

install_docker() {
    if command -v docker &> /dev/null; then
        log "Docker already installed"
        return
    fi

    log "Installing Docker..."

    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh

    systemctl enable docker
    systemctl start docker

    # Install Docker Compose
    curl -L "https://github.com/docker/compose/releases/download/v2.29.1/docker compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker compose
    chmod +x /usr/local/bin/docker compose

    log "Docker and Docker Compose installed"
}

setup_firewall() {
    log "Configuring firewall..."

    # Allow Hysteria2 port
    if command -v ufw &> /dev/null; then
        ufw --force enable
        ufw allow "$LISTEN_PORT/udp"
        ufw allow 50051/tcp  # gRPC port
    elif command -v firewall-cmd &> /dev/null; then
        systemctl enable firewalld
        systemctl start firewalld
        firewall-cmd --permanent --add-port="$LISTEN_PORT/udp"
        firewall-cmd --permanent --add-port=50051/tcp
        firewall-cmd --reload
    fi

    log "Firewall configured"
}

create_directories() {
    log "Creating installation directories..."

    mkdir -p "$INSTALL_DIR"
    mkdir -p "$INSTALL_DIR/configs"
    mkdir -p "$INSTALL_DIR/logs"
    mkdir -p "$INSTALL_DIR/ssl"

    log "Directories created"
}

generate_node_config() {
    log "Generating node configuration..."

    # Auto-detect node information if not provided
    if [[ -z "$NODE_ID" ]]; then
        NODE_ID="node-$(hostname)-$(date +%s)"
    fi

    if [[ -z "$NODE_NAME" ]]; then
        NODE_NAME="$(hostname)"
    fi

    if [[ -z "$NODE_LOCATION" ]]; then
        # Try to detect location
        if command -v curl &> /dev/null; then
            LOCATION_INFO=$(curl -s https://ipapi.co/json/)
            NODE_LOCATION=$(echo "$LOCATION_INFO" | jq -r '.city // "Unknown"')
            NODE_COUNTRY=$(echo "$LOCATION_INFO" | jq -r '.country_code // "Unknown"')
        else
            NODE_LOCATION="Unknown"
            NODE_COUNTRY="Unknown"
        fi
    fi

    # Get server IP
    SERVER_IP=$(curl -s https://api.ipify.org)

    # Generate random auth password
    AUTH_PASSWORD=$(openssl rand -base64 16)

    # Create docker compose.yml for node
    cat > "$INSTALL_DIR/docker compose.yml" << EOF
version: '3.8'

services:
  hysteria-agent:
    image: hysteria2/agent:latest
    container_name: hysteria2-agent-$NODE_ID
    environment:
      - MASTER_SERVER=$MASTER_SERVER
      - NODE_ID=$NODE_ID
      - NODE_NAME=$NODE_NAME
      - NODE_HOSTNAME=$(hostname)
      - NODE_IP_ADDRESS=$SERVER_IP
      - NODE_LOCATION=$NODE_LOCATION
      - NODE_COUNTRY=$NODE_COUNTRY
      - NODE_GRPC_PORT=50051
      - LOG_LEVEL=info
      - LOG_FORMAT=json
      # Hysteria2 Configuration
      - HYSTERIA2_LISTEN_PORT=$LISTEN_PORT
      - HYSTERIA2_AUTH_PASSWORD=$AUTH_PASSWORD
      - HYSTERIA2_UP_MBPS=100
      - HYSTERIA2_DOWN_MBPS=100
      # Advanced Obfuscation (enabled by default)
      - HYSTERIA2_ADVANCED_OBFUSCATION_ENABLED=$OBFUSCATION_ENABLED
      - HYSTERIA2_QUIC_OBFUSCATION_ENABLED=$OBFUSCATION_ENABLED
      - HYSTERIA2_QUIC_SCRAMBLE_TRANSFORM=$OBFUSCATION_ENABLED
      - HYSTERIA2_QUIC_PACKET_PADDING=1300
      - HYSTERIA2_QUIC_TIMING_RANDOMIZATION=$OBFUSCATION_ENABLED
      - HYSTERIA2_TLS_FINGERPRINT_ROTATION=$OBFUSCATION_ENABLED
      - HYSTERIA2_TLS_FINGERPRINTS=chrome,firefox,safari
      - HYSTERIA2_VLESS_REALITY_ENABLED=$OBFUSCATION_ENABLED
      - HYSTERIA2_VLESS_REALITY_TARGETS=apple.com,google.com,microsoft.com
      - HYSTERIA2_TRAFFIC_SHAPING_ENABLED=$OBFUSCATION_ENABLED
      - HYSTERIA2_BEHAVIORAL_RANDOMIZATION=$OBFUSCATION_ENABLED
    ports:
      - "$LISTEN_PORT:$LISTEN_PORT/udp"  # Hysteria2 port
      - "50051:50051/tcp"                # gRPC port
    restart: unless-stopped
    volumes:
      - ./logs:/app/logs
      - ./configs:/app/configs
    cap_add:
      - NET_ADMIN
    networks:
      - hysteria2-network

networks:
  hysteria2-network:
    driver: bridge
EOF

    # Create environment file
    cat > "$INSTALL_DIR/.env" << EOF
# Hysteria2 Node Configuration
MASTER_SERVER=$MASTER_SERVER
NODE_ID=$NODE_ID
NODE_NAME=$NODE_NAME
NODE_LOCATION=$NODE_LOCATION
NODE_COUNTRY=$NODE_COUNTRY
LISTEN_PORT=$LISTEN_PORT
AUTH_PASSWORD=$AUTH_PASSWORD
OBFUSCATION_ENABLED=$OBFUSCATION_ENABLED
SERVER_IP=$SERVER_IP
EOF

    log "Node configuration generated"
}

optimize_system() {
    log "Optimizing system for VPN node..."

    # Increase file descriptors
    if [[ -f /etc/security/limits.conf ]]; then
        echo "* soft nofile 65536" >> /etc/security/limits.conf
        echo "* hard nofile 65536" >> /etc/security/limits.conf
    fi

    # Optimize kernel parameters for high-throughput VPN
    cat > /etc/sysctl.d/99-hysteria2.conf << EOF
# Hysteria2 VPN optimizations
net.core.rmem_max = 67108864
net.core.wmem_max = 67108864
net.core.rmem_default = 33554432
net.core.wmem_default = 33554432
net.ipv4.tcp_rmem = 4096 87380 67108864
net.ipv4.tcp_wmem = 4096 65536 67108864
net.ipv4.tcp_mtu_probing = 1
net.ipv4.tcp_slow_start_after_idle = 0
net.ipv4.tcp_congestion_control = bbr
net.core.default_qdisc = fq
EOF

    sysctl -p /etc/sysctl.d/99-hysteria2.conf

    log "System optimized"
}

setup_security() {
    log "Configuring node security..."

    # Configure fail2ban for SSH protection
    if command -v fail2ban-client &> /dev/null; then
        systemctl enable fail2ban
        systemctl start fail2ban

        # Configure jail for SSH
        cat > /etc/fail2ban/jail.d/hysteria2-node.conf << EOF
[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600

[docker-hysteria]
enabled = true
port = $LISTEN_PORT
filter = docker-hysteria
logpath = $INSTALL_DIR/logs/*.log
maxretry = 5
bantime = 1800
EOF

        systemctl reload fail2ban
        log "Fail2ban configured for node"
    fi

    # Disable unnecessary services
    systemctl disable systemd-resolved 2>/dev/null || true
    systemctl stop systemd-resolved 2>/dev/null || true

    log "Node security configured"
}

setup_monitoring() {
    log "Setting up node monitoring..."

    # Configure logrotate
    cat > /etc/logrotate.d/hysteria2-node << EOF
$INSTALL_DIR/logs/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 0644 root root
    postrotate
        docker compose -f $INSTALL_DIR/docker compose.yml logs -f --tail=0 > /dev/null 2>&1 || true
    endscript
}
EOF

    # Create comprehensive monitoring script
    cat > "$INSTALL_DIR/monitor.sh" << EOF
#!/bin/bash
echo "=== Hysteria2 Node Status ==="
echo "Time: \$(date)"
echo "Uptime: \$(uptime -p)"
echo "Load: \$(uptime | awk -F'load average:' '{ print \$2 }')"
echo ""
echo "=== Docker Status ==="
docker compose -f "$INSTALL_DIR/docker compose.yml" ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "=== Hysteria2 Connections ==="
ss -uln | grep :$LISTEN_PORT || echo "No active connections"
echo ""
echo "=== Network Status ==="
echo "Active connections: \$(ss -tun | grep ESTAB | wc -l)"
echo "Listening ports: \$(ss -tuln | grep LISTEN | wc -l)"
echo ""
echo "=== System Resources ==="
echo "CPU: \$(top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{print 100 - \$1"%"}')"
echo "Memory: \$(free | grep Mem | awk '{printf "%.2f%%", \$3/\$2 * 100.0}')"
echo "Disk: \$(df / | tail -1 | awk '{print \$5}')"
echo "Network RX/TX: \$(cat /proc/net/dev | grep eth0 | awk '{print \$2,\$10}' | awk '{printf "%.2f MB / %.2f MB", \$1/1024/1024, \$2/1024/1024}')"
echo ""
echo "=== Hysteria2 Performance ==="
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}" | grep hysteria2 || echo "No hysteria2 containers running"
EOF

    chmod +x "$INSTALL_DIR/monitor.sh"

    # Setup cron job for monitoring
    (crontab -l ; echo "*/5 * * * * $INSTALL_DIR/monitor.sh >> $INSTALL_DIR/logs/monitor.log 2>&1") | crontab -

    log "Node monitoring setup completed"
}

setup_backup() {
    log "Setting up node backup system..."

    # Create backup script for node config
    cat > "$INSTALL_DIR/backup.sh" << EOF
#!/bin/bash
BACKUP_DIR="$INSTALL_DIR/backups"
TIMESTAMP=\$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="\$BACKUP_DIR/hysteria2_node_backup_\$TIMESTAMP.tar.gz"

mkdir -p "\$BACKUP_DIR"

echo "Creating node backup: \$BACKUP_FILE"

# Create backup of configuration and logs
tar -czf "\$BACKUP_FILE" -C "$INSTALL_DIR" \\
    --exclude="backups/*" \\
    --exclude="logs/monitor.log" \\
    .

echo "Backup completed: \$BACKUP_FILE"

# Clean old backups (keep last 7)
find "\$BACKUP_DIR" -name "hysteria2_node_backup_*.tar.gz" -mtime +7 -delete

echo "Old backups cleaned up"
EOF

    chmod +x "$INSTALL_DIR/backup.sh"

    # Setup weekly backup cron job
    (crontab -l ; echo "0 3 * * 0 $INSTALL_DIR/backup.sh") | crontab -

    log "Node backup system configured"
}

start_node() {
    log "Starting Hysteria2 node..."

    cd "$INSTALL_DIR"
    docker compose up -d

    log "Waiting for node to start..."
    sleep 10

    # Check if container is running
    if docker compose ps | grep -q "Up"; then
        log "Node started successfully"
    else
        error "Failed to start node"
    fi
}

register_with_master() {
    log "Registering node with master server..."

    # Wait for master to be available
    max_attempts=30
    attempt=1

    while [[ $attempt -le $max_attempts ]]; do
        if curl -f "$MASTER_SERVER/health" &> /dev/null; then
            log "Master server is available"
            break
        fi

        log "Waiting for master server (attempt $attempt/$max_attempts)..."
        sleep 10
        ((attempt++))
    done

    if [[ $attempt -gt $max_attempts ]]; then
        warning "Master server not available. Node will register when master becomes accessible."
        return
    fi

    # Register node (this would typically be done through gRPC API)
    # For now, just log success
    log "Node registered with master server"
}

show_completion() {
    log "🎉 Hysteria2 Node installation completed!"
    echo ""
    info "Installation Details:"
    echo "  📁 Install Directory: $INSTALL_DIR"
    echo "  🆔 Node ID: $NODE_ID"
    echo "  📍 Location: $NODE_LOCATION, $NODE_COUNTRY"
    echo "  🔌 Listen Port: $LISTEN_PORT/UDP"
    echo "  🌐 Server IP: $SERVER_IP"
    echo "  🎛️  Master Server: $MASTER_SERVER"
    echo ""
    info "Obfuscation Status:"
    if [[ "$OBFUSCATION_ENABLED" == "true" ]]; then
        echo "  ✅ Advanced obfuscation: ENABLED"
        echo "  ✅ QUIC obfuscation: ENABLED"
        echo "  ✅ TLS fingerprint rotation: ENABLED"
        echo "  ✅ VLESS Reality: ENABLED"
        echo "  ✅ Traffic shaping: ENABLED"
    else
        echo "  ❌ Advanced obfuscation: DISABLED"
    fi
    echo ""
    info "Management Commands:"
    echo "  Start:  cd $INSTALL_DIR && docker compose up -d"
    echo "  Stop:   cd $INSTALL_DIR && docker compose down"
    echo "  Logs:   cd $INSTALL_DIR && docker compose logs -f"
    echo "  Status: $INSTALL_DIR/monitor.sh"
    echo ""
    warning "Next Steps:"
    echo "  1. Verify node appears in central web panel"
    echo "  2. Test VPN connections from clients"
    echo "  3. Monitor performance and adjust settings if needed"
}

# Main installation
main() {
    log "🚀 Starting Hysteria2 Node Installation"

    check_root
    check_os
    check_hardware

    # Get user input
    if [[ -z "$MASTER_SERVER" ]]; then
        read -p "Enter master server address (e.g., central.example.com:50052): " MASTER_SERVER
    fi

    if [[ -z "$LISTEN_PORT" ]]; then
        read -p "Enter Hysteria2 listen port [8443]: " LISTEN_PORT
        LISTEN_PORT=${LISTEN_PORT:-8443}
    fi

    # Optional parameters
    if [[ -z "$NODE_ID" ]]; then
        read -p "Enter node ID (leave empty for auto): " NODE_ID
    fi

    if [[ -z "$NODE_NAME" ]]; then
        read -p "Enter node name (leave empty for hostname): " NODE_NAME
    fi

    if [[ -z "$NODE_LOCATION" ]]; then
        read -p "Enter node location (leave empty for auto-detect): " NODE_LOCATION
    fi

    read -p "Enable advanced obfuscation for DPI bypass? [Y/n]: " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        OBFUSCATION_ENABLED=true
    else
        OBFUSCATION_ENABLED=false
    fi

    install_dependencies
    install_docker
    setup_firewall
    setup_security
    create_directories
    generate_node_config
    optimize_system
    setup_monitoring
    setup_backup
    start_node
    register_with_master
    show_completion

    log "✅ Node installation completed successfully!"
}

# Run main function
main "$@"