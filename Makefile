DB_MIGRATE_URL ?= postgres://default:1234@localhost:5446/withdrawal?sslmode=disable
MIGRATE_PATH ?= ./migration/postgres

run:
	go run ./cmd/app/main.go

run-ui:
	go run ./cmd/ui/main.go

up:
	docker compose up --build --force-recreate

down:
	docker compose down

.PHONY: test
test:
	go test -v -cover ./...

integration-test:
	go test -count=1 -v -tags=integration ./test/integration

migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

migrate-create:
ifndef NAME
	$(error Usage: make migrate-create NAME=init_schema)
endif
	migrate create -ext sql -dir "$(MIGRATE_PATH)" -seq $(NAME)

migrate-up:
	migrate -database "$(DB_MIGRATE_URL)" -path "$(MIGRATE_PATH)" up

migrate-up-one:
	migrate -database "$(DB_MIGRATE_URL)" -path "$(MIGRATE_PATH)" up 1

migrate-down:
	migrate -database "$(DB_MIGRATE_URL)" -path "$(MIGRATE_PATH)" down -all

migrate-down-one:
	migrate -database "$(DB_MIGRATE_URL)" -path "$(MIGRATE_PATH)" down 1

migrate-version:
	migrate -database "$(DB_MIGRATE_URL)" -path "$(MIGRATE_PATH)" version

migrate-force:
ifndef VERSION
	$(error Usage: make migrate-force VERSION=1)
endif
	migrate -database "$(DB_MIGRATE_URL)" -path "$(MIGRATE_PATH)" force $(VERSION)

generate:
	go generate ./...

mockery-install:
	go install github.com/vektra/mockery/v3@v3.2.5
