#!/bin/bash

# Networking Utilities Library
# Functions for network checks, DNS validation, and connectivity testing

# Check if required ports are available
check_ports() {
    local ports=("$@")
    
    print_step "Checking required ports availability"
    
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            print_warning "Port $port is already in use"
            local process=$(lsof -Pi :$port -sTCP:LISTEN -t | tail -1 | awk '{print $1}')
            echo -e "${YELLOW}  Process: $process${NC}"
            echo -e "${YELLOW}  To free the port: kill -9 $(lsof -t -i :$port)${NC}"
        else
            print_success "Port $port is available"
        fi
    done
}

# Validate DNS resolution
validate_dns() {
    local domain="$1"
    local expected_ip="$2"  # Optional
    
    print_step "Validating DNS resolution for $domain"
    
    # Try multiple DNS tools
    local resolved_ip=""
    
    if command -v dig >/dev/null 2>&1; then
        resolved_ip=$(dig +short "$domain" A 2>/dev/null | head -1)
    elif command -v nslookup >/dev/null 2>&1; then
        resolved_ip=$(nslookup "$domain" 2>/dev/null | grep -A1 "Name:" | grep "Address:" | head -1 | awk '{print $2}')
    elif command -v host >/dev/null 2>&1; then
        resolved_ip=$(host -t A "$domain" 2>/dev/null | grep "has address" | awk '{print $4}')
    fi
    
    if [[ -z "$resolved_ip" ]]; then
        print_error "Domain $domain does not resolve to any IP address"
        print_info "Check your DNS records and wait for propagation"
        return 1
    fi
    
    print_success "Domain $domain resolves to: $resolved_ip"
    
    # Check against expected IP if provided
    if [[ -n "$expected_ip" ]]; then
        if [[ "$resolved_ip" == "$expected_ip" ]]; then
            print_success "DNS matches expected IP: $expected_ip"
        else
            print_warning "Resolved IP ($resolved_ip) differs from expected ($expected_ip)"
            print_info "This may indicate DNS propagation delay or misconfiguration"
        fi
    fi
    
    # Get this server's IP for comparison
    local server_ip=$(get_server_public_ip)
    print_info "This server's public IP: $server_ip"
    
    if [[ "$resolved_ip" == "$server_ip" ]]; then
        print_success "DNS points to this server ✓"
    else
        print_warning "DNS points to different server IP"
        print_info "Ensure your domain's A record points to: $server_ip"
    fi
    
    return 0
}

