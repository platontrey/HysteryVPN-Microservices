#!/bin/bash

# Security Utilities Library
# Functions for mTLS certificate generation, security hardening, and access control

# Generate Certificate Authority (CA) for mTLS
generate_ca_cert() {
    local ca_dir="$1"
    local ca_name="${2:-HysteriaVPN-CA}"
    local validity_days="${3:-3650}"  # 10 years default
    
    print_step "Generating Certificate Authority"
    
    mkdir -p "$ca_dir"
    
    # CA private key
    openssl genrsa -out "$ca_dir/ca.key" 4096 2>/dev/null
    chmod 600 "$ca_dir/ca.key"
    
    # CA certificate
    openssl req -new -x509 -days "$validity_days" \
        -key "$ca_dir/ca.key" \
        -out "$ca_dir/ca.crt" \
        -subj "/C=US/ST=State/L=City/O=HysteriaVPN/OU=Security/CN=$ca_name" 2>/dev/null
    
    # Generate certificate serial number file
    echo "1000" > "$ca_dir/ca.srl"
    
    print_success "Certificate Authority generated: $ca_name"
    print_info "CA Certificate: $ca_dir/ca.crt"
    print_info "CA Private Key: $ca_dir/ca.key"
}

# Generate server certificate signed by CA
generate_server_cert() {
    local cert_name="$1"
    local ca_dir="$2"
    local domain="${3:-localhost}"
    local validity_days="${4:-365}"
    
    print_step "Generating server certificate for $cert_name"
    
    # Server private key
    openssl genrsa -out "$ca_dir/$cert_name.key" 2048 2>/dev/null
    chmod 600 "$ca_dir/$cert_name.key"
    
    # Certificate signing request (CSR)
    openssl req -new -key "$ca_dir/$cert_name.key" \
        -out "$ca_dir/$cert_name.csr" \
        -subj "/C=US/ST=State/L=City/O=HysteriaVPN/OU=Services/CN=$domain" 2>/dev/null
    
    # Sign certificate with CA
    openssl x509 -req -days "$validity_days" \
        -in "$ca_dir/$cert_name.csr" \
        -CA "$ca_dir/ca.crt" \
        -CAkey "$ca_dir/ca.key" \
        -CAcreateserial \
        -out "$ca_dir/$cert_name.crt" \
        -extfile <(echo "subjectAltName=DNS:$domain,IP:127.0.0.1") 2>/dev/null
    
    # Clean up CSR
    rm -f "$ca_dir/$cert_name.csr"
    
    print_success "Server certificate generated: $cert_name"
}

# Generate client certificate for agents
generate_client_cert() {
    local client_name="$1"
    local ca_dir="$2"
    local validity_days="${3:-365}"
    
    print_step "Generating client certificate for $client_name"
    
    # Client private key
    openssl genrsa -out "$ca_dir/$client_name.key" 2048 2>/dev/null
    chmod 600 "$ca_dir/$client_name.key"
    
    # Client CSR
    openssl req -new -key "$ca_dir/$client_name.key" \
        -out "$ca_dir/$client_name.csr" \
        -subj "/C=US/ST=State/L=City/O=HysteriaVPN/OU=Agents/CN=$client_name" 2>/dev/null
    
    # Sign certificate with CA
    openssl x509 -req -days "$validity_days" \
        -in "$ca_dir/$client_name.csr" \
        -CA "$ca_dir/ca.crt" \
        -CAkey "$ca_dir/ca.key" \
        -CAcreateserial \
        -out "$ca_dir/$client_name.crt" 2>/dev/null
    
    # Clean up CSR
    rm -f "$ca_dir/$client_name.csr"
    
    print_success "Client certificate generated: $client_name"
}

# Setup complete mTLS infrastructure for all services
setup_mtls_infrastructure() {
    print_header "🔐 MTLS SECURITY SETUP"
    
    local certs_dir="$SCRIPT_DIR/generated-configs/certs"
    mkdir -p "$certs_dir/ca"
    mkdir -p "$certs_dir/services"
    mkdir -p "$certs_dir/clients"
    
    # Generate CA
    generate_ca_cert "$certs_dir/ca" "HysteriaVPN-InternalCA" 3650
    
    # Generate service certificates
    print_step "Generating service certificates"
    
    # Orchestrator
    generate_server_cert "orchestrator-server" "$certs_dir/ca" "orchestrator.local" 365
    
    # API Service
    generate_server_cert "api-service" "$certs_dir/ca" "api.local" 365
    
    # Web Service
    generate_server_cert "web-service" "$certs_dir/ca" "web.local" 365
    
    # Agent Client Certificate
    generate_client_cert "agent-client" "$certs_dir/ca" 365
    
    print_success "All mTLS certificates generated"
    
    # Set proper permissions
    chmod -R 600 "$certs_dir"
    chmod 644 "$certs_dir/ca/ca.crt"
    
    print_info "CA Certificate: $certs_dir/ca/ca.crt"
    print_info "Service certs: $certs_dir/services/"
    print_info "Client certs: $certs_dir/clients/"
}

