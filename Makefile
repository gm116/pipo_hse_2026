.PHONY: test build run-auth run-task run-gateway dev-up dev-down docker-up docker-down schema-check sqlc-generate db-backup-now db-backup-list db-restore db-restore-clean

test:
	go test ./...

build:
	go build ./cmd/auth-service
	go build ./cmd/task-service
	go build ./cmd/gateway

run-auth:
	@set -a; [ -f .env ] && . ./.env; set +a; APP_NAME=auth-service PORT=8081 go run ./cmd/auth-service

run-task:
	@set -a; [ -f .env ] && . ./.env; set +a; APP_NAME=task-service PORT=8082 go run ./cmd/task-service

run-gateway:
	@set -a; [ -f .env ] && . ./.env; set +a; APP_NAME=gateway PORT=8080 AUTH_SERVICE_URL=$${AUTH_SERVICE_URL:-http://localhost:8081} TASK_SERVICE_URL=$${TASK_SERVICE_URL:-http://localhost:8082} go run ./cmd/gateway

docker-up:
	docker compose up --build -d
	@DOMAIN=$$(grep -E '^PUBLIC_DOMAIN=' .env 2>/dev/null | cut -d= -f2-); [ -n "$$DOMAIN" ] || DOMAIN=:80; \
	echo ""; \
	echo "Public entrypoint (Caddy):"; \
	if [ "$$DOMAIN" = ":80" ]; then echo "  http://localhost"; else echo "  https://$$DOMAIN"; fi
	@echo ""
	@echo "Services are up:"
	@echo "  Gateway:      http://localhost:8080"
	@echo "  Swagger:      http://localhost:8080/docs"
	@echo "  Auth API:     http://localhost:8081"
	@echo "  Task API:     http://localhost:8082"
	@echo "  Postgres:     localhost:5432"
	@echo "  Backups dir:  ./backups"
	@echo ""
	@echo "Metrics:"
	@echo "  Gateway:      http://localhost:8080/metrics"
	@echo "  Auth:         http://localhost:8081/metrics"
	@echo "  Tasks:        http://localhost:8082/metrics"
	@echo "  Auth (proxy): http://localhost:8080/api/auth/metrics"
	@echo "  Task (proxy): http://localhost:8080/api/tasks/metrics"

docker-down:
	docker compose down -v

db-backup-now:
	docker compose up -d postgres
	docker compose run --rm -T --entrypoint /scripts/db_backup.sh db-backup

db-backup-list:
	ls -lh backups

db-restore:
	@test -n "$(BACKUP)" || (echo "Usage: make db-restore BACKUP=backups/task_tracker_YYYYMMDD_HHMMSS.sql.gz"; exit 1)
	gzip -dc "$(BACKUP)" | docker compose exec -T postgres psql -U postgres -d task_tracker

db-restore-clean:
	@test -n "$(BACKUP)" || (echo "Usage: make db-restore-clean BACKUP=backups/task_tracker_YYYYMMDD_HHMMSS.sql.gz"; exit 1)
	docker compose up -d postgres
	docker compose stop auth-service task-service gateway db-backup || true
	docker compose exec -T postgres psql -U postgres -d task_tracker -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	gzip -dc "$(BACKUP)" | docker compose exec -T postgres psql -U postgres -d task_tracker
	docker compose up -d db-backup auth-service task-service gateway
	@echo "Restore completed from $(BACKUP)"

schema-check:
	go test ./internal/db -run TestMigrationSchemaContract

sqlc-generate:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

new-entity:
	@echo "Usage: make new-entity NAME=Comment FIELDS=text:string,done:bool TABLE=comments"
	go run ./cmd/entitygen --name $(NAME) --fields $(FIELDS) $(if $(TABLE),--table $(TABLE),)
