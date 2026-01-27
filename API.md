# HysteriaVPN Microservices API Documentation

## Обзор

HysteriaVPN предоставляет REST API для управления VPN инфраструктурой, включая аутентификацию пользователей, управление узлами, мониторинг трафика и WebSocket соединения для реального времени.

**Базовый URL:** `https://your-domain.com/api/v1`

**Формат данных:** JSON

**Аутентификация:** JWT Bearer Token

---

## Аутентификация

Все защищенные эндпоинты требуют JWT токен в заголовке:
```
Authorization: Bearer <your-jwt-token>
```

### Регистрация пользователя

Регистрирует нового пользователя в системе.

**Endpoint:** `POST /api/v1/auth/register`

**Тело запроса:**
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "secure_password123",
  "full_name": "John Doe"
}
```

**Успешный ответ (201):**
```json
{
  "user": {
    "id": "uuid",
    "username": "john_doe",
    "email": "john@example.com",
    "full_name": "John Doe",
    "status": "active",
    "role": "user",
    "data_limit": 0,
    "data_used": 0,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600
  }
}
```

**Ошибки:**
- `400 Bad Request` - Неверные данные
- `409 Conflict` - Пользователь уже существует

### Вход в систему

Аутентифицирует пользователя и возвращает JWT токены.

**Endpoint:** `POST /api/v1/auth/login`

**Тело запроса:**
```json
{
  "username": "john_doe",
  "password": "secure_password123"
}
```

**Успешный ответ (200):**
```json
{
  "user": {
    "id": "uuid",
    "username": "john_doe",
    "email": "john@example.com",
    "role": "user"
  },
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600
  }
}
```

### Обновление токена

Обновляет access токен используя refresh токен.

**Endpoint:** `POST /api/v1/auth/refresh`

**Тело запроса:**
```json
{
  "refresh_token": "your-refresh-token"
}
```

**Успешный ответ (200):**
```json
{
  "access_token": "new-access-token",
  "expires_in": 3600
}
```

---

## Управление пользователями

### Получить всех пользователей

Возвращает список всех пользователей (только администраторы).

**Endpoint:** `GET /api/v1/users`

**Query параметры:**
- `page` (integer, optional) - Номер страницы (по умолчанию: 1)
- `limit` (integer, optional) - Количество элементов на странице (по умолчанию: 50)
- `status` (string, optional) - Фильтр по статусу (active, suspended, deleted)
- `role` (string, optional) - Фильтр по роли (admin, user)

**Успешный ответ (200):**
```json
{
  "users": [
    {
      "id": "uuid",
      "username": "john_doe",
      "email": "john@example.com",
      "full_name": "John Doe",
      "status": "active",
      "role": "user",
      "data_limit": 1073741824,
      "data_used": 524288000,
      "expiry_date": "2024-12-31T23:59:59Z",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-20T15:45:00Z",
      "last_login": "2024-01-20T15:45:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 1,
    "total_pages": 1
  }
}
```

### Создать пользователя

Создает нового пользователя (только администраторы).

**Endpoint:** `POST /api/v1/users`

**Тело запроса:**
```json
{
  "username": "new_user",
  "email": "new@example.com",
  "password": "password123",
  "full_name": "New User",
  "role": "user",
  "data_limit": 1073741824,
  "expiry_date": "2024-12-31T23:59:59Z"
}
```

**Успешный ответ (201):**
```json
{
  "user": {
    "id": "uuid",
    "username": "new_user",
    "email": "new@example.com",
    "full_name": "New User",
    "status": "active",
    "role": "user",
    "data_limit": 1073741824,
    "data_used": 0,
    "expiry_date": "2024-12-31T23:59:59Z",
    "created_at": "2024-01-20T16:00:00Z"
  }
}
```

### Получить пользователя по ID

Возвращает информацию о конкретном пользователе.

**Endpoint:** `GET /api/v1/users/{id}`

**Успешный ответ (200):**
```json
{
  "user": {
    "id": "uuid",
    "username": "john_doe",
    "email": "john@example.com",
    "full_name": "John Doe",
    "status": "active",
    "role": "user",
    "data_limit": 1073741824,
    "data_used": 524288000,
    "expiry_date": "2024-12-31T23:59:59Z",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-20T15:45:00Z",
    "last_login": "2024-01-20T15:45:00Z"
  }
}
```

### Обновить пользователя

Обновляет информацию о пользователе.

**Endpoint:** `PUT /api/v1/users/{id}`

**Тело запроса:**
```json
{
  "full_name": "Updated Name",
  "data_limit": 2147483648,
  "status": "active"
}
```

**Успешный ответ (200):**
```json
{
  "user": {
    "id": "uuid",
    "username": "john_doe",
    "email": "john@example.com",
    "full_name": "Updated Name",
    "status": "active",
    "role": "user",
    "data_limit": 2147483648,
    "data_used": 524288000,
    "updated_at": "2024-01-20T16:30:00Z"
  }
}
```

### Удалить пользователя

Удаляет пользователя (мягкое удаление).

**Endpoint:** `DELETE /api/v1/users/{id}`

**Успешный ответ (204):**
```json
{
  "message": "User deleted successfully"
}
```

### Получить устройства пользователя

Возвращает список устройств пользователя.

**Endpoint:** `GET /api/v1/users/{userId}/devices`

**Успешный ответ (200):**
```json
{
  "devices": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "name": "John's iPhone",
      "device_id": "unique-device-id",
      "public_key": "ed25519-public-key",
      "ip_address": "192.168.1.100",
      "status": "active",
      "data_used": 104857600,
      "created_at": "2024-01-15T10:30:00Z",
      "last_seen": "2024-01-20T16:00:00Z"
    }
  ]
}
```

---

## Управление узлами (Nodes)

### Получить все узлы

Возвращает список всех VPN узлов.

**Endpoint:** `GET /api/v1/nodes`

**Query параметры:**
- `status` (string, optional) - Фильтр по статусу (online, offline, maintenance)
- `location` (string, optional) - Фильтр по локации
- `country` (string, optional) - Фильтр по стране (ISO 3166-1 alpha-2)

**Успешный ответ (200):**
```json
{
  "nodes": [
    {
      "id": "uuid",
      "name": "US-West-1",
      "hostname": "us-west-1.vpn.example.com",
      "ip_address": "1.2.3.4",
      "location": "Los Angeles, CA",
      "country": "US",
      "grpc_port": 50051,
      "status": "online",
      "version": "2.0.1",
      "capabilities": {
        "hysteria2": true,
        "bandwidth": "1Gbps"
      },
      "created_at": "2024-01-10T08:00:00Z",
      "last_heartbeat": "2024-01-20T16:00:00Z"
    }
  ],
  "total": 1
}
```

### Создать узел

Создает новый VPN узел.

**Endpoint:** `POST /api/v1/nodes`

**Тело запроса:**
```json
{
  "name": "EU-Frankfurt-1",
  "hostname": "eu-frankfurt-1.vpn.example.com",
  "ip_address": "5.6.7.8",
  "location": "Frankfurt, Germany",
  "country": "DE",
  "grpc_port": 50051
}
```

**Успешный ответ (201):**
```json
{
  "node": {
    "id": "uuid",
    "name": "EU-Frankfurt-1",
    "hostname": "eu-frankfurt-1.vpn.example.com",
    "ip_address": "5.6.7.8",
    "location": "Frankfurt, Germany",
    "country": "DE",
    "grpc_port": 50051,
    "status": "offline",
    "created_at": "2024-01-20T16:30:00Z"
  }
}
```

### Получить узел по ID

Возвращает информацию о конкретном узле.

**Endpoint:** `GET /api/v1/nodes/{id}`

**Успешный ответ (200):**
```json
{
  "node": {
    "id": "uuid",
    "name": "US-West-1",
    "hostname": "us-west-1.vpn.example.com",
    "ip_address": "1.2.3.4",
    "location": "Los Angeles, CA",
    "country": "US",
    "grpc_port": 50051,
    "status": "online",
    "version": "2.0.1",
    "capabilities": {
      "hysteria2": true,
      "bandwidth": "1Gbps"
    },
    "created_at": "2024-01-10T08:00:00Z",
    "last_heartbeat": "2024-01-20T16:00:00Z",
    "metadata": {
      "cpu_cores": 4,
      "memory_gb": 8
    }
  }
}
```

### Обновить узел

Обновляет информацию об узле.

**Endpoint:** `PUT /api/v1/nodes/{id}`

**Тело запроса:**
```json
{
  "name": "US-West-Updated",
  "status": "maintenance",
  "capabilities": {
    "hysteria2": true,
    "bandwidth": "2Gbps"
  }
}
```

**Успешный ответ (200):**
```json
{
  "node": {
    "id": "uuid",
    "name": "US-West-Updated",
    "hostname": "us-west-1.vpn.example.com",
    "status": "maintenance",
    "capabilities": {
      "hysteria2": true,
      "bandwidth": "2Gbps"
    },
    "updated_at": "2024-01-20T16:45:00Z"
  }
}
```

### Удалить узел

Удаляет узел из системы.

**Endpoint:** `DELETE /api/v1/nodes/{id}`

**Успешный ответ (204):**
```json
{
  "message": "Node deleted successfully"
}
```

### Получить метрики узла

Возвращает метрики производительности узла.

**Endpoint:** `GET /api/v1/nodes/{id}/metrics`

**Query параметры:**
- `limit` (integer, optional) - Количество записей метрик (по умолчанию: 100)

**Успешный ответ (200):**
```json
{
  "metrics": [
    {
      "id": "uuid",
      "node_id": "uuid",
      "cpu_usage": 45.67,
      "memory_usage": 67.89,
      "bandwidth_up": 104857600,
      "bandwidth_down": 209715200,
      "active_connections": 42,
      "recorded_at": "2024-01-20T16:00:00Z"
    }
  ]
}
```

### Перезапустить узел

Отправляет команду перезапуска узлу.

**Endpoint:** `POST /api/v1/nodes/{id}/restart`

**Успешный ответ (200):**
```json
{
  "message": "Node restart initiated",
  "task_id": "uuid"
}
```

### Получить логи узла

Возвращает логи узла.

**Endpoint:** `GET /api/v1/nodes/{id}/logs`

**Query параметры:**
- `lines` (integer, optional) - Количество строк логов (по умолчанию: 100)

**Успешный ответ (200):**
```json
{
  "logs": [
    "2024-01-20 16:00:00 INFO Agent started",
    "2024-01-20 16:00:01 INFO Connected to orchestrator",
    "2024-01-20 16:00:02 INFO Hysteria2 server listening on :443"
  ]
}
```

---

## Статистика трафика

### Получить трафик пользователя

Возвращает статистику трафика для конкретного пользователя.

**Endpoint:** `GET /api/v1/traffic/users/{userId}`

**Query параметры:**
- `from` (string, optional) - Дата начала в формате ISO 8601
- `to` (string, optional) - Дата окончания в формате ISO 8601

**Успешный ответ (200):**
```json
{
  "user_id": "uuid",
  "username": "john_doe",
  "period": {
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-01-31T23:59:59Z"
  },
  "total": {
    "upload": 1073741824,
    "download": 2147483648,
    "total": 3221225472
  },
  "daily_stats": [
    {
      "date": "2024-01-20",
      "upload": 52428800,
      "download": 104857600,
      "total": 157286400
    }
  ],
  "devices": [
    {
      "device_id": "uuid",
      "device_name": "John's iPhone",
      "upload": 536870912,
      "download": 1073741824,
      "total": 1610612736
    }
  ]
}
```

### Получить сводку трафика

Возвращает общую статистику трафика по всем пользователям (только администраторы).

**Endpoint:** `GET /api/v1/traffic/summary`

**Query параметры:**
- `from` (string, optional) - Дата начала в формате ISO 8601
- `to` (string, optional) - Дата окончания в формате ISO 8601

**Успешный ответ (200):**
```json
{
  "total_users": 150,
  "active_users": 89,
  "total_upload": 1099511627776,
  "total_download": 2199023255552,
  "total_data_transfer": 3298534883328,
  "top_users": [
    {
      "user_id": "uuid",
      "username": "power_user",
      "upload": 10737418240,
      "download": 21474836480,
      "total": 32212254720,
      "device_count": 3
    }
  ],
  "top_devices": [
    {
      "device_id": "uuid",
      "user_id": "uuid",
      "device_name": "Gaming PC",
      "username": "gamer123",
      "upload": 5368709120,
      "download": 10737418240,
      "total": 16106127360
    }
  ],
  "from": "2024-01-01T00:00:00Z",
  "to": "2024-01-31T23:59:59Z"
}
```

---

## WebSocket соединения

### Установка WebSocket соединения

Устанавливает WebSocket соединение для получения обновлений в реальном времени.

**Endpoint:** `GET /ws`

**Протокол:** WebSocket

**Аутентификация:** JWT токен в query параметре `token`

**Пример подключения:**
```javascript
const ws = new WebSocket('wss://your-domain.com/ws?token=your-jwt-token');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};
```

**Типы сообщений:**

#### Обновление статуса узла
```json
{
  "type": "node_status_update",
  "data": {
    "node_id": "uuid",
    "status": "online",
    "active_connections": 25,
    "timestamp": "2024-01-20T16:00:00Z"
  }
}
```

#### Новое подключение пользователя
```json
{
  "type": "user_connected",
  "data": {
    "user_id": "uuid",
    "username": "john_doe",
    "device_name": "iPhone",
    "node_id": "uuid",
    "connected_at": "2024-01-20T16:00:00Z"
  }
}
```

#### Обновление трафика
```json
{
  "type": "traffic_update",
  "data": {
    "user_id": "uuid",
    "upload": 1048576,
    "download": 2097152,
    "timestamp": "2024-01-20T16:00:00Z"
  }
}
```

#### Системное уведомление
```json
{
  "type": "system_notification",
  "data": {
    "level": "info",
    "message": "Scheduled maintenance in 30 minutes",
    "timestamp": "2024-01-20T16:00:00Z"
  }
}
```

---

## Системные эндпоинты

### Проверка здоровья системы

Проверяет работоспособность API сервиса.

**Endpoint:** `GET /health`

**Аутентификация:** Не требуется

**Успешный ответ (200):**
```json
{
  "status": "ok",
  "timestamp": "2024-01-20T16:00:00Z",
  "version": "1.0.0"
}
```

---

## Коды ошибок

### HTTP коды состояний
- `200 OK` - Успешный запрос
- `201 Created` - Ресурс успешно создан
- `204 No Content` - Успешный запрос без содержимого
- `400 Bad Request` - Неверные данные запроса
- `401 Unauthorized` - Требуется аутентификация
- `403 Forbidden` - Недостаточно прав доступа
- `404 Not Found` - Ресурс не найден
- `409 Conflict` - Конфликт с существующим ресурсом
- `422 Unprocessable Entity` - Невалидные данные
- `429 Too Many Requests` - Превышен лимит запросов
- `500 Internal Server Error` - Внутренняя ошибка сервера

### Структура ошибок
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid email format",
    "details": {
      "field": "email",
      "value": "invalid-email"
    }
  },
  "timestamp": "2024-01-20T16:00:00Z",
  "request_id": "req-12345"
}
```

