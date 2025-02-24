
install-tools:
	@echo installing tools
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/matryer/moq@latest


mock_generate:
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