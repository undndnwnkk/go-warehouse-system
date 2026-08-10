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

.PHONY: migrate-up
migrate-up:
	migrate -path migrations/warehouse -database "postgres://user:password@localhost:5433/warehouse_db?sslmode=disable" -verbose up