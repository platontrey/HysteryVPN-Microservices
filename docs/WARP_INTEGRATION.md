# WARP Proxy Integration Documentation

## Overview

This document describes the complete integration of Cloudflare WARP with the Hysteria2 VPN microservices system, enabling end-to-end traffic routing: **VPN Client → Hysteria2 Server → WARP Proxy → Internet**.

## Architecture

```
[VPN Client] → [Hysteria2 Server] → [WARP SOCKS5 Proxy] → [Cloudflare WARP] → [Internet]
               (VPS Node)          (Port 1080)           (Encrypted Tunnel)
```

### Components

1. **WARP Manager** (`warp_manager.go`)
   - WARP client installation and configuration
   - Connection management
   - Proxy mode setup
   - License management

2. **Traffic Router** (`traffic_router.go`)
   - iptables rules configuration
   - Traffic routing through WARP
   - ACL management

3. **WARP Monitor** (`warp_monitor.go`)
   - Real-time status monitoring
   - Health checks
   - Performance metrics
   - Historical data collection

4. **WARP Proxy Handler** (`warp_proxy_handler.go`)
   - gRPC API endpoints
   - Comprehensive proxy management
   - End-to-end testing

## Installation and Setup

### Prerequisites

- Linux-based VPS (Ubuntu/Debian/RHEL/CentOS)
- Root/sudo privileges
- Internet connectivity
- Hysteria2 installed

### 1. Automatic WARP Installation

```bash
# Via gRPC API
grpcurl -plaintext localhost:50051 node_management.NodeManager/InstallWARPClient

# Via environment variables
export WARP_ENABLED=true
export WARP_AUTO_CONNECT=true
export WARP_PROXY_PORT=1080
```

### 2. Manual WARP Installation

```bash
# Ubuntu/Debian
curl -fsSL https://pkg.cloudflareclient.com/pubkey.gpg | sudo gpg --yes --dearmor --output /usr/share/keyrings/cloudflare-warp-archive-keyring.gpg
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/cloudflare-warp-archive-keyring.gpg] https://pkg.cloudflareclient.com/ $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/cloudflare-client.list
sudo apt update
sudo apt install -y cloudflare-warp

# RHEL/CentOS
sudo rpm --import https://pkg.cloudflareclient.com/pubkey.gpg
echo '[cloudflare-warp]
name=Cloudflare WARP
baseurl=https://pkg.cloudflareclient.com/rhel/$releasever/x86_64
gpgcheck=1
gpgkey=https://pkg.cloudflareclient.com/pubkey.gpg
enabled=1' | sudo tee /etc/yum.repos.d/cloudflare-warp.repo
sudo yum install -y cloudflare-warp
```

### 3. WARP Registration and Connection

```bash
# Register WARP client
sudo warp-cli registration new

# Connect to WARP
sudo warp-cli connect

# Enable proxy mode
sudo warp-cli mode proxy
sudo warp-cli proxy port 1080

# Verify connection
sudo warp-cli status
```

## Configuration

### Environment Variables

```yaml
# config.yaml
hysteria2:
  warp_enabled: true
  warp_proxy_port: 1080
  warp_auto_connect: true
  warp_notify_on_fail: true
  warp_client_type: "local"  # "local" or "docker"
  warp_license_key: ""       # Optional for teams
  warp_organization: ""      # Optional for teams
```

### gRPC API Configuration

```protobuf
message SetupWARPProxyEndpointRequest {
  bool warp_enabled = true;
  int32 warp_proxy_port = 1080;
  bool auto_connect = true;
  string vpn_interface = "eth0";
  bool configure_hysteria = true;
  bool setup_traffic_routing = true;
  bool enable_masquerading = true;
}
```

## Traffic Flow Configuration

### 1. Hysteria2 Configuration

```json
{
  "listen": ":8080",
  "tls": {
    "cert": "/etc/hysteria/cert.pem",
    "key": "/etc/hysteria/key.pem"
  },
  "auth": {
    "type": "password",
    "password": "your-password"
  },
  "outbound": {
    "name": "warp-proxy",
    "type": "socks5",
    "addr": "127.0.0.1:1080"
  },
  "acl": {
    "file": "/etc/hysteria/acl.yaml"
  }
}
```

### 2. ACL Configuration

```yaml
# /etc/hysteria/acl.yaml
# Bypass local networks
- src, 127.0.0.1/8, dst, 127.0.0.1/8
- src, 10.0.0.0/8, dst, 10.0.0.0/8
- src, 172.16.0.0/12, dst, 172.16.0.0/12
- src, 192.168.0.0/16, dst, 192.168.0.0/16

# Route all other traffic through WARP
# (handled by outbound proxy configuration)
```

