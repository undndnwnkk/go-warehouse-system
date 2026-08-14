MIGRATIONS_DIR=migrations

.PHONY: generate
generate:
	mkdir -p internal/warehouse/pb
	protoc --proto_path=api/proto \
		--go_out=internal/warehouse/pb --go_opt=paths=source_relative \
		--go-grpc_out=internal/warehouse/pb --go-grpc_opt=paths=source_relative \
		api/proto/warehouse.proto

.PHONY: deps
deps:
	go mod tidy
	go mod download

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(name)" ]; then \
		exit 1; \
	fi
	@mkdir -p $(MIGRATIONS_DIR)
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

.PHONY: migrate-up	
migrate-up:
	migrate -path migrations/ -database "postgres://user:password@localhost:5433/warehouse_db?sslmode=disable" -verbose up

.PHONY: run-auth
run-auth:
	go build -o auth.exe ./internal/auth
	./auth.exe

.PHONY: run-order
run-order:
	go build -o order.exe ./internal/order
	./order.exe

.PHONY: run-warehouse
run-warehouse:
	go build -o warehouse.exe ./internal/warehouse
	./warehouse.exe

.PHONY: run-notification
run-notification:
	go build -o notification.exe ./internal/notification
	./notification.exe

.PHONY: run-swagger
run-swagger:
	go run api/openapi/main.go

.PHONY: test
test:
	go test ./...

.PHONY: up
up:
	docker compose up -d --build

.PHONY: down
down:
	docker compose down

.PHONY: down-v
down-v:
	docker compose down -v

.PHONY: logs
logs:
	docker compose logs -f

.PHONY: ps
ps:
	docker compose ps