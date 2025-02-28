
install-tools:
	@echo installing tools
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install go.uber.org/mock/mockgen@latest


run:
	@echo "Running the application using docker-compose in production mode"
	@docker compose --env-file .env.production up -d --build

run-dev:
	@echo "Running the application in development mode"
	docker compose up -d db nats --wait
	@go run cmd/main.go

test:
	@go test -v ./...


mockgen:
	@go generate ./... 

migration-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: Please provide a migration name using 'make migration-create name=<migration_name>'"; \
		exit 1; \
	fi
	goose -s -dir ./migrations create $(name) sql

migration-up:
	@go run migrations/migration.go up

migration-down:
	@go run migrations/migration.go down