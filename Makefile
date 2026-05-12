.PHONY: build test vet lint run run_sync run_budget run_budget_cards_and_others run_budget_unified run_budget_full run_balance run_budget_category mocks mocks/clean infra-up infra-down infra-logs infra-migrate infra-wait infra-db-create infra-finalize

# Mockery v2 pinned — see ADR-003. v2.46.0 incompatível com Go 1.26; mínimo v2.53.6.
MOCKERY ?= go run github.com/vektra/mockery/v2@v2.53.6

ENTRYPOINT := ./cmd/financialcontrol-api
BINARY := financial_control

build:
	@echo "Building the project..."
	@CGO_ENABLED=0 go build -o $(BINARY) $(ENTRYPOINT)
	@echo "Build complete."

test:
	@echo "Running tests..."
	@go test --coverprofile tests/coverage.out ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

run:
	@echo "Running the project..."
	@ENVIRONMENT=${ENVIRONMENT} go run $(ENTRYPOINT)

run_sync:
	@echo "Running the sync command..."
	@ENVIRONMENT=${ENVIRONMENT} go run $(ENTRYPOINT) sync

run_budget:
	@echo "Running the budget command..."
	@ENVIRONMENT=${ENVIRONMENT} go run $(ENTRYPOINT) budget --date=${DATE}

run_budget_cards_and_others:
	@echo "Running the budget cards and others command..."
	@ENVIRONMENT=${ENVIRONMENT} go run $(ENTRYPOINT) budget-cards-and-others --date=${DATE}

run_budget_unified:
	@echo "Running the unified budget command..."
	@ENVIRONMENT=${ENVIRONMENT} go run $(ENTRYPOINT) budget-unified --date=${DATE}

run_budget_full:
	@echo "Running the full budget command..."
	@ENVIRONMENT=${ENVIRONMENT} go run $(ENTRYPOINT) budget-full --date=${DATE}

run_balance:
	@echo "Running the balance command..."
	@ENVIRONMENT=${ENVIRONMENT} go run $(ENTRYPOINT) balance --date=${DATE}

run_budget_category:
	@echo "Running the budget category command..."
	@ENVIRONMENT=${ENVIRONMENT} go run $(ENTRYPOINT) budget-category --date=${DATE} --category=${CATEGORY}

mocks/clean:
	@echo "Removing generated mocks..."
	@find . -type d -name mocks -prune -exec rm -rf {} +
	@echo "Mocks removed."

mocks: mocks/clean
	@echo "Generating mocks..."
	@$(MOCKERY) --config mockery.yml
	@echo "Mocks generated."

VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
MIGRATION_LDFLAGS := -X 'github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/migration.Version=$(VERSION)' \
                     -X 'github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/migration.Commit=$(COMMIT)' \
                     -X 'github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/migration.BuildDate=$(BUILD_DATE)'

.PHONY: migrate-build migrate-up migrate-baseline
migrate-build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w $(MIGRATION_LDFLAGS)" -o bin/migration ./cmd/migration
migrate-up: migrate-build
	./bin/migration
migrate-baseline: migrate-build
	@test -n "$(BASELINE)" || (echo "usage: make migrate-baseline BASELINE=N" && exit 2)
	MIGRATION_BASELINE=$(BASELINE) ./bin/migration

COMPOSE_FILE       := deployments/docker/docker-compose.yml
COMPOSE            := docker compose -f $(COMPOSE_FILE)
MSSQL_DB           ?= financial_control
MSSQL_SA_PASSWORD  ?= @docker@2021
MSSQL_WAIT_RETRIES ?= 60

infra-up: infra-wait infra-db-create infra-migrate infra-finalize
	@echo "Infra ready: mssql up, database $(MSSQL_DB) present, migrations applied up to v2."

infra-wait:
	@echo "Starting MSSQL container..."
	@$(COMPOSE) up -d mssql
	@echo "Waiting for MSSQL to accept connections..."
	@for i in $$(seq 1 $(MSSQL_WAIT_RETRIES)); do \
	  if docker exec mssql /opt/mssql-tools18/bin/sqlcmd -No -S localhost -U sa -P '$(MSSQL_SA_PASSWORD)' -Q 'SELECT 1' >/dev/null 2>&1; then \
	    echo "MSSQL ready (attempt $$i)"; \
	    exit 0; \
	  fi; \
	  printf '.'; sleep 2; \
	done; \
	echo "MSSQL did not become ready in time"; exit 1

infra-db-create:
	@echo "Ensuring database $(MSSQL_DB) exists..."
	@docker exec mssql /opt/mssql-tools18/bin/sqlcmd -No -S localhost -U sa -P '$(MSSQL_SA_PASSWORD)' \
	  -Q "IF DB_ID('$(MSSQL_DB)') IS NULL CREATE DATABASE [$(MSSQL_DB)]"

infra-migrate:
	@echo "Running migrations..."
	@# 000003 is a prod-cutover migration that drops legacy tables and requires
	@# the smoke hook (HTTP runtime) — it intentionally fires THROW 50001 here.
	@# The leading "-" lets the chain continue; infra-finalize cleans dirty state.
	-@$(COMPOSE) --profile migrate run --rm --build migration

infra-finalize:
	@echo "Finalizing schema_migrations (cap at v2 for local dev)..."
	@docker exec mssql /opt/mssql-tools18/bin/sqlcmd -No -S localhost -U sa -P '$(MSSQL_SA_PASSWORD)' \
	  -d $(MSSQL_DB) \
	  -Q "IF EXISTS (SELECT 1 FROM schema_migrations WHERE version = 3 AND dirty = 1) UPDATE schema_migrations SET version = 2, dirty = 0;"

infra-down:
	@$(COMPOSE) --profile migrate down

infra-logs:
	@$(COMPOSE) logs -f mssql
