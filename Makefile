.PHONY: build test vet run run_sync run_budget run_budget_cards_and_others run_budget_unified run_budget_full run_balance run_budget_category

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
