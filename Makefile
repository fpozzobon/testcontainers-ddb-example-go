.PHONY: test

help:
	@grep -E '^[a-zA-Z0-9_.$$()-/]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test.unit: ## Run unit test
	go test -cover -p 8 -v ./...