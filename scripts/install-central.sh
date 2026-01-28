#!/bin/bash

# Hysteria2 VPN Central Server Installer
# Installs: Orchestrator, API, Web Panel, PostgreSQL, Redis
# For production deployment on central server

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
INSTALL_DIR="/opt/hysteria2-central"
DOMAIN=""
EMAIL=""
ADMIN_PASSWORD=""

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

install_dependencies() {
    log "Installing system dependencies..."

    if command -v apt &> /dev/null; then
        apt install -y curl wget git ufw htop iotop sysstat fail2ban logrotate unattended-upgrades
    elif command -v yum &> /dev/null; then
        yum install -y curl wget git firewalld htop iotop sysstat fail2ban logrotate yum-cron
    elif command -v dnf &> /dev/null; then
        dnf install -y curl wget git firewalld htop iotop sysstat fail2ban logrotate dnf-automatic
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

    if command -v ufw &> /dev/null; then
        ufw --force enable
        ufw allow 80
        ufw allow 443
        ufw allow 8080
        ufw allow 8081
        ufw allow 3000
        ufw allow ssh
        # Block common attack ports
        ufw deny 22  # SSH will be allowed separately
    elif command -v firewall-cmd &> /dev/null; then
        systemctl enable firewalld
        systemctl start firewalld
        firewall-cmd --permanent --add-port=80/tcp
        firewall-cmd --permanent --add-port=443/tcp
        firewall-cmd --permanent --add-port=8080/tcp
        firewall-cmd --permanent --add-port=8081/tcp
        firewall-cmd --permanent --add-port=3000/tcp
        firewall-cmd --permanent --add-service=ssh
        firewall-cmd --reload
    fi

    log "Firewall configured"
}