### Общие коды ошибок
- `VALIDATION_ERROR` - Ошибка валидации данных
- `AUTHENTICATION_ERROR` - Ошибка аутентификации
- `AUTHORIZATION_ERROR` - Недостаточно прав
- `RESOURCE_NOT_FOUND` - Ресурс не найден
- `RESOURCE_CONFLICT` - Конфликт ресурсов
- `RATE_LIMIT_EXCEEDED` - Превышен лимит запросов
- `INTERNAL_ERROR` - Внутренняя ошибка сервера

---

## SDK и примеры кода

### JavaScript/TypeScript SDK

```javascript
class HysteriaVPNAPI {
  constructor(baseURL, token) {
    this.baseURL = baseURL;
    this.token = token;
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const config = {
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json',
        ...options.headers
      },
      ...options
    };

    const response = await fetch(url, config);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
  }

  // Аутентификация
  async login(credentials) {
    const response = await fetch(`${this.baseURL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credentials)
    });
    return response.json();
  }

  // Пользователи
  async getUsers(params = {}) {
    return this.request('/users', { method: 'GET' });
  }

  async createUser(userData) {
    return this.request('/users', {
      method: 'POST',
      body: JSON.stringify(userData)
    });
  }

  // Узлы
  async getNodes(params = {}) {
    return this.request('/nodes', { method: 'GET' });
  }

  async getNodeMetrics(nodeId, limit = 100) {
    return this.request(`/nodes/${nodeId}/metrics?limit=${limit}`);
  }

  // Трафик
  async getTrafficSummary(from, to) {
    const params = new URLSearchParams();
    if (from) params.append('from', from);
    if (to) params.append('to', to);
    return this.request(`/traffic/summary?${params}`);
  }
}
```

### Python SDK

```python
import requests
from typing import Optional, Dict, Any
import json

