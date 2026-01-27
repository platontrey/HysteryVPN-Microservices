# Let's Encrypt Integration Guide

## Overview

HysteryVPN now supports automatic Let's Encrypt certificate generation and management for SNI domains. This provides free, trusted SSL certificates that auto-renew automatically.

## Features

### 🚀 Automatic Certificate Generation
- Free SSL/TLS certificates from Let's Encrypt
- Support for multiple domains on single node
- Automatic domain validation
- Fallback to self-signed certificates if Let's Encrypt fails

### 🔄 Auto-Renewal
- Automatic certificate renewal before expiry
- Configurable renewal schedule (default: 30 days before expiry)
- Email notifications for renewal status
- Cron-based scheduling for reliability

### ✅ Domain Validation
- DNS resolution checking
- HTTP challenge validation (port 80)
- Optional DNS TXT record challenge
- Real-time domain ownership verification

## Installation

### Automated Installation (Recommended)

Use the provided installation script for automatic setup:

```bash
# Basic installation with self-signed certificates
./install-with-letsencrypt.sh \
  --node-id node-001 \
  --domains vpn.example.com \
  --sni

# Full Let's Encrypt setup
./install-with-letsencrypt.sh \
  --node-id node-001 \
  --domains vpn1.example.com,vpn2.example.com \
  --email admin@example.com \
  --sni \
  --letsencrypt

# Installation with auto-renewal disabled
./install-with-letsencrypt.sh \
  --node-id node-001 \
  --domains vpn.example.com \
  --email admin@example.com \
  --sni \
  --letsencrypt \
  --no-auto-renew
```

### Manual Setup

1. **Install Certbot**:
```bash
# Ubuntu/Debian
apt-get update && apt-get install -y certbot

# CentOS/RHEL
yum install -y certbot
```

2. **Configure Domain DNS**:
- Ensure domains resolve to your server IP
- Open port 80 for HTTP validation
- Open port 443 for Hysteria2

3. **Generate Certificates**:
```bash
# Single domain
certbot certonly --standalone --email admin@example.com --domains vpn.example.com --agree-tos --non-interactive

# Multiple domains
certbot certonly --standalone --email admin@example.com --domains vpn1.example.com,vpn2.example.com --agree-tos --non-interactive
```

## Web Interface Management

### Enable Let's Encrypt

1. Navigate to **Node Configuration** → **SNI Management** tab
2. Check **"Enable SNI Support"**
3. Check **"Use Let's Encrypt (Auto-SSL certificates)"**
4. Enter your email address
5. Add domains using **"+ Add Domain"** button
6. Click **"🔒 Generate Let's Encrypt Certificates"**

### Configuration Options

| Option | Description | Default |
|--------|-------------|----------|
| **Use Let's Encrypt** | Enable automatic SSL certificates | `false` |
| **Email** | Email for Let's Encrypt notifications | Required for LE |
| **Preferred Challenge** | HTTP-01, DNS-01, or TLS-ALPN-01 | `http-01` |
| **Validate DNS** | Check domain resolution before generation | `true` |
| **Auto-renew certificates** | Enable automatic renewal | `true` |

### Challenge Types

#### HTTP-01 (Recommended)
- Requires port 80 access
- Automatic validation
- Fast and simple
- **Usage**: Standard web servers

#### DNS-01
- Requires DNS provider access
- Manual TXT record creation
- Supports wildcard certificates
- **Usage**: Internal networks, firewall restrictions

#### TLS-ALPN-01
- Advanced method
- Requires port 443 access
- No additional ports needed
- **Usage**: When port 80 is blocked

## API Usage

### Generate Let's Encrypt Certificates

```bash
curl -X POST "https://your-api.com/api/v1/nodes/{nodeId}/sni/certificates/generate-letsencrypt" \
  -H "Content-Type: application/json" \
  -d '{
    "domains": ["vpn1.example.com", "vpn2.example.com"],
    "email": "admin@example.com",
    "preferredChallenge": "http-01",
    "validateDNS": true
  }'
```

### Check Certificate Status

```bash
curl -X GET "https://your-api.com/api/v1/nodes/{nodeId}/sni/certificates"
```

### Validate Domains

```bash
curl -X POST "https://your-api.com/api/v1/nodes/{nodeId}/sni/certificates/validate" \
  -H "Content-Type: application/json" \
  -d '{
    "domains": ["vpn1.example.com", "vpn2.example.com"]
  }'
```

### Renew Certificates

```bash
curl -X POST "https://your-api.com/api/v1/nodes/{nodeId}/sni/certificates/renew"
```

## Configuration Files

### Hysteria2 Config with Let's Encrypt

```yaml
listen: :443

# SNI Configuration with Let's Encrypt certificates
sni:
  enabled: true
  default: "vpn1.example.com"
  domains:
    - domain: "vpn1.example.com"
      cert: "/etc/hysteria/sni/vpn1.example.com.fullchain.pem"
      key: "/etc/hysteria/sni/vpn1.example.com.privkey.pem"
    - domain: "vpn2.example.com"
      cert: "/etc/hysteria/sni/vpn2.example.com.fullchain.pem"
      key: "/etc/hysteria/sni/vpn2.example.com.privkey.pem"

# Fallback TLS
tls:
  cert: "/etc/hysteria/sni/vpn1.example.com.fullchain.pem"
  key: "/etc/hysteria/sni/vpn1.example.com.privkey.pem"

# ... rest of configuration
```

