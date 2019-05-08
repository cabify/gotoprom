.PHONY: test benchmark help fmt install

help: ## Show the help text
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[36m%-20s\033[93m %s\n", $$1, $$2}'

test: ## Run unit tests
	@go test -race ./...

benchmark: ## Run benchmarks
	@go test -bench=. ./...

check-fmt: ## Check file forma
	@GOIMP=$$(for f in $$(find . -type f -name "*.go" ! -path "./.cache/*" ! -path "./vendor/*" ! -name "bindata.go") ; do \
		goimports -l $$f ; \
	done) && echo $$GOIMP && test -z "$$GOIMP"

fmt: ## Format files
	@goimports -w $$(find . -name "*.go" -not -path "./vendor/*")

install: ## Installs dependencies
	GOPATH=$$GOPATH && go get -u -v \
		golang.org/x/tools/cmd/goimports
