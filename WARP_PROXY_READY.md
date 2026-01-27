# WARP Proxy Implementation - Quick Start Guide

## 🎯 What's Been Implemented

Complete end-to-end VPN -> WARP -> Internet proxy functionality for the Hysteria2 microservices system.

## 📁 New Files Created

### Core Components
- `agent-service/internal/services/warp_manager.go` - WARP client management
- `agent-service/internal/services/traffic_router.go` - Traffic routing through WARP  
- `agent-service/internal/services/warp_monitor.go` - Real-time monitoring
- `agent-service/internal/handlers/warp_proxy_handler.go` - gRPC API handlers
- `agent-service/cmd/warp_proxy_test/main.go` - Comprehensive testing

### API Definitions
- `proto/node_management.proto` - Updated with WARP gRPC methods

### Documentation
- `docs/WARP_INTEGRATION.md` - Complete implementation guide

## 🚀 Quick Start

### 1. Update Agent Configuration

```yaml
# agent-service/configs/agent.yaml
hysteria2:
  warp_enabled: true
  warp_proxy_port: 1080
  warp_auto_connect: true
  warp_client_type: "local"
```

### 2. Build and Run

```bash
# Build agent
cd agent-service
go mod tidy
go build -o bin/agent cmd/agent/main_full.go

# Run with WARP
WARP_ENABLED=true ./bin/agent
```

### 3. Test WARP Integration

```bash
# Run comprehensive test suite
cd agent-service/cmd/warp_proxy_test
go run main.go
```

## 🔧 API Usage Examples

### Setup Complete WARP Proxy

```bash
grpcurl -plaintext -d '{
  "nodeId": "node-1",
  "warpEnabled": true,
  "warpProxyPort": 1080,
  "autoConnect": true,
  "vpnInterface": "eth0",
  "configureHysteria": true,
  "setupTrafficRouting": true
}' localhost:50051 node_management.NodeManager/SetupWARPProxyEndpoint
```

### Check Status

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

## 🌊 Traffic Flow

```
[VPN Client] → [Hysteria2 Server] → [WARP SOCKS5 Proxy] → [Cloudflare Network] → [Internet]
               (Port 8080)           (Port 1080)           (Encrypted)
```

## 📊 Features Implemented

### ✅ WARP Client Management
- Automatic installation (Ubuntu/Debian/RHEL/Docker)
- Connection management (connect/disconnect)
- Proxy mode configuration
- License and organization support

### ✅ Traffic Routing  
- iptables rules for traffic flow
- Network interface management
- ACL configuration
- Masquerading support

### ✅ Real-time Monitoring
- Connection status tracking
- Performance metrics (latency, bandwidth)
- Health scoring (0-100)
- Historical data collection

### ✅ gRPC API
- Complete proxy endpoint setup
- Status monitoring
- Connectivity testing
- Service restart

### ✅ Testing Suite
- Automated end-to-end tests
- Health check validation
- Performance benchmarking
- Integration verification

## 🛠️ Configuration Options

### WARP Settings
```yaml
warp_enabled: true          # Enable WARP integration
warp_proxy_port: 1080       # SOCKS5 proxy port  
warp_auto_connect: true     # Auto-connect on start
warp_client_type: "local"   # "local" or "docker"
warp_license_key: ""        # Optional teams license
warp_organization: ""       # Optional org name
```

### Traffic Settings
```yaml
enable_masquerading: true   # Network masquerading
default_interface: "eth0"   # Network interface
enable_bbr: true           # TCP BBR congestion control
```

### Monitoring Settings
```yaml
metrics:
  collect_interval: 30      # Seconds
  report_interval: 60       # Seconds
```

## 🧪 Testing Commands

```bash
# Test WARP installation
grpcurl localhost:50051 node_management.NodeManager/InstallWARPClient

# Configure WARP
grpcurl -plaintext -d '{
  "enabled": true,
  "proxyPort": 1080
}' localhost:50051 node_management.NodeManager/ConfigureWARP

# Connect to WARP  
grpcurl localhost:50051 node_management.NodeManager/ConnectWARP

# Run full test suite
cd agent-service/cmd/warp_proxy_test && go run main.go
```

## 🔍 Status Monitoring

### Monitor WARP Status
```go
status, _ := warpManager.GetWARPStatus()
fmt.Printf("Connected: %v, IP: %s, Location: %s\n", 
  status.Connected, status.IPAddress, status.Location)
```

### Health Check
```go
health := warpMonitor.RunHealthCheck()
fmt.Printf("Health Score: %d/100, Issues: %v\n", 
  health.Score, health.Checks)
```

## ⚡ Performance Optimizations

### System Tuning
```bash
# Enable IP forwarding
sysctl -w net.ipv4.ip_forward=1

# Optimize network buffers
sysctl -w net.core.rmem_max=16777216
sysctl -w net.core.wmem_max=16777216
```

### WARP Configuration
```bash
# Use teams license for better performance
warp-cli teams license YOUR_LICENSE_KEY

# Set optimal proxy port
warp-cli proxy port 1080
```

## 🐛 Troubleshooting

### Common Issues
1. **WARP not connecting**: Check `warp-cli status`
2. **Proxy not working**: Verify proxy mode with `warp-cli settings mode`
3. **Traffic not routing**: Check iptables rules
4. **High latency**: Test different WARP endpoints

### Debug Commands
```bash
# WARP debug logs
sudo warp-cli --log-level debug status

# Network connections
ss -tulpn | grep :1080

# iptables rules
iptables -t nat -L HYSTERIA2-WARP -n -v
```

## 📈 Monitoring Metrics

The system tracks:
- **Connection Status**: WARP connected/disconnected
- **Performance**: Latency (ms), bandwidth (Mbps)
- **Traffic**: Bytes sent/received
- **Health**: Overall score (0-100)
- **Uptime**: Connection duration
- **Disconnections**: Count and frequency

## 🎯 Production Deployment

### Docker
```dockerfile
FROM ubuntu:22.04
RUN apt update && apt install -y cloudflare-warp
COPY agent-service/bin/agent /usr/local/bin/
EXPOSE 50051 8080
CMD ["/usr/local/bin/agent"]
```

### Kubernetes
```yaml
env:
- name: WARP_ENABLED
  value: "true"
- name: WARP_PROXY_PORT  
  value: "1080"
securityContext:
  privileged: true  # Required for iptables
```

## ✅ Implementation Complete

The WARP proxy integration is now fully implemented and ready for production use. All components are interconnected and tested:

1. ✅ **WARP Client Integration** - Complete
2. ✅ **Traffic Routing** - Complete  
3. ✅ **API Integration** - Complete
4. ✅ **Monitoring System** - Complete
5. ✅ **Testing Suite** - Complete
6. ✅ **Documentation** - Complete

**Traffic flow: VPN Client → Hysteria2 → WARP Proxy → Internet** is now functional! 🚀