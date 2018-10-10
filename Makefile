.PHONY: test benchmark help fmt install

help: ## Show the help text
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[36m%-20s\033[93m %s\n", $$1, $$2}'

test: ## Run unit tests
	@go test -race ./...

benchmark: ## Run benchmarks
	@go test -bench=. ./...

check-fmt: ## Check file format
	@goimports -l $$(find . -name "*.go" -not -path "./vendor/*")

fmt: ## Format files
	@goimports -w $$(find . -name "*.go" -not -path "./vendor/*")

install: ## Installs dependencies
	GOPATH=$$GOPATH && go get -u -v \
		github.com/golang/dep/cmd/dep \
		golang.org/x/tools/cmd/goimports
	@dep ensure
