swag:
	swag init -g cmd/api/main.go -o docs/

run:
	go run cmd/api/main.go

# Database migrations
migrate-up:
	migrate -path db/migration -database "postgres://postgres:postgres@localhost:2050/techtest_indico?sslmode=disable" up

migrate-down:
	migrate -path db/migration -database "postgres://postgres:postgres@localhost:2050/techtest_indico?sslmode=disable" down

sqlc:
	sqlc generate

.PHONY: swag run migrate-up migrate-down sqlc