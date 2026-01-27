# HysteriaVPN One-Click Installer
# Полнофункциональный установщик для распределенной VPN системы

## 🚀 Быстрый старт

```bash
# Developer установка (localhost, без доменов)
./install-hysteriavpn.sh --dev

# Production установка с доменами
./install-hysteriavpn.sh --domain vpn.yourcompany.com --email admin@company.com --nodes 2

# Интерактивная установка (задаст вопросы)
./install-hysteriavpn.sh
```

## 🎯 Что устанавливает скрипт

### 🔧 Автоматическая установка компонентов:
- **Docker + Docker Compose** (если не установлены)
- **PostgreSQL + Redis** базы данных
- **Orchestrator Service** (центральный сервер управления)
- **API Service** (REST API для веб-интерфейса)
- **Web Dashboard** (React интерфейс управления)
- **Prometheus + Grafana** мониторинг (опционально)
- **VPS Agent сервисы** для nodes (опционально)

### 🔐 Автоматическая настройка безопасности:
- **Let's Encrypt сертификаты** (автоматическое получение и обновление)
- **mTLS** безопасная коммуникация между сервисами
- **Rate limiting** защита от DDoS атак
- **Firewall правила** (UFW/Firewalld)
- **Fail2Ban** (опционально для защиты SSH)

### ⚙️ Функции

#### Развертывание
- **Однокомандная установка**
- **Интерактивный режим** (интеллектуальные вопросы)
- **Production-ready** конфигурации
- **Docker containerization**
- **Автоматические health checks**

#### Безопасность
- **SSL/TLS шифрование** для всех интерфейсов
- **mTLS между сервисами** (не доверяет, проверяет сертификаты)
- **Secure passwords** (автоматическая генерация)
- **Network isolation** (Docker networks)
- **Rate limiting** встроено

#### Мониторинг
- **Prometheus metrics** коллектор
- **Grafana dashboards** (JWT + Node monitoring)
- **AlertManager** уведомления
- **Node Exporter** системные метрики

#### Клиентская интеграция
- **QR коды** для VPN клиентов
- **Config файлы** (.yaml для HysteriaVPN)
- **Connection strings** для мобильных приложений

## 📋 Детальные инструкции

### 🔨 Требования к системе

| Ресурс | Минимум | Рекомендовано |
|--------|---------|---------------|
| RAM | 4GB | 8GB+ |
| CPU | 2 ядра | 4 ядра+ |
| Диск | 20GB | 50GB+ |
| OS | Ubuntu 18+, Debian 10+, CentOS 7+ | Ubuntu 20.04+ |

### 🎚️ Режимы установки

#### 1. Development Mode (`--dev`)
- **Использование**: Для локальной разработки
- **Особенности**:
  - `localhost` вместо доменов
  - Self-signed сертификаты
  - Нет DNS валидации
  - Быстрая установка (<5 мин)

```bash
# Примеры:
./install-hysteriavpn.sh --dev
./install-hysteriavpn.sh --dev --nodes 1 --no-monitoring
```

#### 2. Production Mode (по умолчанию)
- **Использование**: В production окружении
- **Особенности**:
  - Домены обязательны
  - Let's Encrypt интеграция
  - mTLS между сервисами
  - Полная валидация DNS
  - Время установки: 10-15 мин

```bash
# Примеры:
./install-hysteriavpn.sh
./install-hysteriavpn.sh --domain vpn.company.com --email admin@company.com --nodes 3
./install-hysteriavpn.sh --domain vpn.company.com --email admin@company.com --nodes 3 --no-monitoring
```

### 📝 Шаги установки

#### Автоматическая установка:
```bash
# Сделать исполняемым
chmod +x install-hysteriavpn.sh

# Запустить с нужными параметрами
./install-hysteriavpn.sh --domain yourdomain.com --email admin@yourdomain.com --nodes 2
```

#### Что произойдет:
1. **Pre-flight checks** (проверка системы)
2. **Интерактивный ввод** (если параметры не указаны)
3. **Установка зависимостей** (Docker, certbot, etc.)
4. **Генерация сертификатов** (LE + mTLS)
5. **Создание конфигураций** (.env файлы)
6. **Docker развертывание**
7. **Final verification** (health checks)
8. **Completion summary** (URL и пароли)

### 🌐 Доступ после установки

