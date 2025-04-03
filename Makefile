build:
	@echo "Building the project..."
	@CGO_ENABLED=0 go build -o financial_control main.go
	@echo "Build complete."

run: 
	@echo "Running the project..."
	@ENVIRONMENT=${ENVIRONMENT} go run main.go

run_budget: 
	@echo "Running the project..."
	@ENVIRONMENT=${ENVIRONMENT} go run main.go budget --date=${DATE}

run_budget_cards_and_others: 
	@echo "Running the project..."
	@ENVIRONMENT=${ENVIRONMENT} go run main.go budget-cards-and-others --date=${DATE}