.DEFAULT_GOAL := help
.PHONY: help build run test integration vet check generate tools migrate-up migrate-down
help:
	@echo 'build run test integration vet check generate tools migrate-up migrate-down'
build:
	go build -trimpath -o out .
run: build
	./out
test:
	go test -race ./...
integration:
	@test -n "$(TEST_DB_URL)" || (echo 'Set TEST_DB_URL to a migrated disposable database'; exit 1)
	go test -race -run TestAPIIntegration -count=1 ./...
vet:
	go vet ./...
check: vet test
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './.git/*'))"
generate:
	sqlc generate
tools:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
	go install github.com/pressly/goose/v3/cmd/goose@v3.28.0
migrate-up:
	@set -a; if [ -f .env ]; then . ./.env; fi; set +a; goose -dir sql/schema postgres "$$DB_URL" up
migrate-down:
	@set -a; if [ -f .env ]; then . ./.env; fi; set +a; goose -dir sql/schema postgres "$$DB_URL" down
