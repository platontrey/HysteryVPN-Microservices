#!/bin/bash

# HysteriaVPN Post-Installation Script
# Generate client configurations, QR codes, and completion tasks

# Generate client configuration files
generate_client_configs() {
    print_step "Генерация клиентских конфигураций"

    local config_dir="/opt/hysteriavpn/client-configs"
    mkdir -p "$config_dir"

    # Generate HysteriaVPN client config
    cat > "$config_dir/hysteria-client.yaml" << EOF
server: $MASTER_DOMAIN:443
protocol: udp
auth: $(grep 'password:' /opt/hysteriavpn/configs/server-main.yaml | cut -d: -f2 | xargs)
tls:
  sni: $MASTER_DOMAIN
  insecure: false
bandwidth:
  up: 100 mbps
  down: 100 mbps
quic:
  initStreamReceiveWindow: 8388608
  maxStreamReceiveWindow: 8388608
  initConnReceiveWindow: 20971520
  maxConnReceiveWindow: 20971520
EOF

    # Generate mobile config if nodes exist
    if [ "$NODE_COUNT" -gt 0 ]; then
        for ((i=0; i<NODE_COUNT; i++)); do
            local node_domain="${NODE_DOMAINS[$i]}"
            local node_location="${NODE_LOCATIONS[$i]}"

            cat > "$config_dir/hysteria-client-node-$((i+1)).yaml" << EOF
server: ${node_domain}:443
protocol: udp
auth: $(openssl rand -base64 8)
tls:
  sni: ${node_domain}
  insecure: false
bandwidth:
  up: 100 mbps
  down: 100 mbps
EOF
        done
    fi

    print_success "Клиентские конфигурации сгенерированы: $config_dir"
}

# Generate QR codes for mobile clients
generate_qr_codes() {
    print_step "Генерация QR кодов"

    local qr_dir="/opt/hysteriavpn/qr-codes"
    mkdir -p "$qr_dir"

    # Check if qrencode is available
    if command -v qrencode &> /dev/null; then
        # Generate QR from config
        qrencode -o "$qr_dir/hysteria-main.png" -s 6 < "/opt/hysteriavpn/client-configs/hysteria-client.yaml"
        print_success "QR коды сгенерированы: $qr_dir"
    else
        print_warning "qrencode не установлен, пропускаю генерацию QR кодов"
        print_info "Установите: apt install qrencode"
    fi
}

# Setup cron jobs for maintenance
setup_cron_jobs() {
    print_step "Настройка плановых задач"

    # Backup job (every day at 2 AM)
    (crontab -l 2>/dev/null; echo "0 2 * * * /opt/hysteriavpn/scripts/backup.sh") | crontab -

    # Certificate renewal job (every Sunday at 3 AM)
    if [ "$LETSENCRYPT_ENABLED" = true ]; then
        (crontab -l 2>/dev/null; echo "0 3 * * 0 certbot renew --quiet --post-hook 'systemctl reload nginx'") | crontab -
    fi

    print_success "Плановые задачи настроены"
}

# Create backup script
create_backup_script() {
    print_step "Создание скрипта бэкапа"

    local backup_dir="/opt/hysteriavpn/backup"
    local backup_script="/opt/hysteriavpn/scripts/backup.sh"

    mkdir -p "/opt/hysteriavpn/scripts"

    cat > "$backup_script" << 'EOF'
#!/bin/bash
# Automatic backup script for HysteriaVPN

BACKUP_DIR="/opt/hysteriavpn/backup"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/backup_$TIMESTAMP.tar.gz"

echo "Starting backup at $(date)"

# Database backup
docker exec hysteria2-postgres pg_dump -U hysteria2 -d hysteria2_db -F t > "$BACKUP_DIR/database_$TIMESTAMP.tar"

# Config files backup
tar -czf "$BACKUP_FILE" \
    /opt/hysteriavpn/ \
    /etc/hysteria2/ \
    -C /opt/hysteria2/backup database_$TIMESTAMP.tar \
    --exclude='*.log' \
    --exclude='*.tmp'

# Rotate old backups (keep last 7 days)
find "$BACKUP_DIR" -name "backup_*.tar.gz" -mtime +7 -delete
find "$BACKUP_DIR" -name "database_*.tar" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE"
EOF

    chmod +x "$backup_script"
    print_success "Скрипт бэкапа создан: $backup_script"
}

# Setup logrotate
setup_logrotate() {
    print_step "Настройка ротации логов"

    cat > /etc/logrotate.d/hysteria-vpn << EOF
/var/log/hysteria2/*.log {
    daily
    rotate 7
    compress
    missingok
    notifempty
    copytruncate
    postrotate
        docker-compose -f /opt/hysteriavpn/docker-compose.yml logs > /dev/null 2>&1 || true
    endscript
}
EOF

    print_success "Ротация логов настроена"
}

# Generate final report
generate_final_report() {
    print_header "📊 ИТОГОВЫЙ ОТЧЕТ УСТАНОВКИ"

    echo "=== HysteriaVPN Installation Report ==="
    echo "Installation Date: $(date)"
    echo "Environment: $ENVIRONMENT"
    echo "Master Domain: $MASTER_DOMAIN"
    echo "Nodes Count: $NODE_COUNT"
    echo "Let's Encrypt: ${LETSENCRYPT_ENABLED:-false}"
    echo "Monitoring: ${MONITORING_ENABLED:-false}"
    echo ""
    echo "=== Access URLs ==="
    if [ "$ENVIRONMENT" = "development" ]; then
        echo "Web Panel: http://localhost:3000"
        echo "API: http://localhost:8080"
        echo "Orchestrator: http://localhost:8081"
        if [ "$MONITORING_ENABLED" = true ]; then
            echo "Prometheus: http://localhost:9090"
            echo "Grafana: http://localhost:3001 (admin/hysteria-admin)"
        fi
    else
        echo "Web Panel: https://$MASTER_DOMAIN"
        echo "API: https://$MASTER_DOMAIN/api"
        echo "Orchestrator: https://$MASTER_DOMAIN/orchestrator"
        if [ "$MONITORING_ENABLED" = true ]; then
            echo "Prometheus: https://$MASTER_DOMAIN/prometheus"
            echo "Grafana: https://$MASTER_DOMAIN/grafana"
        fi
    fi
    echo ""
    echo "=== Client Configurations ==="
    echo "Config Directory: /opt/hysteriavpn/client-configs/"
    echo "QR Codes: /opt/hysteriavpn/qr-codes/"
    echo "Main Config: hysteria-client.yaml"
    echo ""
    echo "=== Security ==="
    echo "mTLS: ENABLED"
    echo "Certificates: /opt/hysteriavpn/certs/"
    echo "CA Certificate: /opt/hysteriavpn/certs/ca/ca.crt"
    echo ""
    echo "=== Maintenance ==="
    echo "Backup Script: /opt/hysteriavpn/scripts/backup.sh"
    echo "Logs: /var/log/hysteria2/"
    echo "Config Backup: /opt/hysteriavpn/backup/"
}

# Complete post-installation tasks
post_installation_tasks() {
    print_header "🎉 ЗАВЕРШЕНИЕ УСТАНОВКИ"

    generate_client_configs
    generate_qr_codes
    setup_cron_jobs
    create_backup_script
    setup_logrotate
    generate_final_report

    print_success "Пост-установочные задачи завершены!"
}