class HysteriaVPNAPI:
    def __init__(self, base_url: str, token: Optional[str] = None):
        self.base_url = base_url.rstrip('/')
        self.token = token
        self.session = requests.Session()
        if token:
            self.session.headers.update({
                'Authorization': f'Bearer {token}',
                'Content-Type': 'application/json'
            })

    def _request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        url = f"{self.base_url}/api/v1{endpoint}"
        response = self.session.request(method, url, **kwargs)
        response.raise_for_status()
        return response.json()

    # Authentication
    def login(self, username: str, password: str) -> Dict[str, Any]:
        url = f"{self.base_url}/auth/login"
        response = requests.post(url, json={
            'username': username,
            'password': password
        })
        response.raise_for_status()
        return response.json()

    # Users
    def get_users(self, page: int = 1, limit: int = 50) -> Dict[str, Any]:
        return self._request('GET', f'/users?page={page}&limit={limit}')

    def create_user(self, user_data: Dict[str, Any]) -> Dict[str, Any]:
        return self._request('POST', '/users', json=user_data)

    def get_user(self, user_id: str) -> Dict[str, Any]:
        return self._request('GET', f'/users/{user_id}')

    # Nodes
    def get_nodes(self, status: Optional[str] = None) -> Dict[str, Any]:
        params = {}
        if status:
            params['status'] = status
        return self._request('GET', '/nodes', params=params)

    def get_node(self, node_id: str) -> Dict[str, Any]:
        return self._request('GET', f'/nodes/{node_id}')

    def create_node(self, node_data: Dict[str, Any]) -> Dict[str, Any]:
        return self._request('POST', '/nodes', json=node_data)

    # Traffic
    def get_user_traffic(self, user_id: str, from_date: Optional[str] = None, to_date: Optional[str] = None) -> Dict[str, Any]:
        params = {}
        if from_date:
            params['from'] = from_date
        if to_date:
            params['to'] = to_date
        return self._request('GET', f'/traffic/users/{user_id}', params=params)
