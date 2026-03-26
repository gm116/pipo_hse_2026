# Трекер задач

## Сервисы
- `auth-service` (`:8081`) - регистрация, логин, профиль пользователя, выдача JWT.
- `task-service` (`:8082`) - CRUD задач, доступ только к задачам текущего пользователя.
- `gateway` (`:8080`) - единая точка входа, прокси к сервисам, frontend и Swagger.

## Архитектура
- `gateway -> auth-service`
- `gateway -> task-service`
- `auth-service/task-service -> PostgreSQL`

## API-first
Контракт API описан в [`api/openapi.yaml`](api/openapi.yaml). Swagger доступен через gateway:
- `http://localhost:8080/docs`

## Быстрый запуск
### Вариант 1: Docker Compose
```bash
docker compose up --build -d
```
После запуска:
- UI: `http://localhost:8080`
- Swagger: `http://localhost:8080/docs`

### Вариант 2: локально (если `go` в PATH)

1. Поднять PostgreSQL
2. Запустить сервисы в 3 терминалах:
```bash
make run-auth
make run-task
make run-gateway
```

## Тесты и сборка
```bash
make test
make build
```

## Логи и метрики
- Логи: структурированные JSON-логи (`slog`) во всех сервисах.
- Метрики Prometheus:
  - `http://localhost:8080/metrics`
  - `http://localhost:8081/metrics`
  - `http://localhost:8082/metrics`

## Бэкапы PostgreSQL
- В docker-compose `db-backup`.
- Делает `pg_dump` в `./backups`:
  - далее каждые 24 часа (`BACKUP_INTERVAL_SECONDS=86400`);
  - хранение 7 дней (`BACKUP_RETENTION_DAYS=7`).
  - запуск на старте отключен (`BACKUP_RUN_ON_START=false`).

Команды:
```bash
make db-backup-now           # сделать бэкап сразу
make db-backup-list          # список файлов в ./backups
make db-restore-clean BACKUP=backups/task_tracker_YYYYMMDD_HHMMSS.sql.gz 
```

`db-restore-clean` - сама остановит сервисы, очистит схему `public`,
загрузит дамп и поднимет сервисы обратно.

## Миграции и схема
- Миграции: [`internal/db/migrations`](internal/db/migrations)
- При старте сервисов миграции применяются автоматически.
- Версионирование через номера файлов (`001_...sql`).

## Генератора сущностей
```bash
make new-entity NAME=Comment FIELDS=text:string,done:bool TABLE=comments
```
Создаёт:
- SQL-миграцию в `internal/db/migrations`
- Go-модель в `internal/generated`

## По пунктам из задания
- `1.1` Аутентификация и контроль доступа: JWT + доступ к задачам только своего пользователя.
- `1.2` HTTP API с бизнес-методами: создание/получение/изменение/удаление задач.
- `1.3` Тесты: unit + функциональные (`httptest`) для сервисов и gateway.
- `1.4` Внешняя БД: PostgreSQL.
- `1.5` Схема создаётся на старте, миграции версионируются.
- `1.6` Контракт схемы проверяется тестом `TestMigrationSchemaContract`; генератор entity снижает ручную работу при расширении схемы.
- `2.1` Логирование: `slog` middleware и события сервисов.
- `2.2` Метрики: `/metrics` с метриками HTTP-запросов.
- `3.1` CI/CD: GitHub Actions в [`.github/workflows/ci.yml`](.github/workflows/ci.yml).
- `3.2` Swagger-документация: [`api/openapi.yaml`](api/openapi.yaml) + `/docs`.
