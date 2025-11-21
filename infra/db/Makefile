DB_URL ?= postgres://user:password@localhost:5432/portfolio?sslmode=disable

.PHONY: migrate-up migrate-down migrate-force migrate-create

migrate-up:
	migrate -path . -database "$(DB_URL)" up

migrate-down:
	migrate -path . -database "$(DB_URL)" down

migrate-force:
	migrate -path . -database "$(DB_URL)" force $(VERSION)

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir . -seq $$name
