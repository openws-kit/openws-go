.DEFAULT_GOAL := lint

.PHONY: fmt lint test

tool.sum:
	@go mod tidy -modfile=tool.mod

fmt: tool.sum
	@go tool -modfile=tool.mod gofumpt -w -l .
	@go tool -modfile=tool.mod goimports -w -l .
	@go tool -modfile=tool.mod golangci-lint run --fix

lint: tool.sum
	@go tool -modfile=tool.mod golangci-lint run
