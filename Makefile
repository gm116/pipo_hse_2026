.PHONY: test build run-auth run-task run-gateway dev-up dev-down docker-up docker-down schema-check

test:
	go test ./...

build:
	go build ./cmd/auth-service
	go build ./cmd/task-service
	go build ./cmd/gateway

run-auth:
	APP_NAME=auth-service PORT=8081 go run ./cmd/auth-service

run-task:
	APP_NAME=task-service PORT=8082 go run ./cmd/task-service

run-gateway:
	APP_NAME=gateway PORT=8080 AUTH_SERVICE_URL=http://localhost:8081 TASK_SERVICE_URL=http://localhost:8082 go run ./cmd/gateway

docker-up:
	docker compose up --build -d
	@echo ""
	@echo "Services are up:"
	@echo "  Gateway:      http://localhost:8080"
	@echo "  Swagger:      http://localhost:8080/docs"
	@echo "  Auth API:     http://localhost:8081"
	@echo "  Task API:     http://localhost:8082"
	@echo "  Postgres:     localhost:5432"
	@echo ""
	@echo "Metrics:"
	@echo "  Gateway:      http://localhost:8080/metrics"
	@echo "  Auth:         http://localhost:8081/metrics"
	@echo "  Tasks:        http://localhost:8082/metrics"
	@echo "  Auth (proxy): http://localhost:8080/api/auth/metrics"
	@echo "  Task (proxy): http://localhost:8080/api/tasks/metrics"

docker-down:
	docker compose down -v

schema-check:
	go test ./internal/db -run TestMigrationSchemaContract

new-entity:
	@echo "Usage: make new-entity NAME=Comment FIELDS=text:string,done:bool TABLE=comments"
	go run ./cmd/entitygen --name $(NAME) --fields $(FIELDS) $(if $(TABLE),--table $(TABLE),)