#### Development URLs:
- **Веб-панель**: http://localhost:3000
- **API**: http://localhost:8080
- **Orchestrator**: http://localhost:8081
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001

#### Production URLs:
- **Веб-панель**: https://yourdomain.com
- **API**: https://yourdomain.com/api
- **Orchestrator**: https://yourdomain.com/orchestrator
- **Prometheus**: https://yourdomain.com/prometheus
- **Grafana**: https://yourdomain.com/grafana

### 🔐 Безопасность и учетки

#### По умолчанию:
- **Login**: admin
- **Password**: admin123 (обязательно изменить!)

#### Сертификаты:
- **Let's Encrypt** автоматически обновляются каждые 90 дней
- **mTLS сертификаты** генерируются для внутренних коммуникаций
- **CA сертификаты** хранятся в `/opt/hysteriavpn/certs/ca/`

## 🛠️ Управление после установки

### Docker команды:
```bash
# Просмотр состояния
docker-compose -f docker-compose.generated.yml ps

# Просмотр логов
docker-compose -f docker-compose.generated.yml logs -f [service-name]

# Остановить все
docker-compose -f docker-compose.generated.yml down

# Запустить все
docker-compose -f docker-compose.generated.yml up -d

# Рестарт конкретного сервиса
docker-compose -f docker-compose.generated.yml restart api-service
```

### Бэкапы:
```bash
# Бэкап базы данных
docker exec hysteria2-postgres pg_dump -U hysteria2 hysteria2_db > backup_$(date +%Y%m%d).sql

# Восстановление
cat backup.sql | docker exec -i hysteria2-postgres psql -U hysteria2 -d hysteria2_db
```

### Масштабирование nodes:
```bash
# Запустить дополнительный agent
docker-compose -f docker-compose.generated.yml up -d agent-node-4

# Рассмотреть все agents
docker-compose -f docker-compose.generated.yml scale agent-node=5
```

## 📊 Мониторинг и алерты

### Prometheus метрики:
- **API requests/response times**
- **Database connections**
- **Traffic bandwidth per node**
- **Certificate expiry notifications**
- **System CPU/RAM/Disk usage**

### Grafana dashboards:
- **Service overview** (uptime, health)
- **Traffic analysis** (VPN connections, bandwidth)
- **Node performance** (CPU, memory, disk)
- **API performance** (requests, errors)

### Алерт правила:
- **Service down** notifications
- **High resource usage** alerts
- **Certificate expiry** warnings
- **Failed node connections** alerts

## 🌍 Распределенная архитектура

### Центральный сервер (обязательный):
- **Orchestrator**: Управление nodes через gRPC
- **API Service**: REST API для веб-интерфейса
- **Web Dashboard**: Управление VPN системой
- **PostgreSQL + Redis**: Хранение данных и кэш

### VPS Nodes (опциональные):
- **Agent services**: Агенты на VPS серверах
- **Hysteria2 servers**: VPN endpoints
- **Local monitoring**: Node-specific метрики
- **Geographic distribution**: US/Europe/Asia nodes

## 🔍 Устранение проблем

### Общие проблемы:

#### 1. DNS не просматривается:
```bash
# Проверить DNS
nslookup your-domain.com

# Проверить A запись
dig your-domain.com A
```

#### 2. Сертификаты LE не выдаются:
```bash
# Проверить порт 80 доступен
netstat -tulpn | grep :80

# Тестировать вручную
certbot certonly --standalone -d your-domain.com --dry-run
```

#### 3. Docker не запускается:
```bash
# Проверить статус Docker
systemctl status docker

# Запустить Docker
sudo systemctl start docker
sudo systemctl enable docker
```

#### 4. Сервисы не запускаются:
```bash
# Проверить логи
docker-compose -f docker-compose.generated.yml logs

# Проверить ресурсы
docker system df

# Restart один сервис
docker-compose -f docker-compose.generated.yml restart orchestrator-service
```

## 📞 Техническая поддержка

- **Документация**: https://docs.hysteriavpn.com/installer
- **GitHub Issues**: https://github.com/hysteriavpn/installer/issues
- **Discord**: https://discord.gg/hysteriavpn
- **Email**: support@hysteriavpn.com

## 📄 Лицензия

Copyright (c) 2024 HysteriaVPN Project
Licensed under MIT License - see LICENSE file for details.