```

### C# SDK (для вашего проекта)

```csharp
using System;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Collections.Generic;

namespace HysteriaVPN.Client
{
    public class HysteriaVPNAPI
    {
        private readonly HttpClient _httpClient;
        private readonly string _baseUrl;

        public HysteriaVPNAPI(string baseUrl, string token = null)
        {
            _baseUrl = baseUrl.TrimEnd('/');
            _httpClient = new HttpClient();
            if (!string.IsNullOrEmpty(token))
            {
                _httpClient.DefaultRequestHeaders.Authorization =
                    new AuthenticationHeaderValue("Bearer", token);
            }
        }

        private async Task<T> RequestAsync<T>(HttpMethod method, string endpoint, object data = null)
        {
            var url = $"{_baseUrl}/api/v1{endpoint}";
            var request = new HttpRequestMessage(method, url);

            if (data != null)
            {
                var json = JsonSerializer.Serialize(data);
                request.Content = new StringContent(json, Encoding.UTF8, "application/json");
            }

            var response = await _httpClient.SendAsync(request);
            response.EnsureSuccessStatusCode();

            var content = await response.Content.ReadAsStringAsync();
            return JsonSerializer.Deserialize<T>(content, new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            });
        }

        // Authentication
        public async Task<LoginResponse> LoginAsync(string username, string password)
        {
            var url = $"{_baseUrl}/auth/login";
            var request = new HttpRequestMessage(HttpMethod.Post, url)
            {
                Content = new StringContent(JsonSerializer.Serialize(new
                {
                    username,
                    password
                }), Encoding.UTF8, "application/json")
            };

            var response = await _httpClient.SendAsync(request);
            response.EnsureSuccessStatusCode();

            var content = await response.Content.ReadAsStringAsync();
            return JsonSerializer.Deserialize<LoginResponse>(content, new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            });
        }

        // Users
        public async Task<UserListResponse> GetUsersAsync(int page = 1, int limit = 50)
        {
            return await RequestAsync<UserListResponse>(HttpMethod.Get, $"/users?page={page}&limit={limit}");
        }

        public async Task<UserResponse> CreateUserAsync(CreateUserRequest request)
        {
            return await RequestAsync<UserResponse>(HttpMethod.Post, "/users", request);
        }

        public async Task<UserResponse> GetUserAsync(string userId)
        {
            return await RequestAsync<UserResponse>(HttpMethod.Get, $"/users/{userId}");
        }

        public async Task<UserResponse> UpdateUserAsync(string userId, UpdateUserRequest request)
        {
            return await RequestAsync<UserResponse>(HttpMethod.Put, $"/users/{userId}", request);
        }

        // Nodes
        public async Task<NodeListResponse> GetNodesAsync(string status = null, string location = null)
        {
            var query = new List<string>();
            if (!string.IsNullOrEmpty(status)) query.Add($"status={status}");
            if (!string.IsNullOrEmpty(location)) query.Add($"location={location}");

            var queryString = query.Count > 0 ? "?" + string.Join("&", query) : "";
            return await RequestAsync<NodeListResponse>(HttpMethod.Get, $"/nodes{queryString}");
        }

        public async Task<NodeResponse> GetNodeAsync(string nodeId)
        {
            return await RequestAsync<NodeResponse>(HttpMethod.Get, $"/nodes/{nodeId}");
        }

        public async Task<NodeResponse> CreateNodeAsync(CreateNodeRequest request)
        {
            return await RequestAsync<NodeResponse>(HttpMethod.Post, "/nodes", request);
        }

        public async Task<NodeMetricsResponse> GetNodeMetricsAsync(string nodeId, int limit = 100)
        {
            return await RequestAsync<NodeMetricsResponse>(HttpMethod.Get, $"/nodes/{nodeId}/metrics?limit={limit}");
        }

        // Traffic
        public async Task<UserTrafficResponse> GetUserTrafficAsync(string userId, DateTime? from = null, DateTime? to = null)
        {
            var query = new List<string>();
            if (from.HasValue) query.Add($"from={from.Value.ToString("O")}");
            if (to.HasValue) query.Add($"to={to.Value.ToString("O")}");

            var queryString = query.Count > 0 ? "?" + string.Join("&", query) : "";
            return await RequestAsync<UserTrafficResponse>(HttpMethod.Get, $"/traffic/users/{userId}{queryString}");
        }

        public async Task<TrafficSummaryResponse> GetTrafficSummaryAsync(DateTime? from = null, DateTime? to = null)
        {
            var query = new List<string>();
            if (from.HasValue) query.Add($"from={from.Value.ToString("O")}");
            if (to.HasValue) query.Add($"to={to.Value.ToString("O")}");

            var queryString = query.Count > 0 ? "?" + string.Join("&", query) : "";
            return await RequestAsync<TrafficSummaryResponse>(HttpMethod.Get, $"/traffic/summary{queryString}");
        }
    }

    // Response models
    public class LoginResponse
    {
        public User User { get; set; }
        public TokenPair Tokens { get; set; }
    }

    public class TokenPair
    {
        public string AccessToken { get; set; }
        public string RefreshToken { get; set; }
        public int ExpiresIn { get; set; }
    }

    public class User
    {
        public string Id { get; set; }
        public string Username { get; set; }
        public string Email { get; set; }
        public string FullName { get; set; }
        public string Status { get; set; }
        public string Role { get; set; }
        public long DataLimit { get; set; }
        public long DataUsed { get; set; }
        public DateTime? ExpiryDate { get; set; }
        public DateTime CreatedAt { get; set; }
        public DateTime UpdatedAt { get; set; }
        public DateTime? LastLogin { get; set; }
    }

    public class VPSNode
    {
        public string Id { get; set; }
        public string Name { get; set; }
        public string Hostname { get; set; }
        public string IpAddress { get; set; }
        public string Location { get; set; }
        public string Country { get; set; }
        public int GrpcPort { get; set; }
        public string Status { get; set; }
        public string Version { get; set; }
        public Dictionary<string, object> Capabilities { get; set; }
        public DateTime CreatedAt { get; set; }
        public DateTime? LastHeartbeat { get; set; }
    }

    // Request models
    public class CreateUserRequest
    {
        public string Username { get; set; }
        public string Email { get; set; }
        public string Password { get; set; }
        public string FullName { get; set; }
        public string Role { get; set; }
        public long? DataLimit { get; set; }
        public DateTime? ExpiryDate { get; set; }
    }

    public class UpdateUserRequest
    {
        public string FullName { get; set; }
        public string Status { get; set; }
        public long? DataLimit { get; set; }
        public DateTime? ExpiryDate { get; set; }
    }

    public class CreateNodeRequest
    {
        public string Name { get; set; }
        public string Hostname { get; set; }
        public string IpAddress { get; set; }
        public string Location { get; set; }
        public string Country { get; set; }
        public int? GrpcPort { get; set; }
    }

    // Response models
    public class UserListResponse
    {
        public List<User> Users { get; set; }
        public PaginationInfo Pagination { get; set; }
    }

    public class PaginationInfo
    {
        public int Page { get; set; }
        public int Limit { get; set; }
        public int Total { get; set; }
        public int TotalPages { get; set; }
    }

    public class UserResponse
    {
        public User User { get; set; }
    }

    public class NodeListResponse
    {
        public List<VPSNode> Nodes { get; set; }
        public int Total { get; set; }
    }

    public class NodeResponse
    {
        public VPSNode Node { get; set; }
    }

    public class NodeMetricsResponse
    {
        public List<NodeMetric> Metrics { get; set; }
    }

    public class NodeMetric
    {
        public string Id { get; set; }
        public string NodeId { get; set; }
        public double CpuUsage { get; set; }
        public double MemoryUsage { get; set; }
        public long BandwidthUp { get; set; }
        public long BandwidthDown { get; set; }
        public int ActiveConnections { get; set; }
        public DateTime RecordedAt { get; set; }
    }

    public class UserTrafficResponse
    {
        public string UserId { get; set; }
        public string Username { get; set; }
        public TrafficPeriod Period { get; set; }
        public TrafficData Total { get; set; }
        public List<DailyTraffic> DailyStats { get; set; }
        public List<DeviceTraffic> Devices { get; set; }
    }

    public class TrafficPeriod
    {
        public DateTime From { get; set; }
        public DateTime To { get; set; }
    }

    public class TrafficData
    {
        public long Upload { get; set; }
        public long Download { get; set; }
        public long Total { get; set; }
    }

    public class DailyTraffic
    {
        public string Date { get; set; }
        public long Upload { get; set; }
        public long Download { get; set; }
        public long Total { get; set; }
    }

    public class DeviceTraffic
    {
        public string DeviceId { get; set; }
        public string DeviceName { get; set; }
        public long Upload { get; set; }
        public long Download { get; set; }
        public long Total { get; set; }
    }

    public class TrafficSummaryResponse
    {
        public int TotalUsers { get; set; }
        public int ActiveUsers { get; set; }
        public long TotalUpload { get; set; }
        public long TotalDownload { get; set; }
        public long TotalDataTransfer { get; set; }
        public List<UserTrafficRank> TopUsers { get; set; }
        public List<DeviceTrafficRank> TopDevices { get; set; }
        public DateTime From { get; set; }
        public DateTime To { get; set; }
    }

    public class UserTrafficRank
    {
        public string UserId { get; set; }
        public string Username { get; set; }
        public long Upload { get; set; }
        public long Download { get; set; }
        public long Total { get; set; }
        public int DeviceCount { get; set; }
    }

    public class DeviceTrafficRank
    {
        public string DeviceId { get; set; }
        public string UserId { get; set; }
        public string DeviceName { get; set; }
        public string Username { get; set; }
        public long Upload { get; set; }
        public long Download { get; set; }
        public long Total { get; set; }
    }
}
```

---

## Версионирование API

API использует семантическое версионирование:

- **MAJOR.MINOR.PATCH**
- MAJOR - breaking changes
- MINOR - новые функции (backward compatible)
- PATCH - bug fixes

**Текущая версия:** v1

**Заголовок версии:** `Accept: application/vnd.hysteriavpn.v1+json`

---

## Ограничения и квоты

### Rate Limiting
- **Аутентифицированные запросы:** 1000 запросов в час
- **Неаутентифицированные запросы:** 100 запросов в час
- **WebSocket соединения:** 10 одновременных соединений на пользователя

### Размер данных
- **Максимальный размер запроса:** 10MB
- **Максимальный размер ответа:** 50MB
- **Максимальное количество устройств на пользователя:** 10

### Время ожидания
- **Таймаут запроса:** 30 секунд
- **WebSocket ping/pong:** 60 секунд

---

## Поддержка и обратная связь

### Контакты
- **Email:** support@hysteriavpn.com
- **Документация:** https://docs.hysteriavpn.com
- **GitHub Issues:** https://github.com/hysteriavpn/api/issues

### Сообщество
- **Discord:** https://discord.gg/hysteriavpn
- **Telegram:** @hysteriavpn_support

---

*Последнее обновление: 20 января 2024 г. | Версия API: v1.0.0*</content>
<parameter name="filePath">API.md