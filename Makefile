.PHONY: build run run-api test test-coverage test-ui clean help docker-build-api docker-build-web docker-build docker-push-api docker-push-web docker-push docker-auth security-audit security-audit-fix docker-down docker-down-test

# Build variables
BINARY_NAME=api-server
CMD_PATH=./cmd/api-server
BUILD_DIR=./bin
DOCKER_REGISTRY=us-central1-docker.pkg.dev/zeitgeistlabs/us-central1-docker
DOCKER_IMAGE_API=professors-research-api
DOCKER_IMAGE_WEB=professors-research-web

# Compute git hash with -dirty suffix if there are uncommitted changes
GIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY := $(shell git status --porcelain 2>/dev/null | grep -q . && echo "-dirty" || echo "")
DOCKER_TAG ?= $(GIT_HASH)$(GIT_DIRTY)
API_IMAGE=$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_API):$(DOCKER_TAG)
WEB_IMAGE=$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_WEB):$(DOCKER_TAG)

# Default target
.DEFAULT_GOAL := help

## build: Build the API server binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-web: Build the frontend web application
build-web:
	@echo "Building web frontend..."
	@if [ ! -d "web" ]; then \
		echo "Error: web directory not found"; \
		exit 1; \
	fi
	@mkdir -p web/dist
	@cd web && npm install && npm run build
	@echo "Web frontend build complete"

## run: Run the full application (API server + web dev server) in Docker
run:
	@echo "Starting full application in Docker containers..."
	@echo "API server: http://localhost:$${PORT:-8080}"
	@echo "Web dev server: http://localhost:5173"
	@echo "Press Ctrl+C to stop both servers"
	@PORT=$${PORT:-8080} docker compose -f docker-compose.dev.yml up --build

## run-api: Run only the API server in Docker
run-api:
	@echo "Starting API server in Docker container..."
	@echo "API server: http://localhost:$${PORT:-8080}"
	@echo "Press Ctrl+C to stop the server"
	@PORT=$${PORT:-8080} docker compose -f docker-compose.dev.yml up --build api

## test: Run all Go tests in Docker
test:
	@echo "Running Go tests in Docker container..."
	@docker compose -f docker-compose.dev.yml run --rm --no-deps api go test -v ./internal/... ./pkg/...

## test-coverage: Run tests with coverage in Docker
test-coverage:
	@echo "Running tests with coverage in Docker container..."
	@docker compose -f docker-compose.dev.yml run --rm --no-deps api sh -c "go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html"
	@echo "Coverage report generated: coverage.html"

## test-ui: Run UI acceptance tests in Docker (starts API and web servers automatically)
test-ui:
	@echo "Running UI acceptance tests in Docker containers..."
	@echo "Cleaning up any existing test containers and networks..."
	@docker compose -f docker-compose.test.yml down --remove-orphans --volumes 2>/dev/null || true
	@docker rm -f professors-research-api-test professors-research-web-test professors-research-test-runner 2>/dev/null || true
	@EXIT_CODE=0; \
	PORT=$${PORT:-8080} docker compose -f docker-compose.test.yml --profile test up --build --abort-on-container-exit --exit-code-from test || EXIT_CODE=$$?; \
	docker compose -f docker-compose.test.yml down --remove-orphans --volumes 2>/dev/null || true; \
	docker rm -f professors-research-api-test professors-research-web-test professors-research-test-runner 2>/dev/null || true; \
	echo "Test containers and networks cleaned up"; \
	exit $$EXIT_CODE

## docker-build-api: Build the API server Docker image
docker-build-api:
	@echo "Building API server Docker image for linux/amd64..."
	@docker build --platform linux/amd64 -t $(API_IMAGE) -t $(DOCKER_IMAGE_API):$(DOCKER_TAG) -f Dockerfile .
	@echo "API server Docker image built: $(API_IMAGE)"

## docker-build-web: Build the web frontend Docker image
docker-build-web:
	@echo "Building web frontend Docker image for linux/amd64..."
	@docker build --platform linux/amd64 -t $(WEB_IMAGE) -t $(DOCKER_IMAGE_WEB):$(DOCKER_TAG) -f web/Dockerfile web/
	@echo "Web frontend Docker image built: $(WEB_IMAGE)"