# Get server's public IP address
get_server_public_ip() {
    local ip=""
    
    # Try multiple services
    local services=(
        "https://api.ipify.org?format=text"
        "https://ipinfo.io/ip"
        "https://icanhazip.com"
        "https://ifconfig.me"
    )
    
    for service in "${services[@]}"; do
        ip=$(curl -s --max-time 5 "$service" 2>/dev/null)
        if [[ -n "$ip" ]] && [[ "$ip" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
            echo "$ip"
            return 0
        fi
    done
    
    print_error "Could not determine server's public IP"
    return 1
}

# Check internet connectivity
check_internet() {
    print_step "Checking internet connectivity"
    
    local test_hosts=("8.8.8.8" "1.1.1.1" "google.com" "cloudflare.com")
    local reachable=0
    
    for host in "${test_hosts[@]}"; do
        if ping -c 1 -W 2 "$host" &>/dev/null; then
            ((reachable++))
        fi
    done
    
    if [[ $reachable -ge 2 ]]; then
        print_success "Internet connectivity verified ($reachable/${#test_hosts[@]} hosts reachable)"
        return 0
    else
        print_error "Internet connectivity issues detected ($reachable/${#test_hosts[@]} hosts reachable)"
        return 1
    fi
}

# Check network interface configuration
check_network_config() {
    print_step "Checking network configuration"
    
    # Get primary interface
    local primary_interface=$(ip route | grep default | awk '{print $5}')
    print_info "Primary interface: $primary_interface"
    
    # Get interface IP
    local interface_ip=$(ip addr show "$primary_interface" | grep inet | head -1 | awk '{print $2}' | cut -d/ -f1)
    print_info "Interface IP: $interface_ip"
    
    # Check MTU
    local mtu=$(ip link show "$primary_interface" | grep -o 'mtu [0-9]*' | awk '{print $2}')
    print_info "MTU: $mtu"
    
    # Check if interface is up
    if ip link show "$primary_interface" | grep -q "state UP"; then
        print_success "Primary interface is up and configured"
    else
        print_error "Primary interface is not UP"
        return 1
    fi
    
    return 0
}

# Check firewall status
check_firewall() {
    print_step "Checking firewall configuration"
    
    local firewall_active=false
    local firewall_type=""
    
    # Check UFW (Ubuntu/Debian)
    if command -v ufw >/dev/null 2>&1; then
        firewall_type="UFW"
        if ufw status | grep -q "Status: active"; then
            firewall_active=true
        fi
        
        # Check required ports
        local required_ports=("80" "443" "5432" "6379" "8080" "8081" "3000" "50052")
        for port in "${required_ports[@]}"; do
            if ufw status | grep -q "$port"; then
                local status=$(ufw status | grep "$port" | awk '{print $2}')
                print_info "Port $port: $status"
            else
                print_warning "Port $port: NOT CONFIGURED"
            fi
        done
    fi
    
    # Check Firewalld (CentOS/RHEL)
    if command -v firewall-cmd >/dev/null 2>&1; then
        firewall_type="Firewalld"
        if firewall-cmd --state >/dev/null 2>&1; then
            firewall_active=true
        fi
        
        # Check active zones
        local active_zones=$(firewall-cmd --get-active-zones 2>/dev/null)
        print_info "Active zones: $active_zones"
        
        # Check open ports
        local open_ports=$(firewall-cmd --list-ports 2>/dev/null)
        print_info "Open ports: $open_ports"
    fi
    
    # Check iptables directly
    if ! $firewall_active && command -v iptables >/dev/null 2>&1; then
        local iptables_rules=$(iptables -L -n | wc -l)
        if [[ $iptables_rules -gt 0 ]]; then
            firewall_type="iptables"
            firewall_active=true
            print_info "iptables rules found: $iptables_rules"
        fi
    fi
    
    if $firewall_active; then
        print_success "Firewall is active: $firewall_type"
    else
        print_warning "No active firewall detected"
        print_info "Consider enabling a firewall for security"
    fi
    
    return 0
}

# Test latency to target host
test_latency() {
    local host="$1"
    local count="${2:-5}"
    
    print_step "Testing latency to $host"
    
    if command -v ping >/dev/null 2>&1; then
        local result=$(ping -c "$count" "$host" 2>&1 | tail -2)
        local avg_latency=$(echo "$result" | grep "avg" | awk -F'/' '{print $5}')
        local packet_loss=$(echo "$result" | grep "packet loss" | awk '{print $6}')
        
        print_success "Average latency: ${avg_latency}ms"
        print_info "Packet loss: $packet_loss"
        
        if [[ $(echo "$avg_latency < 100" | bc -l) -eq 1 ]]; then
            print_success "Good latency to $host"
        else
            print_warning "High latency to $host (${avg_latency}ms)"
        fi
    else
        print_warning "ping command not available for latency testing"
    fi
}

# Check DNS resolution speed
check_dns_performance() {
    local domain="$1"
    
    print_step "Testing DNS resolution speed for $domain"
    
    local dns_servers=("8.8.8.8" "1.1.1.1" "208.67.222.222")
    
    for server in "${dns_servers[@]}"; do
        local start_time=$(date +%s%N)
        local result=$(dig @"$server" +short "$domain" 2>/dev/null)
        local end_time=$(date +%s%N)
        local duration=$((end_time - start_time))
        
        if [[ -n "$result" ]]; then
            print_success "$server: ${duration}ms"
        else
            print_warning "$server: FAILED"
        fi
    done
}

# Validate network requirements for HysteriaVPN
validate_hysteria_requirements() {
    print_header "🌐 HYSTERIA2 NETWORK REQUIREMENTS"
    
    local required_ports=("443/udp" "80/tcp")
    local issues=0
    
    # Check UDP 443 (Hysteria2 default)
    if command -v nc >/dev/null 2>&1; then
        print_step "Testing UDP port 443 connectivity"
        if timeout 2 nc -uvz localhost 443 2>/dev/null; then
            print_success "UDP 443 is reachable locally"
        else
            print_warning "UDP 443 may not be properly configured"
            ((issues++))
        fi
    fi
    
    # Check internet speed (basic test)
    print_step "Testing basic internet connectivity"
    local test_result=$(curl -s --max-time 10 -w "%{speed_download}" -o /dev/null http://speedtest.tele2.net/100mb.dat 2>/dev/null)
    if [[ -n "$test_result" ]] && (( $(echo "$test_result > 0.5" | bc -l) )); then
        local speed_mb=$(echo "scale=2; $test_result * 8 / 1000000" | bc)
        print_success "Internet speed detected: ${speed_mb} Mbps"
    else
        print_warning "Could not determine internet speed"
        print_info "Hysteria2 requires stable internet connection"
        ((issues++))
    fi
    
    # Check for QUIC support
    print_step "Checking QUIC protocol support"
    if command -v ss >/dev/null 2>&1; then
        if ss -H state established '( dport = 443 or sport = 443 )' | grep -q "quic"; then
            print_success "QUIC connections detected on port 443"
        else
            print_info "No active QUIC connections (normal if not yet running)"
        fi
    fi
    
    # Summary
    echo
    if [[ $issues -eq 0 ]]; then
        print_success "All network requirements validated ✓"
        return 0
    else
        print_warning "Found $issues network issue(s)"
        print_info "Review firewall and network configuration"
        return 1
    fi
}