import React from 'react';

const Documentation = () => {
  return (
    <div className="documentation">
      <h1>HysteryVPN Orchestrator Documentation</h1>

      <section>
        <h2>Быстрый старт</h2>
        <h3>Установка оркестратора</h3>
        <pre>
{`# На сервере оркестратора
wget https://raw.githubusercontent.com/your-repo/HysteryVPN/main/hysteria2-microservices/deployments/docker/install-orchestrator.sh
chmod +x install-orchestrator.sh
./install-orchestrator.sh`}
        </pre>

        <h3>Установка узла</h3>
        <pre>
{`# На VPS узла
wget https://raw.githubusercontent.com/your-repo/HysteryVPN/main/hysteria2-microservices/deployments/docker/install-node.sh
chmod +x install-node.sh
./install-node.sh`}
        </pre>
      </section>

      <section>
        <h2>Управление узлами</h2>
        <h3>Просмотр узлов</h3>
        <p>На главной странице отображается список всех зарегистрированных узлов с их статусом, локацией и нагрузкой.</p>

        <h3>Настройка конфигурации узла</h3>
        <ol>
          <li>Выберите узел из списка</li>
          <li>Перейдите на вкладку "Configuration"</li>
          <li>Заполните параметры Hysteria2:
            <ul>
              <li><strong>Listen Port:</strong> UDP порт для VPN соединений (по умолчанию 8080)</li>
              <li><strong>Obfs:</strong> Метод обфускации (salamander)</li>
              <li><strong>Obfs Password:</strong> Пароль для обфускации</li>
              <li><strong>Auth:</strong> Метод аутентификации</li>
              <li><strong>Auth Password:</strong> Пароль аутентификации</li>
            </ul>
          </li>
          <li>Нажмите "Save Configuration"</li>
        </ol>

        <h3>Мониторинг узлов</h3>
        <p>Для каждого узла доступна информация:</p>
        <ul>
          <li>Статус соединения (Online/Offline)</li>
          <li>Загрузка CPU и RAM</li>
          <li>Сетевой трафик (Upload/Download)</li>
          <li>Количество активных пользователей</li>
        </ul>
      </section>

      <section>
        <h2>Управление пользователями</h2>
        <h3>Создание пользователя</h3>
        <ol>
          <li>Перейдите в раздел "Users"</li>
          <li>Нажмите "Add User"</li>
          <li>Заполните:
            <ul>
              <li>Username</li>
              <li>Email</li>
              <li>Password</li>
              <li>Assign to Node (выберите узел)</li>
            </ul>
          </li>
          <li>Нажмите "Create"</li>
        </ol>

        <h3>Генерация VPN конфигурации</h3>
        <p>После создания пользователя:</p>
        <ol>
          <li>Выберите пользователя</li>
          <li>Нажмите "Generate Config"</li>
          <li>Скопируйте hy2:// URI</li>
          <li>Импортируйте в HysteryVPN клиент</li>
        </ol>
      </section>

      <section>
        <h2>Подключение клиентов</h2>
        <h3>HysteryVPN WPF Client</h3>
        <ol>
          <li>Запустите HysteryVPN.exe</li>
          <li>Вставьте hy2:// ссылку от пользователя</li>
          <li>Нажмите "Connect"</li>
        </ol>

        <h3>Параметры URI</h3>
        <pre>
{`hy2://server:port?obfs=salamander&obfs-password=password&auth=password&auth-password=userpass`}
        </pre>
      </section>

      <section>
        <h2>Безопасность</h2>
        <ul>
          <li>Используйте сложные пароли для obfs и auth</li>
          <li>Регулярно меняйте JWT секреты</li>
          <li>Ограничьте доступ к оркестратору файрволом</li>
          <li>Включайте mTLS для gRPC соединений</li>
        </ul>
      </section>

      <section>
        <h2>Устранение неполадок</h2>
        <h3>Узел не регистрируется</h3>
        <ul>
          <li>Проверьте MASTER_SERVER в .env агента</li>
          <li>Убедитесь, что порты 50051 открыты</li>
          <li>Проверьте логи: <code>docker logs hysteria-agent-node-id</code></li>
        </ul>

        <h3>Клиент не подключается</h3>
        <ul>
          <li>Проверьте obfs параметры на клиенте и сервере</li>
          <li>Убедитесь, что UDP порт открыт на узле</li>
          <li>Проверьте логи Hysteria2 на узле</li>
        </ul>

        <h3>Команды для диагностики</h3>
        <pre>
{`# Логи оркестратора
docker-compose logs -f orchestrator-service

# Логи узла
docker logs -f hysteria-agent-node-id

# Проверка портов
netstat -tlnp | grep :8080
netstat -tlnp | grep :50051

# Тест соединения
curl http://localhost:8081/api/v1/nodes`}
        </pre>
      </section>

      <section>
        <h2>API Reference</h2>
        <h3>Nodes</h3>
        <ul>
          <li><code>GET /api/v1/nodes</code> - список узлов</li>
          <li><code>GET /api/v1/nodes/{id}</code> - детали узла</li>
          <li><code>PUT /api/v1/nodes/{id}</code> - обновить узел</li>
          <li><code>GET /api/v1/nodes/{id}/config</code> - получить конфигурацию</li>
          <li><code>PUT /api/v1/nodes/{id}/config</code> - обновить конфигурацию</li>
        </ul>

        <h3>Users</h3>
        <ul>
          <li><code>GET /api/v1/users</code> - список пользователей</li>
          <li><code>POST /api/v1/users</code> - создать пользователя</li>
          <li><code>PUT /api/v1/users/{id}</code> - обновить пользователя</li>
          <li><code>DELETE /api/v1/users/{id}</code> - удалить пользователя</li>
        </ul>
      </section>
    </div>
  );
};

export default Documentation;