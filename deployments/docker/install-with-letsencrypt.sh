#!/bin/bash

# Hysteria2 Node Installation Script with Let's Encrypt Support
# This script automatically sets up Hysteria2 with SSL certificates

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration defaults
NODE_ID=""
DOMAINS=""
EMAIL=""
INSTALL_DIR="/opt/hysteria2"
CONFIG_DIR="/etc/hysteria2"
SNI_ENABLED=false
LETSENCRYPT_ENABLED=false
AUTO_RENEW=true

# Helper functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
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

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --node-id)
                NODE_ID="$2"
                shift 2
                ;;
            --domains)
                DOMAINS="$2"
                shift 2
                ;;
            --email)
                EMAIL="$2"
                shift 2
                ;;
            --sni)
                SNI_ENABLED=true
                shift
                ;;
            --letsencrypt)
                LETSENCRYPT_ENABLED=true
                shift
                ;;
            --no-auto-renew)
                AUTO_RENEW=false
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

show_help() {
    cat << EOF
Hysteria2 Node Installation Script with Let's Encrypt Support

Usage: $0 [OPTIONS]

Required Options:
    --node-id TEXT        Node identifier
    --domains TEXT        Comma-separated list of domains (e.g., vpn1.example.com,vpn2.example.com)
    --email TEXT          Email address for Let's Encrypt (required if using Let's Encrypt)

Optional Options:
    --sni                 Enable SNI support (default: false)
    --letsencrypt          Use Let's Encrypt certificates (default: false)
    --no-auto-renew       Disable automatic certificate renewal (default: enabled)

Examples:
    # Basic installation with self-signed certificates
    $0 --node-id node-001 --domains vpn.example.com --sni

    # Installation with Let's Encrypt certificates
    $0 --node-id node-001 --domains vpn1.example.com,vpn2.example.com --email admin@example.com --sni --letsencrypt

    # Installation with auto-renewal disabled
    $0 --node-id node-001 --domains vpn.example.com --email admin@example.com --sni --letsencrypt --no-auto-renew

EOF
}

# Validate inputs
validate_inputs() {
    if [[ -z "$NODE_ID" ]]; then
        print_error "Node ID is required (--node-id)"
        exit 1
    fi

    if [[ -z "$DOMAINS" ]]; then
        print_error "Domains are required (--domains)"
        exit 1
    fi

    if [[ "$LETSENCRYPT_ENABLED" == true && -z "$EMAIL" ]]; then
        print_error "Email is required when using Let's Encrypt (--email)"
        exit 1
    fi

    if [[ "$SNI_ENABLED" != true && -n "$DOMAINS" && "$LETSENCRYPT_ENABLED" == true ]]; then
        print_warning "Let's Encrypt requires SNI to be enabled. Enabling SNI automatically."
        SNI_ENABLED=true
    fi
}

# Check system requirements
check_requirements() {
    print_info "Checking system requirements..."

    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root"
        exit 1
    fi

    # Check if ports 80 and 443 are available
    if ! netstat -tuln | grep -q ":80 "; then
        print_warning "Port 80 may be required for Let's Encrypt domain validation"
    fi

    if ! netstat -tuln | grep -q ":443 "; then
        print_info "Port 443 is available for Hysteria2"
    else
        print_warning "Port 443 is already in use"
    fi

    # Check OS
    if [[ -f /etc/debian_version ]]; then
        print_info "Detected Debian-based system"
        OS_TYPE="debian"
    elif [[ -f /etc/redhat-release ]]; then
        print_info "Detected RedHat-based system"
        OS_TYPE="redhat"
    else
        print_error "Unsupported operating system"
        exit 1
    fi
}

# Install system dependencies
install_dependencies() {
    print_info "Installing system dependencies..."

    if [[ "$OS_TYPE" == "debian" ]]; then
        apt-get update
        apt-get install -y \
            curl \
            wget \
            gnupg \
            software-properties-common \
            certbot \
            nginx \
            ufw \
            cron
    elif [[ "$OS_TYPE" == "redhat" ]]; then
        yum update -y
        yum install -y \
            curl \
            wget \
            gnupg \
            certbot \
            nginx \
            firewalld \
            crontabs
    fi

    print_success "System dependencies installed"
}