### Environment Variables

```bash
# Enable Let's Encrypt
HYSTERIA2_SNI_LETS_ENCRYPT=true

# Set email for notifications
HYSTERIA2_SNI_EMAIL=admin@example.com

# Preferred challenge type
HYSTERIA2_SNI_PREFERRED_CHALLENGE=http-01

# Enable DNS validation
HYSTERIA2_SNI_VALIDATE_DNS=true
```

## Troubleshooting

### Common Issues

#### 1. "Domain does not resolve to server IP"
**Cause**: DNS A/AAAA record doesn't point to server
**Solution**: Update DNS records to point to server IP

#### 2. "Port 80 not accessible"
**Cause**: Firewall blocking port 80 or port already in use
**Solution**: 
```bash
# Open port 80
ufw allow 80/tcp

# Stop conflicting services
systemctl stop apache2 nginx
```

#### 3. "Too many certificates issued for this domain"
**Cause**: Let's Encrypt rate limits
**Solution**: Wait for rate limit to reset (1 week for duplicate certificates)

#### 4. "Certificate validation failed"
**Cause**: Domain validation challenges failing
**Solution**: Check DNS propagation and server accessibility

### Debug Commands

```bash
# Check DNS resolution
dig +short vpn.example.com

# Test HTTP accessibility
curl -I http://vpn.example.com

# Check certbot logs
journalctl -u certbot -f

# Manual certificate generation with debug
certbot certonly --standalone --dry-run -v --domains vpn.example.com

# Test renewal
certbot renew --dry-run
```

### Certificate Locations

```
/etc/letsencrypt/live/
├── vpn1.example.com/
│   ├── cert.pem       # Server certificate
│   ├── chain.pem      # Intermediate certificate
│   ├── fullchain.pem  # Full certificate chain
│   └── privkey.pem    # Private key
└── vpn2.example.com/
    ├── cert.pem
    ├── chain.pem
    ├── fullchain.pem
    └── privkey.pem

# Hysteria2 certificate directory
/etc/hysteria/sni/
├── vpn1.example.com.crt    # Copied from fullchain.pem
├── vpn1.example.com.key     # Copied from privkey.pem
├── vpn2.example.com.crt
└── vpn2.example.com.key
```

## Security Considerations

### 🔐 Private Key Protection
- Private keys stored with 600 permissions
- Root ownership only
- Automatic backup to secure location

### 📧 Email Privacy
- Email used only for certificate notifications
- Never shared with third parties
- Optional for internal deployments

### 🌐 Network Security
- Temporary port 80 opening only during validation
- Automatic firewall rules management
- Challenge-specific port requirements

## Monitoring and Maintenance

### Certificate Monitoring
```bash
# Check certificate expiry
for domain in /etc/hysteria/sni/*.crt; do
  domain_name=$(basename "$domain" .crt)
  expiry=$(openssl x509 -in "$domain" -noout -enddate | cut -d= -f2)
  echo "$domain_name expires on: $expiry"
done
```

### Auto-Renewal Verification
```bash
# Test renewal process
certbot renew --dry-run

# Check cron jobs
crontab -l | grep certbot

# View renewal logs
tail -f /var/log/letsencrypt/letsencrypt.log
```

## Best Practices

### 🎯 Domain Management
1. **Use separate subdomains** for each VPN server
2. **Maintain DNS records** and ensure proper propagation
3. **Monitor certificate expiry** even with auto-renewal
4. **Test renewal process** in staging environment first

### 🛡️ Security
1. **Restrict access** to certificate directories
2. **Use strong email** for recovery notifications
3. **Monitor renewal logs** for failures
4. **Implement backup** certificate strategy

### 📈 Performance
1. **Enable HTTP challenge** for fastest validation
2. **Use wildcard certificates** only when necessary
3. **Optimize cron schedule** to avoid peak hours
4. **Monitor system resources** during renewal

## Integration Examples

### Nginx Reverse Proxy
```nginx
server {
    listen 80;
    server_name vpn1.example.com vpn2.example.com;
    
    # Let's Encrypt validation
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    # Redirect to Hysteria2
    location / {
        proxy_pass http://127.0.0.1:443;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Docker Integration
```dockerfile
FROM hysteria2:latest

# Install certbot
RUN apt-get update && apt-get install -y certbot

# Copy startup script
COPY setup-certificates.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/setup-certificates.sh

# Set environment variables
ENV HYSTERIA2_SNI_LETS_ENCRYPT=true
ENV HYSTERIA2_SNI_EMAIL=admin@example.com

# Run setup on startup
CMD ["/usr/local/bin/setup-certificates.sh", "--daemon"]
```

This comprehensive Let's Encrypt integration provides enterprise-grade SSL certificate management for your Hysteria2 VPN nodes!