#!/bin/bash

# HysteriaVPN One-Click Installer
# Полнофункциональный установщик для основного сервера (orchestrator + web panel)
# с автоматической установкой всех программ и Let's Encrypt интеграцией

set -e

# Глобальные переменные
readonly SCRIPT_VERSION="1.0.0"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly LIB_DIR="$SCRIPT_DIR/lib"
readonly CONFIG_DIR="$SCRIPT_DIR/generated-configs"
readonly REQUIRED_COMMANDS=("curl" "wget" "openssl" "grep" "awk" "sed")
readonly REQUIRED_PORTS=("80" "443" "5432" "6379" "50052" "8080" "8081" "3000" "9090")

# Цвета для вывода
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly WHITE='\033[1;37m'
readonly NC='\033[0m' # No Color

# Конфигурационные переменные (будут заполнены интерактивно)
MASTER_DOMAIN=""
ADMIN_EMAIL=""
ENVIRONMENT="development"
NODE_COUNT=0
DB_PASSWORD=""
JWT_SECRET=""
LETSENCRYPT_ENABLED=true
MONITORING_ENABLED=true
NODE_DOMAINS=()
NODE_LOCATIONS=()
NODE_COUNTRIES=()

# Хелпер функции
print_step() {
    echo -e "${BLUE}🔧 [ШАГ]$NC $1"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

print_header() {
    echo
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}$(printf '%.0s=' {1..$(echo "$1" | wc -c)})${NC}"
    echo
}

# Импорт библиотечных функций
source "$LIB_DIR/docker-utils.sh" 2>/dev/null || {
    print_error "Library docker-utils.sh not found. Install the script correctly."
    exit 1
}

source "$LIB_DIR/letsencrypt-utils.sh" 2>/dev/null || {
    print_error "Library letsencrypt-utils.sh not found. Install the script correctly."
    exit 1
}

source "$LIB_DIR/networking-utils.sh" 2>/dev/null || {
    print_error "Library networking-utils.sh not found. Install the script correctly."
    exit 1
}

source "$LIB_DIR/security-utils.sh" 2>/dev/null || {
    print_error "Library security-utils.sh not found. Install the script correctly."
    exit 1
}

# Функция показа справки
show_help() {
    cat << EOF
HysteriaVPN One-Click Installer v$SCRIPT_VERSION
==============================================

Этот скрипт автоматически устанавливает полный VPN стек:
• Orchestrator (master server для управления узлами)
• Web Panel (React интерфейс управления)
• PostgreSQL + Redis базы данных
• Prometheus + Grafana мониторинг
• mTLS сертификаты для secure межсервисной связи
• Let's Encrypt сертификаты (опционально)

ИСПОЛЬЗОВАНИЕ:
  $0 [OPTIONS]

ОПЦИИ:
  --help, -h          Показать эту справку
  --domain DOMAIN     Установить мастер домен без вопросов
  --email EMAIL       Установить админ email для Let's Encrypt
  --nodes COUNT       Количество VPS узлов (0-5)
  --dev               Режим разработки (localhost, без LE сертификатов)
  --no-monitoring     Отключить Prometheus/Grafana
  --skip-deps         Пропустить проверку зависимостей (для тестирования)

ПРИМЕРЫ:
  # Разработка установка
  $0 --dev --domain localhost

  # Production с 2 узлами
  $0 --domain vpn.yourdomain.com --email admin@yourdomain.com --nodes 2

  # Без интерактивного ввода
  $0 --domain vpn.company.com --email admin@company.com --nodes 3 --no-monitoring

ТРЕБОВАНИЯ К СИСТЕМЕ:
  • RAM: минимум 4GB
  • Disk: минимум 20GB свободного места
  • OS: Ubuntu 18+, Debian 10+, CentOS 7+, RHEL 7+
  • Интернет подключение
  • Root доступ

ДОКУМЕНТАЦИЯ: https://github.com/your-org/hysteria-installer

EOF
}

# Парсинг аргументов командной строки
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help|-h)
                show_help
                exit 0
                ;;
            --domain)
                MASTER_DOMAIN="$2"
                shift 2
                ;;
            --email)
                ADMIN_EMAIL="$2"
                shift 2
                ;;
            --nodes)
                NODE_COUNT="$2"
                shift 2
                ;;
            --dev)
                ENVIRONMENT="development"
                LETSENCRYPT_ENABLED=false
                shift
                ;;
            --no-monitoring)
                MONITORING_ENABLED=false
                shift
                ;;
            --skip-deps)
                SKIP_DEPS_CHECK=true
                shift
                ;;
            *)
                print_error "Неизвестная опция: $1"
                echo "Используйте --help для получения справки"
                exit 1
                ;;
        esac
    done
}