### 3. iptables Rules

```bash
# Enable IP forwarding
sysctl -w net.ipv4.ip_forward=1

# Create WARP chain
iptables -t nat -N HYSTERIA2-WARP

# Bypass local traffic
iptables -t nat -A HYSTERIA2-WARP -d 127.0.0.0/8 -j RETURN
iptables -t nat -A HYSTERIA2-WARP -d 10.0.0.0/8 -j RETURN
iptables -t nat -A HYSTERIA2-WARP -d 172.16.0.0/12 -j RETURN
iptables -t nat -A HYSTERIA2-WARP -d 192.168.0.0/16 -j RETURN

# Redirect HTTP/HTTPS to WARP proxy
iptables -t nat -A HYSTERIA2-WARP -p tcp --dport 80 -j REDIRECT --to-ports 1080
iptables -t nat -A HYSTERIA2-WARP -p tcp --dport 443 -j REDIRECT --to-ports 1080

# Apply to OUTPUT
iptables -t nat -A OUTPUT -j HYSTERIA2-WARP
```

## API Usage

### Setup Complete WARP Proxy Endpoint

```bash
grpcurl -plaintext -d '{
  "nodeId": "node-1",
  "warpEnabled": true,
  "warpProxyPort": 1080,
  "autoConnect": true,
  "vpnInterface": "eth0",
  "configureHysteria": true,
  "setupTrafficRouting": true,
  "enableMasquerading": true
}' localhost:50051 node_management.NodeManager/SetupWARPProxyEndpoint
```

### Get WARP Proxy Status

```bash
grpcurl -plaintext -d '{
  "nodeId": "node-1"
}' localhost:50051 node_management.NodeManager/GetWARPProxyStatus
```

### Test Connectivity

```bash
grpcurl -plaintext -d '{
  "nodeId": "node-1"
}' localhost:50051 node_management.NodeManager/TestWARPProxyConnectivity
```

## Monitoring and Health Checks

### Monitoring Metrics

- **Connection Status**: WARP connected/disconnected
- **Performance**: Latency, bandwidth (up/down)
- **Traffic**: Bytes sent/received
- **Health Score**: 0-100 overall health rating
- **Uptime**: Connection duration and disconnection count

### Health Check Components

1. **WARP Client Installation**: Verify WARP is installed
2. **WARP Connection**: Check connection status
3. **Proxy Functionality**: Test SOCKS5 proxy
4. **Internet Reachability**: Verify connectivity through WARP
5. **DNS Resolution**: Test DNS queries

### Monitoring API

```go
// Start monitoring
monitor.Start(ctx)

// Get current status
status := monitor.GetCurrentStatus()

// Run health check
health := monitor.RunHealthCheck()

// Get historical data
history, _ := monitor.GetHistoricalData(24 * time.Hour)
```

## Testing

### Automated Testing

```bash
# Run comprehensive test suite
cd agent-service/cmd/warp_proxy_test
go run main.go

# Set custom agent address
AGENT_ADDRESS=localhost:50051 go run main.go
```

### Manual Testing

1. **Install WARP Client**
   ```bash
   grpcurl -plaintext localhost:50051 node_management.NodeManager/InstallWARPClient
   ```

2. **Configure WARP**
   ```bash
   grpcurl -plaintext -d '{
     "enabled": true,
     "proxyPort": 1080,
     "autoConnect": true
   }' localhost:50051 node_management.NodeManager/ConfigureWARP
   ```

3. **Connect to WARP**
   ```bash
   grpcurl -plaintext localhost:50051 node_management.NodeManager/ConnectWARP
   ```

4. **Setup Complete Proxy**
   ```bash
   grpcurl -plaintext -d '{
     "warpEnabled": true,
     "setupTrafficRouting": true
   }' localhost:50051 node_management.NodeManager/SetupWARPProxyEndpoint
   ```

5. **Test End-to-End**
   ```bash
   curl --proxy socks5://127.0.0.1:1080 https://1.1.1.1
   ```

## Troubleshooting

### Common Issues

1. **WARP Installation Failed**
   ```bash
   # Check OS compatibility
   lsb_release -a
   
   # Install prerequisites
   apt update && apt install -y curl gnupg
   ```

2. **WARP Connection Failed**
   ```bash
   # Check registration
   sudo warp-cli registration new
   
   # Check status
   sudo warp-cli status
   
   # Restart WARP service
   sudo systemctl restart warp-svc
   ```