# Configure UFW firewall for security
setup_ufw_security() {
    if ! command -v ufw >/dev/null 2>&1; then
        print_info "UFW not installed, skipping firewall setup"
        return 0
    fi
    
    print_step "Configuring UFW firewall"
    
    # Reset existing rules
    ufw --force reset
    
    # Default policies
    ufw default deny incoming
    ufw default allow outgoing
    
    # Allow SSH (important to not lock yourself out)
    ufw allow 22/tcp comment 'SSH'
    
    # Allow HTTP/HTTPS
    ufw allow 80/tcp comment 'HTTP'
    ufw allow 443/tcp comment 'HTTPS'
    ufw allow 443/udp comment 'Hysteria2 VPN'
    
    # Allow database ports (internal only)
    ufw allow from 172.16.0.0/12 to any port 5432 comment 'PostgreSQL internal'
    ufw allow from 172.16.0.0/12 to any port 6379 comment 'Redis internal'
    
    # Allow application ports (internal)
    ufw allow from 172.16.0.0/12 to any port 8080 comment 'API Service'
    ufw allow from 172.16.0.0/12 to any port 8081 comment 'Orchestrator'
    ufw allow from 172.16.0.0/12 to any port 3000 comment 'Web Interface'
    ufw allow from 172.16.0.0/12 to any port 50052 comment 'gRPC'
    
    # Enable firewall
    ufw --force enable
    
    print_success "UFW firewall configured"
    ufw status verbose
}

# Configure Firewalld for security (CentOS/RHEL)
setup_firewalld_security() {
    if ! command -v firewall-cmd >/dev/null 2>&1; then
        print_info "Firewalld not installed, skipping firewall setup"
        return 0
    fi
    
    print_step "Configuring Firewalld firewall"
    
    # Create hysteria-vpn zone
    firewall-cmd --permanent --new-zone=hysteria-vpn 2>/dev/null || true
    
    # Allow services in hysteria-vpn zone
    firewall-cmd --permanent --zone=hysteria-vpn --add-service=http
    firewall-cmd --permanent --zone=hysteria-vpn --add-service=https
    firewall-cmd --permanent --zone=hysteria-vpn --add-port=443/udp
    
    # Allow internal services
    firewall-cmd --permanent --zone=hysteria-vpn --add-port=5432/tcp
    firewall-cmd --permanent --zone=hysteria-vpn --add-port=6379/tcp
    firewall-cmd --permanent --zone=hysteria-vpn --add-port=8080/tcp
    firewall-cmd --permanent --zone=hysteria-vpn --add-port=8081/tcp
    firewall-cmd --permanent --zone=hysteria-vpn --add-port=3000/tcp
    firewall-cmd --permanent --zone=hysteria-vpn --add-port=50052/tcp
    
    # Apply settings
    firewall-cmd --reload
    
    print_success "Firewalld configured"
}

# Setup rate limiting with iptables
setup_rate_limiting() {
    print_step "Setting up rate limiting protection"
    
    # Rate limit new TCP connections to port 443
    iptables -A INPUT -p tcp --dport 443 -m connlimit --connlimit-above 50 -j DROP 2>/dev/null || true
    
    # Rate limit new connections (SYN flood protection)
    iptables -A INPUT -p tcp --syn -m limit --limit 5/s --limit-burst 10 -j ACCEPT 2>/dev/null || true
    
    # Save iptables rules
    if command -v iptables-save >/dev/null 2>&1; then
        iptables-save > /etc/iptables/rules.v4 2>/dev/null || true
        print_success "Rate limiting rules applied"
    else
        print_warning "Could not save iptables rules"
    fi
}

