.PHONY: lint-all
lint-all: ## lint all packages of the repository
	find . -name go.sum -exec dirname {} \; | xargs -I{} /bin/sh -c "cd {} && golangci-lint run ./..."

.PHONY: lint-fix-all
lint-fix-all:
	find . -name go.sum -exec dirname {} \; | xargs -I{} /bin/sh -c "cd {} && golangci-lint run --fix ./..."
