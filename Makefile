all : install install run run/docker test test/coverage generate generate/api
.PHONY : all

install:
	@go version
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "All tools installed."

run:
	@docker-compose up -d postgres
	@air
run/docker:
	@docker compose up --build --force-recreate -V

test:
	@set -euo pipefail
	@go test ./... -json -v -race 2>&1 | tee /tmp/gotest.log | gotestfmt
test/coverage:
	@set -euo pipefail
	@go test ./... -json -v -race -coverpkg=./... -coverprofile=coverage.out -covermode=atomic 2>&1 | tee /tmp/gotest.log | gotestfmt
	@go tool cover -html coverage.out -o coverage.html
	@open coverage.html

generate: generate/api
generate/api:
	@swag init
