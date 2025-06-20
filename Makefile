.DEFAULT_GOAL := lint

.PHONY: fmt
fmt:
	@gofumpt -w -l .
	@goimports -w -l .
	@golangci-lint run --fix

.PHONY: lint
lint:
	@golangci-lint run

