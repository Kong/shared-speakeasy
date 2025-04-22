.PHONY: install-tools
install-tools: ## install the required dependencies for the repository
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2

.PHONY: lint-all
lint-all: ## lint all packages of the repository
	find . -name go.sum -exec dirname {} \; | xargs -I{} /bin/sh -c "cd {} && golangci-lint run ./..."

.PHONY: lint-fix-all
lint-fix-all:
	find . -name go.sum -exec dirname {} \; | xargs -I{} /bin/sh -c "cd {} && golangci-lint run --fix ./..."

.PHONY: test-all
test-all: ## lint all packages of the repository
	find . -name go.sum -exec dirname {} \; | xargs -I{} /bin/sh -c "cd {} && go test ./... -v"
