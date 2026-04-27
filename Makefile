.PHONY: build test vet lint run run_sync run_budget run_budget_cards_and_others run_budget_unified run_budget_full run_balance run_budget_category mocks mocks/clean

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
