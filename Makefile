# Support positional environment argument for goose commands (e.g., make goose-up local)
SUPPORTED_GOOSE_TARGETS := goose-status goose-up goose-down
FIRST_GOAL := $(firstword $(MAKECMDGOALS))

ifeq ($(filter $(FIRST_GOAL),$(SUPPORTED_GOOSE_TARGETS)),$(FIRST_GOAL))
    # Extract the second argument as the environment
    ENV_ARG := $(word 2,$(MAKECMDGOALS))
    ifneq ($(ENV_ARG),)
        APP_ENV := $(ENV_ARG)
        # Define a no-op target for the environment name to prevent make errors or execution of run targets
        $(eval $(ENV_ARG):;@:)
    endif
endif

# Map env parameter to APP_ENV if provided
ifneq ($(env),)
    APP_ENV := $(env)
endif

# Load environment-specific .env file if APP_ENV is set and exists, then export variables
ifneq ($(APP_ENV),)
    ifneq (,$(wildcard env/$(APP_ENV)/.env))
        include env/$(APP_ENV)/.env
        export
    endif
endif

.PHONY: local dev prod run build-and-run goose-status goose-up goose-down check-env

# Target to validate that APP_ENV is defined and valid
check-env:
	$(if $(APP_ENV),,$(error APP_ENV/env is not set. Please specify the environment (e.g., 'make <target> local', 'make env=local <target>', or 'make APP_ENV=local <target>')))
	$(if $(filter $(APP_ENV),local dev prod),,$(error invalid environment '$(APP_ENV)'. Must be one of: local, dev, prod))

# Targets to run the application (only defined if the main goal is not a goose command)
ifneq ($(filter $(FIRST_GOAL),$(SUPPORTED_GOOSE_TARGETS)),$(FIRST_GOAL))
local:
	@$(MAKE) APP_ENV=local run

dev:
	@$(MAKE) APP_ENV=dev run

prod:
	@$(MAKE) APP_ENV=prod build-and-run
endif

run: check-env
	air

build-and-run: check-env
	go build -o ./tmp/main.exe ./cmd/api
	./tmp/main.exe

# Targets for Goose database migrations
goose-status: check-env
	goose status

goose-up: check-env
	goose up

goose-down: check-env
	goose down