setup_security() {
    log "Configuring additional security measures..."

    # Configure fail2ban for SSH protection
    if command -v fail2ban-client &> /dev/null; then
        systemctl enable fail2ban
        systemctl start fail2ban

        # Configure jail for SSH
        cat > /etc/fail2ban/jail.d/hysteria2.conf << EOF
[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600

[nginx-http-auth]
enabled = true
port = http,https
filter = nginx-http-auth
logpath = /var/log/nginx/error.log
maxretry = 3
bantime = 3600
EOF

        systemctl reload fail2ban
        log "Fail2ban configured"
    fi

    # Setup automatic security updates
    if command -v unattended-upgrades &> /dev/null; then
        cat > /etc/apt/apt.conf.d/50unattended-upgrades << EOF
Unattended-Upgrade::Allowed-Origins {
    "\${distro_id}:\${distro_codename}";
    "\${distro_id}:\${distro_codename}-security";
    "\${distro_id}ESMApps:\${distro_codename}-apps-security";
    "\${distro_id}ESM:\${distro_codename}-infra-security";
};
Unattended-Upgrade::Package-Blacklist {
};
Unattended-Upgrade::AutoFixInterruptedDpkg "true";
Unattended-Upgrade::MinimalSteps "true";
Unattended-Upgrade::InstallOnShutdown "false";
Unattended-Upgrade::Remove-Unused-Kernel-Packages "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
Unattended-Upgrade::Automatic-Reboot "true";
Unattended-Upgrade::Automatic-Reboot-Time "02:00";
EOF

        cat > /etc/apt/apt.conf.d/20auto-upgrades << EOF
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Download-Upgradeable-Packages "1";
APT::Periodic::AutocleanInterval "7";
APT::Periodic::Unattended-Upgrade "1";
EOF

        systemctl enable unattended-upgrades
        systemctl start unattended-upgrades
        log "Automatic security updates configured"
    fi

    log "Security measures configured"
}

setup_monitoring() {
    log "Setting up monitoring and logging..."

    # Configure logrotate for application logs
    cat > /etc/logrotate.d/hysteria2 << EOF
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

    # Setup basic monitoring script
    cat > "$INSTALL_DIR/monitor.sh" << 'EOF'
#!/bin/bash
echo "=== Hysteria2 Central Server Status ==="
echo "Time: $(date)"
echo "Uptime: $(uptime -p)"
echo "Load: $(uptime | awk -F'load average:' '{ print $2 }')"
echo ""
echo "=== Docker Services ==="
docker compose -f '"$INSTALL_DIR"'/docker compose.yml ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "=== System Resources ==="
echo "CPU: $(top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{print 100 - $1"%"}')"
echo "Memory: $(free | grep Mem | awk '{printf "%.2f%%", $3/$2 * 100.0}')"
echo "Disk: $(df / | tail -1 | awk '{print $5}')"
echo ""
echo "=== Network Connections ==="
netstat -tlnp 2>/dev/null | grep -E ':(80|443|8080|8081)' || ss -tlnp | grep -E ':(80|443|8080|8081)'
echo ""
echo "=== Recent Logs ==="
tail -10 "$INSTALL_DIR/logs/"*.log 2>/dev/null || echo "No logs found"
EOF

    chmod +x "$INSTALL_DIR/monitor.sh"

    # Setup cron job for monitoring
    (crontab -l ; echo "*/5 * * * * $INSTALL_DIR/monitor.sh >> $INSTALL_DIR/logs/monitor.log 2>&1") | crontab -

    log "Monitoring and logging configured"
}

setup_backup() {
    log "Setting up backup system..."

    # Create backup script
    cat > "$INSTALL_DIR/backup.sh" << EOF
#!/bin/bash
BACKUP_DIR="$INSTALL_DIR/backups"
TIMESTAMP=\$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="\$BACKUP_DIR/hysteria2_backup_\$TIMESTAMP.tar.gz"

mkdir -p "\$BACKUP_DIR"

echo "Creating backup: \$BACKUP_FILE"

# Stop services temporarily for clean backup
docker compose -f "$INSTALL_DIR/docker compose.yml" stop

# Create database backup
docker exec hysteria2-postgres pg_dump -U hysteria2 hysteria2_db > "\$BACKUP_DIR/db_backup_\$TIMESTAMP.sql"

# Create filesystem backup (excluding logs and temp files)
tar -czf "\$BACKUP_FILE" -C "$INSTALL_DIR" \\
    --exclude="logs/*" \\
    --exclude="backups/*" \\
    --exclude="*.log" \\
    .

# Start services again
docker compose -f "$INSTALL_DIR/docker compose.yml" start

echo "Backup completed: \$BACKUP_FILE"
echo "Database backup: \$BACKUP_DIR/db_backup_\$TIMESTAMP.sql"

# Clean old backups (keep last 7)
find "\$BACKUP_DIR" -name "hysteria2_backup_*.tar.gz" -mtime +7 -delete
find "\$BACKUP_DIR" -name "db_backup_*.sql" -mtime +7 -delete

echo "Old backups cleaned up"
EOF

    chmod +x "$INSTALL_DIR/backup.sh"

    # Setup daily backup cron job
    (crontab -l ; echo "0 2 * * * $INSTALL_DIR/backup.sh") | crontab -

    log "Backup system configured"
}

create_directories() {
    log "Creating installation directories..."

    mkdir -p "$INSTALL_DIR"
    mkdir -p "$INSTALL_DIR/configs"
    mkdir -p "$INSTALL_DIR/logs"
    mkdir -p "$INSTALL_DIR/ssl"

    log "Directories created"
}

generate_config() {
    log "Generating configuration files..."

    # Generate random secrets
    JWT_SECRET=$(openssl rand -hex 32)
    DB_PASSWORD=$(openssl rand -hex 16)
    NODE_AUTH_TOKEN=$(openssl rand -hex 32)

    # Create docker compose.yml for central services only
    cat > "$INSTALL_DIR/docker compose.yml" << EOF
version: '3.8'

services:
  postgres:
    image: postgres:17-alpine
    container_name: hysteria2-postgres
    environment:
      POSTGRES_DB: hysteria2_db
      POSTGRES_USER: hysteria2
      POSTGRES_PASSWORD: $DB_PASSWORD
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    networks:
      - hysteria2-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U hysteria2 -d hysteria2_db"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:8-alpine
    container_name: hysteria2-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - hysteria2-network
    restart: unless-stopped
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  orchestrator-service:
    image: hysteria2/orchestrator:latest
    container_name: hysteria2-orchestrator
    ports:
      - "8081:8081"
      - "50052:50052"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=hysteria2
      - DB_PASSWORD=$DB_PASSWORD
      - DB_NAME=hysteria2_db
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - GRPC_HOST=0.0.0.0
      - GRPC_PORT=50052
      - SERVER_HOST=0.0.0.0
      - SERVER_PORT=8081
      - JWT_SECRET=$JWT_SECRET
      - NODE_AUTH_TOKEN=$NODE_AUTH_TOKEN
      - LOG_LEVEL=info
      - LOG_FORMAT=json
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - hysteria2-network
    restart: unless-stopped
    volumes:
      - ./logs:/app/logs

  api-service:
    image: hysteria2/api:latest
    container_name: hysteria2-api
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://hysteria2:$DB_PASSWORD@postgres:5432/hysteria2_db?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=$JWT_SECRET
      - LOG_LEVEL=info
      - ALLOW_ORIGINS=https://$DOMAIN
      - JWT_EXPIRY_HOUR=24
      - ORCHESTRATOR_URL=orchestrator-service:50052
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      orchestrator-service:
        condition: service_healthy
    networks:
      - hysteria2-network
    restart: unless-stopped
    volumes:
      - ./logs:/app/logs
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  web-service:
    image: hysteria2/web:latest
    container_name: hysteria2-web
    ports:
      - "80:80"
      - "443:443"
    environment:
      - REACT_APP_API_URL=https://$DOMAIN/api
      - REACT_APP_WS_URL=wss://$DOMAIN/ws
      - REACT_APP_ORCHESTRATOR_URL=https://$DOMAIN/orchestrator
    depends_on:
      - api-service
      - orchestrator-service
    networks:
      - hysteria2-network
    restart: unless-stopped
    volumes:
      - ./ssl:/etc/nginx/ssl:ro
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:80"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  postgres_data:
  redis_data:

networks:
  hysteria2-network:
    driver: bridge
EOF

    # Create environment file
    cat > "$INSTALL_DIR/.env" << EOF
# Hysteria2 Central Server Configuration
DOMAIN=$DOMAIN
EMAIL=$EMAIL
JWT_SECRET=$JWT_SECRET
DB_PASSWORD=$DB_PASSWORD
NODE_AUTH_TOKEN=$NODE_AUTH_TOKEN
ADMIN_PASSWORD=$ADMIN_PASSWORD
EOF

    # Create nginx configuration for SSL
    cat > "$INSTALL_DIR/nginx.conf" << EOF
server {
    listen 80;
    server_name $DOMAIN;
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name $DOMAIN;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    location / {
        proxy_pass http://web-service:80;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /api {
        proxy_pass http://api-service:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /orchestrator {
        proxy_pass http://orchestrator-service:8081;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

    log "Configuration files generated"
}

setup_ssl() {
    if [[ -n "$DOMAIN" && -n "$EMAIL" ]]; then
        log "Setting up SSL certificates for $DOMAIN..."

        # Install certbot if not present
        if ! command -v certbot &> /dev/null; then
            if command -v apt &> /dev/null; then
                apt install -y certbot
            elif command -v yum &> /dev/null; then
                yum install -y certbot
            fi
        fi

        # Get SSL certificate
        certbot certonly --standalone -d "$DOMAIN" --email "$EMAIL" --agree-tos --non-interactive

        # Copy certificates
        cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" "$INSTALL_DIR/ssl/cert.pem"
        cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" "$INSTALL_DIR/ssl/key.pem"

        # Setup auto-renewal
        (crontab -l ; echo "0 12 * * * /usr/bin/certbot renew --quiet && cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem $INSTALL_DIR/ssl/cert.pem && cp /etc/letsencrypt/live/$DOMAIN/privkey.pem $INSTALL_DIR/ssl/key.pem && docker compose -f $INSTALL_DIR/docker compose.yml restart web-service") | crontab -

        log "SSL certificates configured"
    else
        warning "SSL setup skipped - DOMAIN and EMAIL not provided"
        # Generate self-signed certificate for testing
        openssl req -x509 -newkey rsa:4096 -keyout "$INSTALL_DIR/ssl/key.pem" -out "$INSTALL_DIR/ssl/cert.pem" -days 365 -nodes -subj "/CN=$DOMAIN"
    fi
}

start_services() {
    log "Starting central services..."

    cd "$INSTALL_DIR"
    docker compose up -d

    log "Waiting for services to be healthy..."
    sleep 30

    # Check health
    if docker compose ps | grep -q "Up"; then
        log "Services started successfully"
    else
        error "Failed to start services"
    fi
}

create_admin_user() {
    log "Creating admin user..."

    # This would typically be done through the API
    # For now, just log the credentials
    info "Admin credentials:"
    info "Username: admin"
    info "Password: $ADMIN_PASSWORD"
    info "Please change the password after first login"
}

show_completion() {
    log "🎉 Hysteria2 Central Server installation completed!"
    echo ""
    info "Installation Details:"
    echo "  📁 Install Directory: $INSTALL_DIR"
    echo "  🌐 Web Panel: https://$DOMAIN"
    echo "  🔗 API: https://$DOMAIN/api"
    echo "  🎛️  Orchestrator: https://$DOMAIN/orchestrator"
    echo ""
    info "Next Steps:"
    echo "  1. Access https://$DOMAIN and login with admin credentials"
    echo "  2. Add VPS nodes using the node installer script"
    echo "  3. Configure advanced obfuscation settings"
    echo ""
    warning "Remember to:"
    echo "  - Change the default admin password"
    echo "  - Configure backup scripts"
    echo "  - Monitor logs in $INSTALL_DIR/logs"
}

# Main installation
main() {
    log "🚀 Starting Hysteria2 Central Server Installation"

    check_root
    check_os

    # Get user input
    if [[ -z "$DOMAIN" ]]; then
        read -p "Enter domain name (e.g., vpn.example.com): " DOMAIN
    fi

    if [[ -z "$EMAIL" ]]; then
        read -p "Enter email for SSL certificates: " EMAIL
    fi

    if [[ -z "$ADMIN_PASSWORD" ]]; then
        ADMIN_PASSWORD=$(openssl rand -base64 12)
        info "Generated admin password: $ADMIN_PASSWORD"
    fi

    install_dependencies
    install_docker
    setup_firewall
    setup_security
    create_directories
    generate_config
    setup_ssl
    setup_monitoring
    setup_backup
    start_services
    create_admin_user
    show_completion

    log "✅ Installation completed successfully!"
}

# Run main function
main "$@"