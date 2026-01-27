# Hysteria2 VPN Microservices - Installation Guide

## Overview

Hysteria2 is a distributed VPN system designed to bypass DPI (Deep Packet Inspection) filters, particularly effective against Russian internet censorship. The system consists of multiple microservices that can be deployed on separate servers.

## Architecture

- **Central Server**: Orchestrator, API, Web Panel, Database
- **Node Servers**: Individual VPS agents with Hysteria2 protocol

## Quick Start

### Option 1: Automated Installation

#### Central Server Installation
```bash
# Download and run installer
wget https://raw.githubusercontent.com/your-repo/hysteria2/main/scripts/install-central.sh
chmod +x install-central.sh
sudo ./install-central.sh
```

The installer will:
- Install Docker, Docker Compose and system dependencies
- Configure PostgreSQL 17 and Redis 8 databases
- Set up SSL certificates with Let's Encrypt (auto-renewal)
- Deploy orchestrator, API, and web services with nginx reverse proxy
- Configure firewall (UFW/Firewalld) with security rules
- Setup fail2ban for intrusion prevention
- Configure automatic security updates
- Setup monitoring and log rotation
- Configure backup system with cron jobs
- Create admin user and generate secure passwords

#### Node Installation
```bash
# Download and run installer
wget https://raw.githubusercontent.com/your-repo/hysteria2/main/scripts/install-node.sh
chmod +x install-node.sh
sudo ./install-node.sh
```

The installer will:
- Install Docker, Docker Compose and system dependencies
- Configure system optimizations for high-throughput VPN
- Set up Hysteria2 with advanced DPI obfuscation (QUIC, TLS, VLESS)
- Configure kernel parameters for optimal performance
- Setup fail2ban and security hardening
- Configure monitoring and comprehensive logging
- Setup backup system for configuration
- Register node with central server automatically
- Configure firewall with secure rules

### Option 2: Manual Installation

#### Prerequisites

**Central Server Requirements:**
- Ubuntu 20.04+ / CentOS 8+ / Debian 11+
- 2GB RAM minimum, 4GB recommended
- 20GB disk space
- Domain name with DNS configured
- Root or sudo access

**Node Server Requirements:**
- Ubuntu 18.04+ / CentOS 7+ / Debian 10+
- 1GB RAM minimum, 2GB recommended
- 10GB disk space
- Public IP address
- Root or sudo access

#### Central Server Setup

1. **Update system and install dependencies:**
```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git ufw
```

2. **Install Docker:**
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo systemctl enable docker
sudo systemctl start docker
```

3. **Install Docker Compose:**
```bash
sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

4. **Clone repository:**
```bash
cd /opt
sudo git clone https://github.com/your-repo/hysteria2-microservices.git hysteria2-central
cd hysteria2-central
```

5. **Configure environment:**
```bash
sudo cp deployments/docker/.env.example .env
sudo nano .env  # Edit configuration
```

6. **Start services:**
```bash
sudo make docker-run
```

7. **Setup SSL (optional but recommended):**
```bash
sudo certbot certonly --standalone -d your-domain.com
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem ssl/cert.pem
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem ssl/key.pem
```

#### Node Server Setup

1. **Update system:**
```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git ufw jq
```

2. **Install Docker:**
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo systemctl enable docker
sudo systemctl start docker
```

3. **Optimize system for VPN:**
```bash
sudo tee /etc/sysctl.d/99-hysteria2.conf > /dev/null <<EOF
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

sudo sysctl -p /etc/sysctl.d/99-hysteria2.conf
```

4. **Configure firewall:**
```bash
sudo ufw allow 8443/udp
sudo ufw allow 50051/tcp
sudo ufw --force enable
```

5. **Create node configuration:**
```bash
sudo mkdir -p /opt/hysteria2-node
cd /opt/hysteria2-node

# Create docker-compose.yml
sudo tee docker-compose.yml > /dev/null <<EOF
version: '3.8'
services:
  hysteria-agent:
    image: hysteria2/agent:latest
    environment:
      - MASTER_SERVER=your-central-server.com:50052
      - NODE_ID=node-$(hostname)
      - NODE_NAME=$(hostname)
      - HYSTERIA2_LISTEN_PORT=8443
      - HYSTERIA2_ADVANCED_OBFUSCATION_ENABLED=true
    ports:
      - "8443:8443/udp"
      - "50051:50051/tcp"
    restart: unless-stopped
EOF
```

6. **Start node:**
```bash
sudo docker-compose up -d
```

## Configuration

### Central Server Configuration (.env)

```bash
# Database
POSTGRES_DB=hysteria2_db
POSTGRES_USER=hysteria2
POSTGRES_PASSWORD=your-secure-password

# Redis
REDIS_PASSWORD=your-redis-password

# JWT
JWT_SECRET=your-jwt-secret-key

# Domain
DOMAIN=your-domain.com
```

### Node Configuration

Nodes automatically configure themselves with optimal settings for DPI bypass:
- QUIC obfuscation enabled
- TLS fingerprint rotation
- VLESS Reality protocol
- Traffic shaping
- Behavioral randomization

## Management

### Web Interface

Access the web panel at `https://your-domain.com` with admin credentials.

### Command Line

```bash
# Central server commands
cd /opt/hysteria2-central
make health          # Check service health
make logs-all        # View all logs
make db-backup       # Backup database

# Node commands
cd /opt/hysteria2-node
docker-compose logs -f    # View logs
docker-compose restart    # Restart services
```

### API Endpoints

- `GET /health` - Service health check
- `GET /api/nodes` - List VPN nodes
- `POST /api/nodes/{id}/obfuscation/enable-advanced` - Enable obfuscation
- `GET /api/users` - User management

## Security Features

### DPI Bypass Capabilities

1. **QUIC Obfuscation**: Scrambles QUIC packets to avoid detection
2. **TLS Fingerprint Rotation**: Changes TLS signatures dynamically
3. **VLESS Reality**: Masquerades traffic as legitimate HTTPS
4. **Traffic Shaping**: Normalizes packet patterns
5. **Multi-Hop Routing**: Routes through multiple nodes

### Production Security

- JWT authentication
- Database encryption
- SSL/TLS everywhere
- Firewall configuration
- Log sanitization

## Troubleshooting

### Common Issues

**Central server not accessible:**
```bash
# Check services
sudo docker-compose ps

# Check logs
sudo docker-compose logs orchestrator-service

# Check firewall
sudo ufw status
```

**Node not connecting:**
```bash
# Check node logs
sudo docker-compose logs hysteria-agent

# Verify master server connectivity
telnet your-central-server.com 50052

# Check time synchronization
timedatectl status
```

**SSL certificate issues:**
```bash
# Renew certificates
sudo certbot renew

# Restart web service
sudo docker-compose restart web-service
```

### Performance Tuning

**For high-traffic nodes:**
```bash
# Increase Docker resources
sudo tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "default-ulimits": {
    "nofile": {
      "Name": "nofile",
      "Hard": 65536,
      "Soft": 65536
    }
  }
}
EOF

sudo systemctl restart docker
```

## Support

- Documentation: https://github.com/your-repo/hysteria2/wiki
- Issues: https://github.com/your-repo/hysteria2/issues
- Community: [Discord/Telegram link]

## License

This project is licensed under the MIT License.