# Install and configure Fail2Ban
setup_fail2ban() {
    print_step "Setting up Fail2Ban intrusion protection"
    
    if ! command -v fail2ban-client >/dev/null 2>&1; then
        print_info "Installing Fail2Ban..."
        
        if [[ -f /etc/debian_version ]]; then
            apt-get update
            apt-get install -y fail2ban
        elif [[ -f /etc/redhat-release ]]; then
            yum install -y fail2ban
        fi
    fi
    
    # Create jail configuration for HysteriaVPN
    cat > /etc/fail2ban/jail.local << EOF
[hysteria-api]
enabled = true
filter = hysteria-api
logpath = /var/log/hysteria2/api.log
maxretry = 5
findtime = 600
bantime = 3600

[hysteria-orchestrator]
enabled = true
filter = hysteria-orchestrator
logpath = /var/log/hysteria2/orchestrator.log
maxretry = 5
findtime = 600
bantime = 3600
EOF
    
    # Create filter for API abuse
    cat > /etc/fail2ban/filter.d/hysteria-api.conf << EOF
[Definition]
failregex = ^.*"status":401.*$
            ^.*"status":403.*$
            ^.*"status":429.*$
ignoreregex =
EOF
    
    # Create filter for orchestrator abuse
    cat > /etc/fail2ban/filter.d/hysteria-orchestrator.conf << EOF
[Definition]
failregex = ^.*authentication failed.*$
            ^.*invalid auth token.*$
            ^.*connection rejected.*$
ignoreregex =
EOF
    
    # Restart Fail2Ban
    systemctl restart fail2ban
    systemctl enable fail2ban
    
    print_success "Fail2Ban configured and started"
}

# Security hardening checklist
run_security_hardening() {
    print_header "🛡️ SECURITY HARDENING"
    
    local checks=0
    local passed=0
    
    # Check 1: Secure passwords
    ((checks++))
    if [[ -n "$DB_PASSWORD" ]] && [[ ${#DB_PASSWORD} -ge 16 ]]; then
        print_success "Database password: Strong"
        ((passed++))
    else
        print_warning "Database password: Too weak (use minimum 16 chars)"
    fi
    
    # Check 2: JWT secret strength
    ((checks++))
    if [[ -n "$JWT_SECRET" ]] && [[ ${#JWT_SECRET} -ge 32 ]]; then
        print_success "JWT secret: Strong"
        ((passed++))
    else
        print_warning "JWT secret: Too weak (use minimum 32 chars)"
    fi
    
    # Check 3: Node auth token
    ((checks++))
    if [[ -n "$NODE_AUTH_TOKEN" ]] && [[ ${#NODE_AUTH_TOKEN} -ge 16 ]]; then
        print_success "Node auth token: Strong"
        ((passed++))
    else
        print_warning "Node auth token: Too weak"
    fi
    
    # Check 4: SSL certificates
    ((checks++))
    if [[ "$LETSENCRYPT_ENABLED" == true ]]; then
        print_success "SSL: Let's Encrypt enabled"
        ((passed++))
    else
        print_info "SSL: Self-signed (OK for development)"
        ((passed++))
    fi
    
    # Check 5: mTLS enabled
    ((checks++))
    if [[ -f "$SCRIPT_DIR/generated-configs/certs/ca/ca.crt" ]]; then
        print_success "mTLS: Configured"
        ((passed++))
    else
        print_warning "mTLS: Not configured"
    fi
    
    # Summary
    echo
    print_info "Security hardening: $passed/$checks checks passed"
    
    if [[ $passed -eq $checks ]]; then
        print_success "Security hardening complete ✓"
    else
        print_warning "Review security recommendations above"
    fi
}

# Generate secure random password
generate_secure_password() {
    local length="${1:-32}"
    openssl rand -base64 "$length" | tr -d '=+/' | cut -c1-"$length"
}

# Validate certificate chain
validate_cert_chain() {
    local cert_file="$1"
    local ca_file="$2"
    
    print_step "Validating certificate chain"
    
    # Verify certificate with CA
    if openssl verify -CAfile "$ca_file" "$cert_file" 2>/dev/null; then
        print_success "Certificate chain is valid"
        
        # Show certificate details
        local subject=$(openssl x509 -noout -subject -in "$cert_file" 2>/dev/null | sed 's/^subject=//')
        local issuer=$(openssl x509 -noout -issuer -in "$cert_file" 2>/dev/null | sed 's/^issuer=//')
        local expiry=$(openssl x509 -noout -enddate -in "$cert_file" 2>/dev/null | cut -d= -f2)
        
        print_info "Subject: $subject"
        print_info "Issuer: $issuer"
        print_info "Expires: $expiry"
        
        return 0
    else
        print_error "Certificate chain validation failed"
        return 1
    fi
}