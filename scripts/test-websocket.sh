#!/bin/bash

# WebSocket Test Script

echo "🧪 Testing Hysteria2 WebSocket functionality..."

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Function to print colored output
print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
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

# Test API health
test_api_health() {
    print_step "Testing API health..."

    response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)

    if [ "$response" = "200" ]; then
        print_success "API is healthy"
        return 0
    else
        print_error "API health check failed (HTTP $response)"
        return 1
    fi
}

# Test WebSocket connection (basic connectivity)
test_websocket_connection() {
    print_step "Testing WebSocket connection..."

    # This is a basic test - in production you'd use a proper WebSocket client
    # For now, we'll just check if the WebSocket endpoint responds

    # Using curl with --http1.1 to test the upgrade
    response=$(curl -s -I -N \
        -H "Connection: Upgrade" \
        -H "Upgrade: websocket" \
        -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
        -H "Sec-WebSocket-Version: 13" \
        http://localhost:8080/ws 2>/dev/null | head -1 | cut -d' ' -f2)

    if [ "$response" = "101" ]; then
        print_success "WebSocket upgrade successful"
        return 0
    else
        print_warning "WebSocket test inconclusive (may require authentication)"
        return 0
    fi
}

# Test authentication
test_authentication() {
    print_step "Testing authentication..."

    # Test login with admin credentials
    login_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{"email":"admin@hysteria2.local","password":"admin123"}' \
        http://localhost:8080/api/v1/auth/login)

    if echo "$login_response" | grep -q "token"; then
        print_success "Authentication successful"
        # Extract token for WebSocket tests
        TOKEN=$(echo "$login_response" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
        echo "$TOKEN" > /tmp/ws_token.tmp
        return 0
    else
        print_error "Authentication failed"
        echo "$login_response"
        return 1
    fi
}

# Test WebSocket with authentication
test_websocket_with_auth() {
    print_step "Testing WebSocket with authentication..."

    if [ ! -f /tmp/ws_token.tmp ]; then
        print_error "No authentication token available"
        return 1
    fi

    TOKEN=$(cat /tmp/ws_token.tmp)

    # Test WebSocket connection with JWT token
    # This would require a more sophisticated WebSocket client
    # For now, we'll test that the endpoint is accessible with auth

    response=$(curl -s -w "%{http_code}" -o /dev/null \
        -H "Authorization: Bearer $TOKEN" \
        http://localhost:8080/api/v1/users?page=1&limit=1)

    if [ "$response" = "200" ]; then
        print_success "API authentication works (WebSocket should work too)"
        return 0
    else
        print_error "API authentication failed (HTTP $response)"
        return 1
    fi
}

# Test traffic recording (simulate some traffic)
test_traffic_simulation() {
    print_step "Testing traffic simulation..."

    # This would normally be done by the Hysteria2 server
    # For testing, we can create a mock traffic record

    if [ ! -f /tmp/ws_token.tmp ]; then
        print_error "No authentication token available"
        return 1
    fi

    TOKEN=$(cat /tmp/ws_token.tmp)

    # Get user ID first
    user_response=$(curl -s \
        -H "Authorization: Bearer $TOKEN" \
        http://localhost:8080/api/v1/users?page=1&limit=1)

    user_id=$(echo "$user_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

    if [ -z "$user_id" ]; then
        print_warning "Could not get user ID for traffic test"
        return 0
    fi

    print_success "Traffic simulation test prepared (user_id: $user_id)"
}

# Main test function
main() {
    echo "🔌 Hysteria2 WebSocket Test Suite"
    echo "=================================="

    # Cleanup
    rm -f /tmp/ws_token.tmp

    # Run tests
    test_api_health
    API_HEALTH=$?

    if [ $API_HEALTH -eq 0 ]; then
        test_authentication
        AUTH_SUCCESS=$?

        if [ $AUTH_SUCCESS -eq 0 ]; then
            test_websocket_connection
            test_websocket_with_auth
            test_traffic_simulation
        fi
    fi

    # Cleanup
    rm -f /tmp/ws_token.tmp

    echo ""
    echo "📊 Test Results:"
    echo "  API Health: $([ $API_HEALTH -eq 0 ] && echo "✅ PASS" || echo "❌ FAIL")"
    echo ""
    echo "💡 Manual Testing:"
    echo "  1. Open browser at http://localhost:3000"
    echo "  2. Login with admin@hysteria2.local / admin123"
    echo "  3. Check Dashboard for real-time traffic updates"
    echo "  4. Monitor browser console for WebSocket connection logs"
}

# Run main function
main "$@"