# Install Hysteria2
install_hysteria2() {
    print_info "Installing Hysteria2..."

    # Install using official script
    bash <(curl -fsSL https://get.hy2.sh/)

    # Verify installation
    if ! command -v hysteria &> /dev/null; then
        print_error "Hysteria2 installation failed"
        exit 1
    fi

    print_success "Hysteria2 installed successfully"
}

# Create directories
create_directories() {
    print_info "Creating directories..."

    mkdir -p "$CONFIG_DIR/sni"
    mkdir -p "$CONFIG_DIR/logs"
    mkdir -p "$INSTALL_DIR"

    chown -R root:root "$CONFIG_DIR"
    chmod 755 "$CONFIG_DIR"
    chmod 700 "$CONFIG_DIR/sni"

    print_success "Directories created"
}

# Validate domains
validate_domains() {
    print_info "Validating domains..."

    IFS=',' read -ra DOMAIN_ARRAY <<< "$DOMAINS"
    for domain in "${DOMAIN_ARRAY[@]}"; do
        domain=$(echo "$domain" | xargs) # trim whitespace
        
        # Check DNS resolution
        if ! dig +short "$domain" &> /dev/null; then
            print_error "Domain $domain does not resolve to any IP address"
            exit 1
        fi
        
        print_success "Domain $domain is valid"
    done
}

# Generate certificates
generate_certificates() {
    print_info "Generating SSL certificates..."

    IFS=',' read -ra DOMAIN_ARRAY <<< "$DOMAINS"
    
    if [[ "$LETSENCRYPT_ENABLED" == true ]]; then
        print_info "Using Let's Encrypt for certificate generation..."
        
        # Stop any web server that might be using port 80
        systemctl stop nginx &> /dev/null || true
        systemctl stop apache2 &> /dev/null || true
        
        for domain in "${DOMAIN_ARRAY[@]}"; do
            domain=$(echo "$domain" | xargs)
            
            print_info "Generating Let's Encrypt certificate for $domain..."
            
            # Generate certificate using certbot
            certbot certonly \
                --standalone \
                --non-interactive \
                --agree-tos \
                --email "$EMAIL" \
                --domains "$domain" \
                --cert-name "$domain"
            
            # Copy certificates to Hysteria2 directory
            cp "/etc/letsencrypt/live/$domain/fullchain.pem" "$CONFIG_DIR/sni/$domain.crt"
            cp "/etc/letsencrypt/live/$domain/privkey.pem" "$CONFIG_DIR/sni/$domain.key"
            
            print_success "Certificate generated for $domain"
        done
        
        # Start nginx again if it was running
        systemctl start nginx &> /dev/null || true
        
    else
        print_info "Generating self-signed certificates..."
        
        for domain in "${DOMAIN_ARRAY[@]}"; do
            domain=$(echo "$domain" | xargs)
            
            # Generate self-signed certificate
            openssl req -x509 \
                -nodes \
                -days 365 \
                -newkey rsa:2048 \
                -keyout "$CONFIG_DIR/sni/$domain.key" \
                -out "$CONFIG_DIR/sni/$domain.crt" \
                -subj "/C=US/ST=State/L=City/O=Organization/CN=$domain"
            
            print_success "Self-signed certificate generated for $domain"
        done
    fi
    
    # Set proper permissions
    chmod 644 "$CONFIG_DIR"/sni/*.crt
    chmod 600 "$CONFIG_DIR"/sni/*.key
    chown root:root "$CONFIG_DIR"/sni/*
}

# Create Hysteria2 configuration
create_config() {
    print_info "Creating Hysteria2 configuration..."

    IFS=',' read -ra DOMAIN_ARRAY <<< "$DOMAINS"
    PRIMARY_DOMAIN="${DOMAIN_ARRAY[0]}"
    
    cat > "$CONFIG_DIR/config.yaml" << EOF
listen: :443

# SNI Configuration
sni:
  enabled: ${SNI_ENABLED}
  default: "${PRIMARY_DOMAIN}"
  domains:
EOF

    if [[ "$SNI_ENABLED" == true ]]; then
        for domain in "${DOMAIN_ARRAY[@]}"; do
            domain=$(echo "$domain" | xargs)
            cat >> "$CONFIG_DIR/config.yaml" << EOF
    - domain: "${domain}"
      cert: "/etc/hysteria2/sni/${domain}.crt"
      key: "/etc/hysteria2/sni/${domain}.key"
EOF
        done
    else
        # Single certificate configuration
        cat >> "$CONFIG_DIR/config.yaml" << EOF
    - domain: "${PRIMARY_DOMAIN}"
      cert: "/etc/hysteria2/sni/${PRIMARY_DOMAIN}.crt"
      key: "/etc/hysteria2/sni/${PRIMARY_DOMAIN}.key"
EOF
    fi

    cat >> "$CONFIG_DIR/config.yaml" << EOF

# Fallback TLS configuration
tls:
  cert: "/etc/hysteria2/sni/${PRIMARY_DOMAIN}.crt"
  key: "/etc/hysteria2/sni/${PRIMARY_DOMAIN}.key"

# Authentication
auth:
  type: password
  password: "$(openssl rand -base64 32)"

# Bandwidth limits
bandwidth:
  up: "500 mbps"
  down: "1 gbps"

# Masquerade (disabled if SNI is enabled for obfuscation compatibility)
$(if [[ "$SNI_ENABLED" == false ]]; then
cat << MASQUERADE
masquerade:
  type: proxy
  proxy:
    url: https://www.google.com
    rewriteHost: true
MASQUERADE
fi)

# Access control
acl:
  file: /etc/hysteria2/acl.txt

# QUIC configuration
quic:
  initStreamReceiveWindow: 8388608
  maxStreamReceiveWindow: 8388608
  initConnReceiveWindow: 20971520
  maxConnReceiveWindow: 20971520
  maxIdleTimeout: 30s
  maxUdpReceiveQueueSize: 1024

# Outbound configuration
outbound:
  name: direct

# Connection limits
maxIncomingStreams: 1000

# Logging
logging:
  level: info
  fields:
    service: hysteria2
    node_id: "${NODE_ID}"
EOF

    print_success "Configuration created"
}

# Setup auto-renewal
setup_auto_renewal() {
    if [[ "$AUTO_RENEW" != true || "$LETSENCRYPT_ENABLED" != true ]]; then
        print_info "Auto-renewal disabled"
        return
    fi
    
    print_info "Setting up automatic certificate renewal..."
    
    # Create renewal script
    cat > "$INSTALL_DIR/renew-certificates.sh" << 'EOF'
#!/bin/bash
# Auto-renewal script for Let's Encrypt certificates

LOG_FILE="/var/log/hysteria2-renewal.log"
CONFIG_DIR="/etc/hysteria2"

echo "$(date): Starting certificate renewal..." >> $LOG_FILE

# Renew certificates
certbot renew --quiet --post-hook "systemctl restart hysteria2"

echo "$(date): Certificate renewal completed" >> $LOG_FILE
EOF
    
    chmod +x "$INSTALL_DIR/renew-certificates.sh"
    
    # Add to cron
    (crontab -l 2>/dev/null; echo "0 3 * * * $INSTALL_DIR/renew-certificates.sh") | crontab -
    
    print_success "Auto-renewal setup completed"
}

# Create systemd service
create_systemd_service() {
    print_info "Creating systemd service..."
    
    cat > "/etc/systemd/system/hysteria2.service" << EOF
[Unit]
Description=Hysteria2 VPN Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/hysteria server -c $CONFIG_DIR/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable hysteria2
    
    print_success "Systemd service created"
}

# Configure firewall
configure_firewall() {
    print_info "Configuring firewall..."
    
    if command -v ufw &> /dev/null; then
        # Ubuntu/Debian with UFW
        ufw allow 443/udp
        ufw allow 443/tcp
        if [[ "$LETSENCRYPT_ENABLED" == true ]]; then
            ufw allow 80/tcp
        fi
        ufw --force enable
    elif command -v firewall-cmd &> /dev/null; then
        # CentOS/RHEL with firewalld
        firewall-cmd --permanent --add-service=https
        if [[ "$LETSENCRYPT_ENABLED" == true ]]; then
            firewall-cmd --permanent --add-service=http
        fi
        firewall-cmd --permanent --add-port=443/udp
        firewall-cmd --reload
    fi
    
    print_success "Firewall configured"
}

# Start services
start_services() {
    print_info "Starting Hysteria2 service..."
    
    systemctl start hysteria2
    
    # Check if service is running
    if systemctl is-active --quiet hysteria2; then
        print_success "Hysteria2 service started successfully"
    else
        print_error "Hysteria2 service failed to start"
        systemctl status hysteria2
        exit 1
    fi
}

# Show completion message
show_completion() {
    print_success "Hysteria2 installation completed!"
    
    echo
    echo "=== Installation Summary ==="
    echo "Node ID: $NODE_ID"
    echo "Domains: $DOMAINS"
    echo "SNI Enabled: $SNI_ENABLED"
    echo "Let's Encrypt: $LETSENCRYPT_ENABLED"
    echo "Auto-renewal: $AUTO_RENEW"
    echo
    echo "=== Configuration Files ==="
    echo "Main Config: $CONFIG_DIR/config.yaml"
    echo "Certificates: $CONFIG_DIR/sni/"
    echo "Logs: journalctl -u hysteria2 -f"
    echo
    echo "=== Commands ==="
    echo "Start service: systemctl start hysteria2"
    echo "Stop service: systemctl stop hysteria2"
    echo "Restart service: systemctl restart hysteria2"
    echo "Check status: systemctl status hysteria2"
    echo "View logs: journalctl -u hysteria2 -f"
    
    if [[ "$LETSENCRYPT_ENABLED" == true ]]; then
        echo "Renew certificates: certbot renew"
        echo "Test renewal: certbot renew --dry-run"
    fi
    
    echo
    echo "=== Client Configuration ==="
    IFS=',' read -ra DOMAIN_ARRAY <<< "$DOMAINS"
    for domain in "${DOMAIN_ARRAY[@]}"; do
        domain=$(echo "$domain" | xargs)
        echo "Connect to: $domain:443"
    done
    
    echo "Password: $(grep 'password:' "$CONFIG_DIR/config.yaml" | awk '{print $2}')"
    echo
}

# Main installation function
main() {
    print_info "Starting Hysteria2 installation with Let's Encrypt support..."
    
    parse_args "$@"
    validate_inputs
    check_requirements
    install_dependencies
    install_hysteria2
    create_directories
    
    if [[ "$LETSENCRYPT_ENABLED" == true ]]; then
        validate_domains
    fi
    
    generate_certificates
    create_config
    setup_auto_renewal
    create_systemd_service
    configure_firewall
    start_services
    show_completion
}

# Run main function
main "$@"