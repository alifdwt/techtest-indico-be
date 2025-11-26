swag:
	swag init -g cmd/api/main.go -o docs/

run:
	go run cmd/api/main.go

.PHONY: swag run