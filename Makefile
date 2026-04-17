build:
	@echo "Building the project..."
	@CGO_ENABLED=0 go build -o financial_control main.go
	@echo "Build complete."

run: 
	@echo "Running the project..."
	@ENVIRONMENT=${ENVIRONMENT} go run main.go

run_sync: 
	@echo "Running the project..."
	@ENVIRONMENT=${ENVIRONMENT} go run main.go sync

run_budget: 
	@echo "Running the project..."
	@ENVIRONMENT=${ENVIRONMENT} go run main.go budget --date=${DATE}

run_budget_cards_and_others:
	@echo "Running the project..."
	@ENVIRONMENT=${ENVIRONMENT} go run main.go budget-cards-and-others --date=${DATE}

run_budget_unified:
	@echo "Running the project..."
	@ENVIRONMENT=${ENVIRONMENT} go run main.go budget-unified --date=${DATE}

run_budget_full:
	@echo "Running the project..."
	@ENVIRONMENT=${ENVIRONMENT} go run main.go budget-full --date=${DATE}

run_balance:
	@echo "Running the project..."
	@ENVIRONMENT=${ENVIRONMENT} go run main.go balance --date=${DATE}

run_budget_category:
	@echo "Running the project..."
	@ENVIRONMENT=${ENVIRONMENT} go run main.go budget-category --date=${DATE} --category=${CATEGORY}