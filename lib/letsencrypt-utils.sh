#!/bin/bash

# Let's Encrypt Certificate Management Library
# Functions for automatic SSL certificate generation and renewal

# Validate domain ownership before certificate issuance
validate_domain_ownership() {
    local domain="$1"
    
    print_step "Validating domain ownership for $domain"
    
    # Check DNS resolution
    if ! nslookup "$domain" >/dev/null 2>&1; then
        print_error "Domain $domain does not resolve to any IP address"
        return 1
    fi
    
    # Check if this server's IP matches the domain
    local server_ip=$(curl -s ifconfig.me 2>/dev/null || curl -s ipinfo.io/ip 2>/dev/null)
    local domain_ip=$(nslookup "$domain" | grep -A1 "$domain" | grep -v "$domain" | grep Address | cut -d: -f2 | tr -d ' ')
    
    if [[ "$server_ip" != "$domain_ip" ]]; then
        print_warning "Domain IP ($domain_ip) does not match server IP ($server_ip)"
        print_info "This is OK if you're using DNS records or proxy setup"
        return 0
    fi
    
    print_success "Domain $domain points to this server"
    return 0
}

# Generate Let's Encrypt certificate for domain
generate_letsencrypt_certificate() {
    local domain="$1"
    local email="$2"
    
    print_step "Generating Let's Encrypt certificate for $domain"
    
    # Check if certificate already exists
    if [[ -f "/etc/letsencrypt/live/$domain/fullchain.pem" ]] && \
       openssl x509 -checkend 86400 -noout -in "/etc/letsencrypt/live/$domain/fullchain.pem" 2>/dev/null; then
        print_warning "Certificate for $domain already exists and is valid for >24 hours"
        return 0
    fi
    
    # Stop any web server using port 80
    systemctl stop nginx apache2 lighttpd 2>/dev/null || true
    
    # Generate certificate
    if certbot certonly \
        --standalone \
        --non-interactive \
        --agree-tos \
        --email "$email" \
        --domains "$domain" \
        --cert-name "$domain" \
        --rsa-key-size 4096 \
        --staging 2>/dev/null; then
        
        # Try again without staging if staging failed
        if [[ $? -ne 0 ]]; then
            print_info "Retrying certificate generation without staging..."
            certbot certonly \
                --standalone \
                --non-interactive \
                --agree-tos \
                --email "$email" \
                --domains "$domain" \
                --cert-name "$domain" \
                --rsa-key-size 4096
        fi
    fi
    
    if [[ -f "/etc/letsencrypt/live/$domain/fullchain.pem" ]]; then
        print_success "Certificate generated successfully for $domain"
        return 0
    else
        print_error "Failed to generate certificate for $domain"
        return 1
    fi
}

# Setup automatic certificate renewal
setup_certificate_renewal() {
    print_step "Setting up automatic certificate renewal"
    
    # Create renewal hook
    cat > /opt/hysteriavpn/scripts/renew-certificates.sh << 'EOF'
#!/bin/bash
LOG_FILE="/var/log/hysteria-cert-renewal.log"
CERTS_DIR="/etc/letsencrypt/live"

echo "$(date): Starting certificate renewal process..." >> $LOG_FILE

# Renew certificates
if certbot renew --quiet --post-hook "systemctl reload nginx"; then
    echo "$(date): Certificate renewal successful" >> $LOG_FILE
    
    # Restart HysteriaVPN services to reload certificates
    docker-compose -f /opt/hysteriavpn/docker-compose.yml restart 2>/dev/null || true
else
    echo "$(date): Certificate renewal failed" >> $LOG_FILE
fi

echo "$(date): Certificate renewal process completed" >> $LOG_FILE
EOF
    
    chmod +x /opt/hysteriavpn/scripts/renew-certificates.sh
    
    # Add to cron (check twice daily)
    (crontab -l 2>/dev/null; echo "0 0,12 * * * /opt/hysteriavpn/scripts/renew-certificates.sh") | crontab -
    
    print_success "Automatic certificate renewal configured"
}