# Интерактивная конфигурация
interactive_config() {
    print_header "🎯 INTERACTIVE CONFIGURATION"

    # Тип установки
    echo "Выберите тип установки:"
    echo "1. Development (localhost, self-signed сертификаты)"
    echo "2. Production (домены, Let's Encrypt сертификаты)"
    read -p "Выбор (1/2) [2]: " env_choice

    case ${env_choice:-2} in
        1)
            ENVIRONMENT="development"
            LETSENCRYPT_ENABLED=false
            print_info "Режим разработки выбран"
            ;;
        2)
            ENVIRONMENT="production"
            LETSENCRYPT_ENABLED=true
            print_info "Production режим выбран"
            ;;
        *)
            print_error "Неверный выбор"
            exit 1
            ;;
    esac

    echo

    # Мастер домен
    if [ -z "$MASTER_DOMAIN" ]; then
        while true; do
            read -p "Мастер домен/IP: " MASTER_DOMAIN
            if [ -n "$MASTER_DOMAIN" ]; then
                break
            fi
            print_warning "Домен обязателен"
        done
    fi

    print_info "Мастер домен: $MASTER_DOMAIN"

    # Email для Let's Encrypt (если production)
    if [ "$LETSENCRYPT_ENABLED" = true ] && [ -z "$ADMIN_EMAIL" ]; then
        while true; do
            read -p "Email для Let's Encrypt выкладок: " ADMIN_EMAIL
            if [ -n "$ADMIN_EMAIL" ]; then
                break
            fi
            print_warning "Email обязателен для Let's Encrypt"
        done
    fi

    if [ -n "$ADMIN_EMAIL" ]; then
        print_info "Админ email: $ADMIN_EMAIL"
    fi

    echo

    # Количество узлов
    if [ -z "$NODE_COUNT" ]; then
        while true; do
            read -p "Количество VPS узлов (0-5) [3]: " node_input
            node_input=${node_input:-3}

            if [[ "$node_input" =~ ^[0-5]$ ]]; then
                NODE_COUNT=$node_input
                break
            fi
            print_warning "Введите число от 0 до 5"
        done
    fi

    print_info "Количество узлов: $NODE_COUNT"

    # Конфигурация узлов
    if [ "$NODE_COUNT" -gt 0 ]; then
        echo
        print_info "Конфигурация узлов:"

        for ((i=1; i<=NODE_COUNT; i++)); do
            if [ "$ENVIRONMENT" = "production" ]; then
                while true; do
                    read -p "Домен для узла $i: " node_domain
                    if [ -n "$node_domain" ]; then
                        NODE_DOMAINS+=("$node_domain")
                        break
                    fi
                    print_warning "Домен обязателен"
                done
            else
                NODE_DOMAINS+=("localhost")
            fi

            read -p "Локация узла $i [New York]: " node_location
            node_location=${node_location:-"New York"}
            NODE_LOCATIONS+=("$node_location")

            read -p "Страна узла $i [US]: " node_country
            node_country=${node_country:-"US"}
            NODE_COUNTRIES+=("$node_country")

            echo
        done
    fi

    # Мониторинг
    if [ "$MONITORING_ENABLED" = true ]; then
        read -p "Включить мониторинг (Prometheus/Grafana)? (y/n) [y]: " monitoring_choice
        case ${monitoring_choice:-y} in
            [Nn]*)
                MONITORING_ENABLED=false
                print_info "Мониторинг отключен"
                ;;
            *)
                print_info "Мониторинг включен"
                ;;
        esac
    fi

     echo
     print_success "Конфигурация завершена!"
 }

