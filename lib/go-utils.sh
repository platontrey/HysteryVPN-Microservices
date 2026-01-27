#!/bin/bash

# Go Utilities Library
# Functions for Go development, dependency management, and build optimization

# Check if Go is installed and meets minimum version requirements
check_go_version() {
    print_step "Checking Go installation"

    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.19+ first."
        return 1
    fi

    local go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+\.[0-9]+' | cut -d'o' -f2)
    print_info "Go version: $go_version"

    # Check minimum version (1.19)
    if ! [[ "$go_version" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
        print_error "Cannot parse Go version: $go_version"
        return 1
    fi

    local major="${BASH_REMATCH[1]}"
    local minor="${BASH_REMATCH[2]}"

    if [[ $major -lt 1 ]] || [[ $major -eq 1 && $minor -lt 19 ]]; then
        print_error "Go version $go_version is too old. Minimum required: 1.19"
        return 1
    fi

    print_success "Go version $go_version is compatible"
    return 0
}

# Download and verify Go modules for a service
download_go_modules() {
    local service_dir="$1"
    local service_name="${2:-$(basename "$service_dir")}"

    print_step "Downloading Go modules for $service_name"

    if [[ ! -d "$service_dir" ]]; then
        print_error "Service directory not found: $service_dir"
        return 1
    fi

    cd "$service_dir" || return 1

    # Check if go.mod exists
    if [[ ! -f "go.mod" ]]; then
        print_error "go.mod not found in $service_dir"
        return 1
    fi

    # Set up Go environment for reliable downloads
    export GO111MODULE=on
    export GOSUMDB=sum.golang.org
    export GOPROXY="https://proxy.golang.org,direct"

    # Configure Go caches for faster builds
    setup_go_cache

    # Download dependencies with timeout and retry
    local max_retries=3
    local retry_count=0

    while [[ $retry_count -lt $max_retries ]]; do
        print_info "Attempting to download Go modules (attempt $((retry_count + 1))/$max_retries)"

        if go mod download -x; then
            # Verify modules integrity
            if go mod verify; then
                print_success "Go modules downloaded and verified for $service_name"
                cd - >/dev/null 2>&1
                return 0
            else
                print_warning "Module verification failed, retrying..."
            fi
        else
            print_warning "Module download failed, retrying..."
        fi

        ((retry_count++))
        [[ $retry_count -lt $max_retries ]] && sleep 2
    done

    print_error "Failed to download Go modules for $service_name after $max_retries attempts"
    cd - >/dev/null 2>&1
    return 1
}

# Setup Go cache directories for better performance
setup_go_cache() {
    export GOCACHE="${GOCACHE:-$HOME/.cache/go-build}"
    export GOMODCACHE="${GOMODCACHE:-$HOME/go/pkg/mod}"
    export GOROOT="${GOROOT:-$(go env GOROOT)}"

    # Create cache directories
    mkdir -p "$GOCACHE" "$GOMODCACHE"

    print_info "Go cache directories configured"
    print_info "GOCACHE: $GOCACHE"
    print_info "GOMODCACHE: $GOMODCACHE"
}

# Verify Go modules integrity and versions
verify_go_modules() {
    local service_dir="$1"
    local service_name="${2:-$(basename "$service_dir")}"

    print_step "Verifying Go modules for $service_name"

    if [[ ! -d "$service_dir" ]]; then
        print_error "Service directory not found: $service_dir"
        return 1
    fi

    cd "$service_dir" || return 1

    # Check if go.mod exists
    if [[ ! -f "go.mod" ]]; then
        print_error "go.mod not found in $service_dir"
        return 1
    fi

    local errors=0

    # Verify module integrity
    if ! go mod verify >/dev/null 2>&1; then
        print_error "Module integrity verification failed"
        ((errors++))
    else
        print_success "Module integrity verified"
    fi

    # Check for outdated dependencies
    print_info "Checking for module updates..."
    local outdated_modules=$(go list -u -m all 2>/dev/null | grep '\[.*\]' | wc -l)
    if [[ $outdated_modules -gt 0 ]]; then
        print_info "$outdated_modules modules have available updates"
    else
        print_success "All modules are up to date"
    fi

    # Check for replace directives that might indicate dependency issues
    if grep -q "^replace" go.mod; then
        print_warning "Found replace directives in go.mod - verify they are necessary"
    fi

    cd - >/dev/null 2>&1
    return $errors
}

# Tidy and optimize go.mod file
optimize_go_modules() {
    local service_dir="$1"
    local service_name="${2:-$(basename "$service_dir")}"

    print_step "Optimizing Go modules for $service_name"

    if [[ ! -d "$service_dir" ]]; then
        print_error "Service directory not found: $service_dir"
        return 1
    fi

    cd "$service_dir" || return 1

    if [[ ! -f "go.mod" ]]; then
        print_error "go.mod not found in $service_dir"
        return 1
    fi

    # Run go mod tidy
    if go mod tidy -v; then
        print_success "Go modules optimized for $service_name"
    else
        print_error "Failed to optimize Go modules"
        cd - >/dev/null 2>&1
        return 1
    fi

    cd - >/dev/null 2>&1
    return 0
}

# Download dependencies for all Go services in the project
download_all_go_modules() {
    print_header "📦 DOWNLOADING GO MODULES FOR ALL SERVICES"

    local services=("orchestrator-service" "api-service" "agent-service")
    local failed_services=()

    for service in "${services[@]}"; do
        if [[ -d "$service" ]]; then
            if ! download_go_modules "$service" "$service"; then
                failed_services+=("$service")
            fi
        else
            print_warning "Service directory not found: $service (skipping)"
        fi
    done

    if [[ ${#failed_services[@]} -eq 0 ]]; then
        print_success "All Go modules downloaded successfully"
        return 0
    else
        print_error "Failed to download modules for: ${failed_services[*]}"
        return 1
    fi
}

# Verify dependencies for all Go services
verify_all_go_modules() {
    print_header "🔍 VERIFYING GO MODULES FOR ALL SERVICES"

    local services=("orchestrator-service" "api-service" "agent-service")
    local errors=0

    for service in "${services[@]}"; do
        if [[ -d "$service" ]]; then
            if ! verify_go_modules "$service" "$service"; then
                ((errors++))
            fi
        else
            print_warning "Service directory not found: $service (skipping)"
        fi
    done

    if [[ $errors -eq 0 ]]; then
        print_success "All Go modules verified successfully"
        return 0
    else
        print_error "$errors service(s) had module verification issues"
        return 1
    fi
}

# Clean Go build cache and modules cache
clean_go_cache() {
    print_step "Cleaning Go caches"

    # Clean module cache
    if command -v go >/dev/null 2>&1; then
        print_info "Cleaning Go module cache..."
        go clean -modcache 2>/dev/null || true

        print_info "Cleaning Go build cache..."
        go clean -cache 2>/dev/null || true

        print_info "Cleaning test cache..."
        go clean -testcache 2>/dev/null || true
    fi

    # Remove cache directories if they exist
    if [[ -d "$HOME/.cache/go-build" ]]; then
        print_info "Removing go-build cache directory..."
        rm -rf "$HOME/.cache/go-build"
    fi

    print_success "Go caches cleaned"
}

# Run go vet on all services
vet_all_services() {
    print_header "🔧 RUNNING GO VET ON ALL SERVICES"

    local services=("orchestrator-service" "api-service" "agent-service")
    local errors=0

    for service in "${services[@]}"; do
        if [[ -d "$service" ]]; then
            print_step "Running go vet on $service"
            cd "$service" || continue

            if go vet ./... 2>&1; then
                print_success "go vet passed for $service"
            else
                print_error "go vet failed for $service"
                ((errors++))
            fi

            cd - >/dev/null 2>&1
        else
            print_warning "Service directory not found: $service (skipping)"
        fi
    done

    if [[ $errors -eq 0 ]]; then
        print_success "All services passed go vet checks"
        return 0
    else
        print_error "$errors service(s) failed go vet"
        return 1
    fi
}

# Get Go build information for a service
get_go_build_info() {
    local service_dir="$1"
    local service_name="${2:-$(basename "$service_dir")}"

    if [[ ! -d "$service_dir" ]]; then
        return 1
    fi

    cd "$service_dir" || return 1

    if [[ ! -f "go.mod" ]]; then
        cd - >/dev/null 2>&1
        return 1
    fi

    echo "### $service_name ###"
    echo "Go Modules:"
    go list -m all | head -10
    echo
    echo "Build Info:"
    echo "Go version: $(go version | cut -d' ' -f3)"
    echo "Modules count: $(go list -m all | wc -l)"

    if [[ -f "go.sum" ]]; then
        echo "Dependencies checksums: $(wc -l < go.sum) lines"
    fi

    cd - >/dev/null 2>&1
}

# Setup Go environment and path for Docker builds
setup_go_build_environment() {
    print_step "Setting up Go build environment"

    # Set Go environment variables
    export CGO_ENABLED=0
    export GOOS="${GOOS:-linux}"
    export GOARCH="${GOARCH:-amd64}"

    print_info "Build environment: GOOS=$GOOS, GOARCH=$GOARCH, CGO_ENABLED=$CGO_ENABLED"

    # Ensure Go binary is in PATH
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go binary not found in PATH"
        return 1
    fi

    print_success "Go build environment ready"
}</content>
<parameter name="filePath">lib/go-utils.sh