## docker-build: Build both Docker images
docker-build: docker-build-api docker-build-web

## docker-push-api: Push the API server Docker image to registry
docker-push-api: docker-build-api
	@echo "Pushing API server Docker image to $(DOCKER_REGISTRY)..."
	@docker push $(API_IMAGE) || ( \
		echo ""; \
		echo "Error: Push failed. If you see 'Unauthenticated request', run 'make docker-auth' to configure authentication."; \
		exit 1 \
	)
	@echo "API server Docker image pushed: $(API_IMAGE)"
	@echo "Updating image tag in kubernetes/api-deployment.yaml..."
	@sed -i.bak "s|\(image: us-central1-docker.pkg.dev/zeitgeistlabs/us-central1-docker/professors-research-api:\).*|\1$(DOCKER_TAG)|" kubernetes/api-deployment.yaml && rm -f kubernetes/api-deployment.yaml.bak
	@echo "Updated kubernetes/api-deployment.yaml with tag: $(DOCKER_TAG)"

## docker-push-web: Push the web frontend Docker image to registry
docker-push-web: docker-build-web
	@echo "Pushing web frontend Docker image to $(DOCKER_REGISTRY)..."
	@docker push $(WEB_IMAGE) || ( \
		echo ""; \
		echo "Error: Push failed. If you see 'Unauthenticated request', run 'make docker-auth' to configure authentication."; \
		exit 1 \
	)
	@echo "Web frontend Docker image pushed: $(WEB_IMAGE)"
	@echo "Updating image tag in kubernetes/web-deployment.yaml..."
	@sed -i.bak "s|\(image: us-central1-docker.pkg.dev/zeitgeistlabs/us-central1-docker/professors-research-web:\).*|\1$(DOCKER_TAG)|" kubernetes/web-deployment.yaml && rm -f kubernetes/web-deployment.yaml.bak
	@echo "Updated kubernetes/web-deployment.yaml with tag: $(DOCKER_TAG)"

## docker-push: Build and push both Docker images to registry
docker-push: docker-push-api docker-push-web

## docker-auth: Configure Docker authentication for Google Artifact Registry
docker-auth:
	@echo "Configuring Docker authentication for $(DOCKER_REGISTRY)..."
	@GCLOUD_BIN=$$(command -v gcloud 2>/dev/null || echo "/Users/vallery/Downloads/gcloud-sdk/bin/gcloud"); \
	if [ -f "$$GCLOUD_BIN" ]; then \
		$$GCLOUD_BIN auth configure-docker us-central1-docker.pkg.dev --quiet; \
		echo "Docker authentication configured successfully"; \
	else \
		echo "Error: gcloud CLI not found"; \
		echo "Expected at: /Users/vallery/Downloads/gcloud-sdk/bin/gcloud"; \
		echo "Please install gcloud CLI or update the path in the Makefile"; \
		echo "See: https://cloud.google.com/sdk/docs/install"; \
		exit 1; \
	fi

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf web/dist web/node_modules
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## security-audit: Run npm audit to check for vulnerabilities in node_modules (in Docker)
security-audit:
	@echo "Running security audit for web dependencies in Docker..."
	@docker compose -f docker-compose.dev.yml run --rm --no-deps web npm audit || ( \
		echo ""; \
		echo "Security audit found vulnerabilities. Review the report above."; \
		echo "Run 'make security-audit-fix' to attempt automatic fixes."; \
		exit 1 \
	)

## security-audit-fix: Run npm audit fix to automatically fix vulnerabilities
security-audit-fix:
	@echo "Running security audit fix for web dependencies in Docker..."
	@docker compose -f docker-compose.dev.yml run --rm --no-deps web npm audit fix
	@echo "Security audit fix complete"

## docker-down: Stop and remove dev Docker containers
docker-down:
	@echo "Stopping dev Docker containers..."
	@docker compose -f docker-compose.dev.yml down
	@echo "Dev containers stopped"

## docker-down-test: Stop and remove test Docker containers
docker-down-test:
	@echo "Stopping test Docker containers..."
	@docker compose -f docker-compose.test.yml down --remove-orphans 2>/dev/null || true
	@echo "Test containers stopped"

## help: Show this help message
help:
	@echo "Available targets:"
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'

