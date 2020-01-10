.PHONY: test help fmt report-coveralls benchmark lint

help: ## Show the help text
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[36m%-20s\033[93m %s\n", $$1, $$2}'

test: ## Run unit tests
	@go test -coverprofile=coverage.out -covermode=atomic -race ./...

lint: # Run linters using golangci-lint 
	@golangci-lint run

fmt: ## Format files
	@goimports -w $$(find . -name "*.go" -not -path "./vendor/*")

benchmark: ## Run benchmarks
	@go test -run=NONE -benchmem -benchtime=5s -bench=. ./...	
