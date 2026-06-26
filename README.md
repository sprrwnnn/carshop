# Carshop Backend

Carshop — это PET-проект на Go, созданный для демонстрации продакшен-подхода в разработке бэкенда и DevOps практик: контейнеризация сервисов, миграции баз данных, интеграция с кешем, асинхронные события, метрики, алерты и CI.

## Что демонстрирует

- REST API на Go с чистыми внутренними границами пакетов
- CQRS-подход: разделение на команды и запросы с проекциями в PostgreSQL
- Публикация событий и отдельный воркер для уведомлений через RabbitMQ
- Кеширование запросов на чтение через KeyDB
- Окружение Docker Compose: API, воркер, Postgres, KeyDB, RabbitMQ, Prometheus, Grafana, Alertmanager и Mailpit
- Swagger документация, эндпоинт здоровья, метрики Prometheus и pprof в локальном режиме
- CI через GitLab CI и GitHub Actions: проверки, тесты и сборка Docker-образов
- Ansible deploy для Linux-хоста с Docker Compose
- Операционные Bash/Python-скрипты: smoke test, healthcheck стека, backup/restore PostgreSQL

## Архитектура

```text
Клиент
  |
  v
Go API ---> KeyDB (кеш)
  |
  v
Postgres: модель записи + проекции чтения
  |
  v
RabbitMQ exchange ---> воркер уведомлений

Prometheus собирает метрики API и RabbitMQ
Grafana визуализирует метрики
Alertmanager управляет правилами алертов
```

## Технологический стек

- Language: Go
- HTTP router: chi
- Database: PostgreSQL
- Cache: KeyDB
- Messaging: RabbitMQ
- Observability: Prometheus, Grafana, Alertmanager, pprof
- Packaging: Docker, Docker Compose
- CI/CD: GitLab CI, GitHub Actions
- Automation: Make, Bash, Python, Ansible

## Quick Start

Создайте локальные конфигурационные файлы из примеров:

```bash
cp .build/config/local.example.yaml .build/config/local.yaml
cp .build/config/notification.example.yaml .build/config/notification.yaml
cp infrastructure/database/.env.example infrastructure/database/.env
cp infrastructure/redis/.env.example infrastructure/redis/.env
cp infrastructure/rabbitmq/.env.example infrastructure/rabbitmq/.env
cp infrastructure/grafana/.env.example infrastructure/grafana/.env
```

Используйте следующие значения для первого локального запуска:

```env
# infrastructure/database/.env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=carshop

# infrastructure/redis/.env
REDIS_PASSWORD=password

# infrastructure/rabbitmq/.env
RABBITMQ_DEFAULT_USER=guest
RABBITMQ_DEFAULT_PASS=guest
RABBITMQ_SERVER_ADDITIONAL_ERL_ARGS="-rabbitmq_prometheus true"

# infrastructure/grafana/.env
GF_SECURITY_ADMIN_PASSWORD=admin
```

Запустите полный стек:

```bash
make compose-up
```

Полезные URLs:

- API: http://localhost:8000
- Swagger: http://localhost:8000/swagger/index.html
- Health: http://localhost:8000/api/v1/healthcheck/
- Metrics: http://localhost:8000/api/v1/metrics
- RabbitMQ UI: http://localhost:15672
- Prometheus: http://localhost:9100
- Grafana: http://localhost:3000
- Alertmanager: http://localhost:9093
- Mailpit: http://localhost:8025

Остановка всех сервисов:

```bash
make compose-down
```

## API Examples

Создать автомобиль:

```bash
curl -X POST http://localhost:8000/api/v1/cars/c/ \
  -H 'Content-Type: application/json' \
  -d '{"name":"BMW M3","colour":"#1122AA","price":75000,"build_date":"2024-05-01"}'
```

Получение списка автомобилей:

```bash
curl http://localhost:8000/api/v1/cars/q/
```

Фильтр автомобилей по цене:

```bash
curl 'http://localhost:8000/api/v1/cars/q/?price_from=10000&price_to=80000'
```

## Development

Запуск тестов:

```bash
make test
```

Запуск покрытия кода тестами:

```bash
make test-coverage
```

Сборка локальных бинарных файлов:

```bash
make build
```

Сборка Docker-образа:

```bash
make docker-build
```

## Operational Scripts

Проверить, что все сервисы Docker Compose запущены:

```bash
scripts/check_stack.sh
```

Запустить smoke test API, healthcheck и метрик:

```bash
scripts/smoke_test.py
```

Создать backup PostgreSQL в `backups/postgres`:

```bash
scripts/backup_postgres.sh
```

Восстановить PostgreSQL из backup:

```bash
scripts/restore_postgres.sh backups/postgres/carshop_YYYYMMDDTHHMMSSZ.dump
```

Переменные окружения для скриптов:

```bash
COMPOSE_PROJECT_NAME=carshop
COMPOSE_FILE=docker-compose.yaml
POSTGRES_USER=postgres
POSTGRES_DB=carshop
BASE_URL=http://localhost:8000
```

## GitLab CI/CD

Pipeline описан в `.gitlab-ci.yml` и включает стадии:

- `validate`: `go mod verify`, `gofmt`, `go vet`
- `test`: race-тесты и coverage artifact
- `build`: сборка бинарников API и notification worker
- `docker`: сборка Docker-образов и push в GitLab Container Registry для default branch/tag
- `deploy`: ручной Ansible deploy

Для deploy job подготовьте инвентарь:

```bash
cp ansible/inventory.example.ini ansible/inventory.ini
```

Затем укажите реальный `ansible_host`, пользователя и SSH-доступ к Linux-хосту.

## Ansible Deploy

Ansible playbook находится в `ansible/playbooks/deploy.yml`. Он копирует проект на Linux-хост, рендерит конфигурации и `.env` файлы, затем запускает стек через Docker Compose.

Локальная проверка синтаксиса:

```bash
ansible-playbook -i ansible/inventory.example.ini ansible/playbooks/deploy.yml --syntax-check
```

Запуск деплоя:

```bash
ansible-playbook -i ansible/inventory.ini ansible/playbooks/deploy.yml
```

Основные переменные лежат в `ansible/group_vars/carshop.yml`. Пароли из файла подходят только для локального стенда; для реального окружения лучше задавать их через Ansible Vault или GitLab CI/CD variables.