3. **Proxy Not Working**
   ```bash
   # Check proxy mode
   sudo warp-cli settings mode
   
   # Enable proxy mode
   sudo warp-cli mode proxy
   sudo warp-cli proxy port 1080
   ```

4. **Traffic Not Routing**
   ```bash
   # Check IP forwarding
   sysctl net.ipv4.ip_forward
   
   # Check iptables rules
   iptables -t nat -L HYSTERIA2-WARP -n -v
   ```

5. **High Latency**
   ```bash
   # Check WARP server location
   sudo warp-cli status
   
   # Test different WARP endpoints
   curl -w "@curl-format.txt" -o /dev/null -s https://1.1.1.1
   ```

### Debug Commands

```bash
# WARP debug logs
sudo warp-cli --log-level debug status

# Network connections
ss -tulpn | grep :1080

# Traffic flow
tcpdump -i any -n port 1080

# iptables tracing
iptables -t nat -A OUTPUT -j LOG --log-prefix "WARP-DEBUG: "
```

## Performance Optimization

### Bandwidth Optimization

```yaml
hysteria2:
  up_mbps: 1000
  down_mbps: 1000
  warp_proxy_port: 1080
```

### System Tuning

```bash
# Increase connection limits
echo 'net.core.somaxconn = 65535' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 65535' >> /etc/sysctl.conf

# Optimize buffer sizes
echo 'net.core.rmem_max = 16777216' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 16777216' >> /etc/sysctl.conf

# Apply changes
sysctl -p
```

### WARP Optimization

```bash
# Set WARP license for teams (if available)
warp-cli teams license YOUR_LICENSE_KEY

# Configure organization
warp-cli teams gateway YOUR_ORG_NAME
```

## Security Considerations

1. **Access Control**: Limit WARP proxy to localhost
2. **Authentication**: Use strong Hysteria2 passwords
3. **Certificates**: Valid SSL certificates for Hysteria2
4. **Monitoring**: Regular health checks and monitoring
5. **Logging**: Comprehensive logging for audit trails

## Deployment

### Docker Deployment

```dockerfile
FROM ubuntu:22.04

# Install WARP
RUN curl -fsSL https://pkg.cloudflareclient.com/pubkey.gpg | gpg --yes --dearmor --output /usr/share/keyrings/cloudflare-warp-archive-keyring.gpg && \
    echo "deb [arch=amd64 signed-by=/usr/share/keyrings/cloudflare-warp-archive-keyring.gpg] https://pkg.cloudflareclient.com/ jammy main" | tee /etc/apt/sources.list.d/cloudflare-client.list && \
    apt update && apt install -y cloudflare-warp

# Install Hysteria2
RUN bash <(curl -fsSL https://get.hy2.sh/)

# Copy agent
COPY agent-service/bin/agent /usr/local/bin/
COPY configs/ /etc/hysteria/

EXPOSE 50051 8080
CMD ["/usr/local/bin/agent"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: warp-vpn-node
spec:
  replicas: 3
  selector:
    matchLabels:
      app: warp-vpn-node
  template:
    metadata:
      labels:
        app: warp-vpn-node
    spec:
      containers:
      - name: agent
        image: hysteria2-warp-agent:latest
        env:
        - name: WARP_ENABLED
          value: "true"
        - name: WARP_PROXY_PORT
          value: "1080"
        - name: HYSTERIA2_LISTEN_PORT
          value: "8080"
        ports:
        - containerPort: 8080
        - containerPort: 50051
        securityContext:
          privileged: true  # Required for iptables
```

## Support and Maintenance

### Regular Maintenance

1. **Update WARP**: `apt update && apt upgrade cloudflare-warp`
2. **Monitor Health**: Check health scores daily
3. **Review Logs**: Analyze connection patterns
4. **Performance Tests**: Monthly bandwidth tests

### Monitoring Alerts

```yaml
alerts:
  - name: WARPDisconnected
    condition: warp_connected == false
    duration: 5m
    action: Restart WARP service
    
  - name: HighLatency
    condition: latency_ms > 500
    duration: 10m
    action: Check WARP endpoint
    
  - name: LowBandwidth
    condition: bandwidth_mbps < 1
    duration: 15m
    action: Investigate network issues
```

## Conclusion

The WARP proxy integration provides a robust, secure, and performant solution for routing VPN traffic through Cloudflare's global network. The implementation includes comprehensive monitoring, health checks, and automated testing to ensure reliable operation.

For additional support or questions, refer to the project documentation or open an issue in the repository.