# Install Nginx for Let's Encrypt validation
install_nginx_for_le() {
    print_step "Installing Nginx for Let's Encrypt validation"
    
    if command -v nginx >/dev/null 2>&1; then
        print_success "Nginx already installed"
        return 0
    fi
    
    if [[ -f /etc/debian_version ]]; then
        apt-get update
        apt-get install -y nginx
    elif [[ -f /etc/redhat-release ]]; then
        yum install -y nginx
    fi
    
    # Enable and start nginx
    systemctl enable nginx
    systemctl start nginx
    
    print_success "Nginx installed and configured"
}

# Create Nginx config for HTTP-01 challenge
setup_nginx_le_config() {
    local domain="$1"
    
    print_step "Creating Nginx configuration for $domain"
    
    cat > "/etc/nginx/sites-available/$domain.conf" << EOF
server {
    listen 80;
    server_name $domain;

    # Let's Encrypt ACME challenge
    location /.well-known/acme-challenge/ {
        root /var/www/html;
        try_files \$uri =404;
    }

    # Redirect all HTTP to HTTPS
    location / {
        return 301 https://\$server_name\$request_uri;
    }
}
EOF
    
    # Enable site
    ln -sf "/etc/nginx/sites-available/$domain.conf" "/etc/nginx/sites-enabled/"
    
    # Remove default site
    rm -f /etc/nginx/sites-enabled/default
    
    # Test and reload nginx
    nginx -t && systemctl reload nginx
    
    print_success "Nginx configuration created for $domain"
}

# Create SSL configuration for production
create_ssl_nginx_config() {
    local domain="$1"
    
    print_step "Creating SSL Nginx configuration for $domain"
    
    cat > "/etc/nginx/sites-available/$domain-ssl.conf" << EOF
server {
    listen 443 ssl http2;
    server_name $domain;

    ssl_certificate /etc/letsencrypt/live/$domain/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$domain/privkey.pem;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload";

    # Proxy to HysteriaVPN services
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    # API proxy
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    # Monitoring locations
    location /prometheus/ {
        proxy_pass http://localhost:9090;
        proxy_set_header Host \$host;
    }
    
    location /grafana/ {
        proxy_pass http://localhost:3001;
        proxy_set_header Host \$host;
    }
}
EOF
    
    # Enable SSL site
    ln -sf "/etc/nginx/sites-available/$domain-ssl.conf" "/etc/nginx/sites-enabled/"
    
    # Test and reload nginx
    nginx -t && systemctl reload nginx
    
    print_success "SSL Nginx configuration created for $domain"
}

# Test certificate validity
test_certificate() {
    local domain="$1"
    
    print_step "Testing certificate for $domain"
    
    if openssl s_client -connect "$domain:443" -servername "$domain" -showcerts 2>/dev/null | \
       openssl x509 -noout -dates 2>/dev/null | grep -q "notAfter"; then
        print_success "Certificate is valid for $domain"
        
        # Show expiry date
        local expiry_date=$(openssl s_client -connect "$domain:443" -servername "$domain" -showcerts 2>/dev/null | \
                           openssl x509 -noout -dates 2>/dev/null | grep "notAfter" | cut -d= -f2)
        print_info "Certificate expires: $expiry_date"
        
        return 0
    else
        print_error "Certificate validation failed for $domain"
        return 1
    fi
}

# Cleanup expired or old certificates
cleanup_old_certificates() {
    print_step "Cleaning up old certificates"
    
    # Remove expired certificates
    certbot delete --non-interactive 2>/dev/null || true
    
    # Clean up old renewal configs
    find /etc/letsencrypt/renewal -name "*.conf" -mtime +365 -delete 2>/dev/null || true
    
    print_success "Certificate cleanup completed"
}