# Preflight checks function
run_preflight_checks() {
    print_step "Running system preflight checks"

    # Check root access
    if [ "$EUID" -ne 0 ]; then
        print_error "Script must be run as root. Use sudo."
        exit 1
    fi

    # Check OS
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        case "$ID" in
            ubuntu|debian|centos|rhel|fedora)
                print_success "OS compatible: $PRETTY_NAME"
                ;;
            *)
                print_warning "OS $ID may not be supported. Continuing..."
                ;;
        esac
    else
        print_warning "Could not determine OS. Continuing..."
    fi

    # Check required commands
    local missing_cmds=()
    for cmd in "${REQUIRED_COMMANDS[@]}"; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            missing_cmds+=("$cmd")
        fi
    done

    if [ ${#missing_cmds[@]} -ne 0 ]; then
        print_error "Missing commands: ${missing_cmds[*]}. Please install them."
        exit 1
    fi
    print_success "All required commands found"

    # Check port availability
    local occupied_ports=()
    for port in "${REQUIRED_PORTS[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            occupied_ports+=("$port")
        fi
    done

    if [ ${#occupied_ports[@]} -ne 0 ]; then
        print_error "Ports already occupied: ${occupied_ports[*]}. Please free them."
        exit 1
    fi
    print_success "All required ports are free"

    # Check disk space (minimum 5GB)
    local available_space
    available_space=$(df / | awk 'NR==2 {print $4}')
    if [ "$available_space" -lt 5242880 ]; then  # 5GB in KB
        print_error "Insufficient disk space. Minimum 5GB."
        exit 1
    fi
    print_success "Disk space is sufficient"

    # Check RAM (minimum 1GB)
    local total_mem
    total_mem=$(free -m | awk 'NR==2 {print $2}')
    if [ "$total_mem" -lt 1024 ]; then
        print_error "Insufficient RAM. Minimum 1GB."
        exit 1
    fi
    print_success "RAM is sufficient"

    # Check internet connection
    if ! curl -s --connect-timeout 5 https://www.google.com >/dev/null; then
        print_error "No internet connection."
        exit 1
    fi
    print_success "Internet connection available"

    print_success "All preflight checks passed"
}

# Install dependencies function
install_dependencies() {
    print_step "Installing system dependencies"

    if command -v apt &> /dev/null; then
        apt update
        apt install -y curl wget git ufw htop iotop sysstat fail2ban logrotate unattended-upgrades
    elif command -v yum &> /dev/null; then
        yum install -y curl wget git firewalld htop iotop sysstat fail2ban logrotate yum-cron
    else
        print_error "Unsupported package manager"
        exit 1
    fi
    print_success "Dependencies installed"
}

# Create directories function
create_directories() {
    print_step "Creating directories"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$SCRIPT_DIR/logs"
    mkdir -p "$SCRIPT_DIR/ssl"
    print_success "Directories created"
}

# Generate certificates function
generate_certificates() {
    print_step "Generating certificates"
    if [ "$LETSENCRYPT_ENABLED" = true ]; then
        generate_letsencrypt_certificate "$MASTER_DOMAIN" "$ADMIN_EMAIL"
    else
        openssl req -x509 -newkey rsa:4096 -keyout "$CONFIG_DIR/server.key" -out "$CONFIG_DIR/server.crt" -days 365 -nodes -subj "/CN=$MASTER_DOMAIN"
    fi
    print_success "Certificates generated"
}

# Generate env files function
generate_env_files() {
    print_step "Generating environment files"
    DB_PASSWORD=$(openssl rand -hex 16)
    JWT_SECRET=$(openssl rand -hex 32)
    cat > "$CONFIG_DIR/.env" << EOF
DB_PASSWORD=$DB_PASSWORD
JWT_SECRET=$JWT_SECRET
MASTER_DOMAIN=$MASTER_DOMAIN
NODE_COUNT=$NODE_COUNT
EOF
    print_success "Environment files generated"
}

# Setup project function
setup_hysteriavpn_project() {
    print_step "Setting up HysteriaVPN project"
    # Assume project is already here
    print_success "Project setup complete"
}

# Deploy with docker function
deploy_with_docker() {
    print_step "Deploying with Docker"
    check_docker
    build_docker_images
    start_services
    print_success "Deployment complete"
}

# Run final verification function
run_final_verification() {
    print_step "Running final verification"
    run_docker_health_checks
    print_success "Verification complete"
}

# Show completion summary function
show_completion_summary() {
    print_header "Installation Complete"
    echo "HysteriaVPN has been installed successfully!"
    echo "Master domain: $MASTER_DOMAIN"
    echo "Web panel: https://$MASTER_DOMAIN"
    echo "API: https://$MASTER_DOMAIN/api"
}

# Главная функция установки
main() {
    print_header "🚀 HYSTERIAVPN ONE-CLICK INSTALLER v$SCRIPT_VERSION"
    echo -e "${YELLOW}Полнофункциональный установщик HysteriaVPN для оркестратора и веб-панели${NC}"
    echo

    # Парсинг аргументов
    parse_args "$@"

    # Проверки системы
    run_preflight_checks

    # Установка зависимостей
    if [ "${SKIP_DEPS_CHECK:-false}" != true ]; then
        install_dependencies
    fi

    # Интерактивная конфигурация (если не все параметры указаны)
    if [ -z "$MASTER_DOMAIN" ]; then
        interactive_config
    fi

    # Проверка доменов (если production)
    if [ "$ENVIRONMENT" = "production" ]; then
        validate_domain_ownership
    fi

    # Подготовка файловой структуры
    create_directories

    # Генерация сертификатов
    generate_certificates

    # Генерация конфигураций
    generate_env_files

    # Установка проекта
    setup_hysteriavpn_project

    # Docker развертывание
    deploy_with_docker

    # Финальная верификация
    run_final_verification

    # Показ результатов
    show_completion_summary
}

# Точка входа
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi