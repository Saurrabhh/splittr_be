# Load environment-specific .env file if APP_ENV is set and exists, then export variables
ifneq ($(APP_ENV),)
    ifneq (,$(wildcard env/$(APP_ENV)/.env))
        include env/$(APP_ENV)/.env
        export
    endif
endif

.PHONY: local dev prod run build-and-run migrate-status migrate-up migrate-down check-env

# Target to validate that APP_ENV is defined and valid
check-env:
	$(if $(APP_ENV),,$(error APP_ENV is not set. Please specify APP_ENV (e.g., make APP_ENV=local <target>) or run 'make local', 'make dev', or 'make prod'))
	$(if $(filter $(APP_ENV),local dev prod),,$(error invalid APP_ENV '$(APP_ENV)'. Must be one of: local, dev, prod))

# Targets to run the application
local:
	@$(MAKE) APP_ENV=local run

dev:
	@$(MAKE) APP_ENV=dev run

run: check-env
	air

prod:
	@$(MAKE) APP_ENV=prod build-and-run

build-and-run: check-env
	go build -o ./tmp/main.exe ./cmd/api
	./tmp/main.exe

# Targets for Goose database migrations
migrate-status: check-env
	goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status

migrate-up: check-env
	goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

migrate-down: check-env
	goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" down
