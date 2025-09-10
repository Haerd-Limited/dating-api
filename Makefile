deps:
	@echo "Installing dependencies..."
	go install github.com/daixiang0/gci@latest
	go mod tidy
	go mod download
.PHONY: deps

DSN="postgres://haerdappuser:H@erd_2025@localhost:5432/haerd-dating-db?sslmode=disable"

migrate-up:
	@goose -dir ./migrations postgres ${DSN} up
.PHONY: migrate-up

migrate-down:
	@goose -dir ./migrations postgres ${DSN} down
.PHONY: migrate-down

migrate-create:
	@cd ./migrations && goose create  create_swipes_table sql
.PHONY: migrate-create


mock:
	go generate ./...
.PHONY: mock

entity:
	@sqlboiler psql -c ./sqlboiler.toml
.PHONY: entity

build:
	docker-compose up --build -d
.PHONY: build

start:
	docker-compose up -d
.PHONY: start

stop:
	docker-compose down -v
.PHONY: stop

test:
	go test -count=1 -p 1 ./...
.PHONY: test

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
.PHONY: cover

lint:
	go mod tidy
	go vet ./...
	gci write -s standard -s default -s "prefix(github.com/Haerd-Limited/dating-api)" .
	gofumpt -l -w .
	wsl -fix ./... 2> /dev/null || true
	golangci-lint run $(p)
	go fmt ./...
.PHONY: lint

it:
	go test -count=1 -p 1 ./internal/integrationtests/...
.PHONY: it

itc:
	go test -coverprofile=integration-coverage.out ./...
	go tool cover -html=integration-coverage.out
	go tool cover -func=integration-coverage.out
.PHONY: itc
