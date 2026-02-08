.PHONY: build run run-api test test-local test-coverage test-ui test-ui-local clean help docker-build-api docker-build-web docker-build docker-push-api docker-push-web docker-push docker-auth security-audit security-audit-fix docker-down docker-down-test

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
	@echo "Cleaning up any existing dev containers that may be holding ports..."
	@docker compose -f docker-compose.dev.yml down --remove-orphans 2>/dev/null || true
	@docker rm -f professors-research-api-dev professors-research-web-dev 2>/dev/null || true
	@PORT=$${PORT:-8080}; \
	if command -v lsof >/dev/null 2>&1; then \
		PIDS=""; \
		for p in $$PORT 5173; do \
			FOUND=$$(lsof -nP -tiTCP:$$p -sTCP:LISTEN 2>/dev/null || true); \
			if [ -n "$$FOUND" ]; then \
				PIDS="$$PIDS $$FOUND"; \
			fi; \
		done; \
		if [ -n "$$PIDS" ]; then \
			if [ "$${FORCE_KILL_PORTS:-0}" = "1" ]; then \
				echo "Ports in use; FORCE_KILL_PORTS=1 set, killing:$$PIDS"; \
				kill $$PIDS 2>/dev/null || true; \
				sleep 1; \
			else \
				echo "Error: required ports are in use (PORT=$$PORT and/or 5173)."; \
				echo "Set FORCE_KILL_PORTS=1 to auto-kill local listeners, or change PORT (e.g. PORT=18080 make run)."; \
				for p in $$PORT 5173; do \
					echo "--- listeners on $$p ---"; \
					lsof -nP -iTCP:$$p -sTCP:LISTEN 2>/dev/null || true; \
				done; \
				exit 1; \
			fi; \
		fi; \
	fi; \
	PORT=$$PORT docker compose -f docker-compose.dev.yml up --build

## run-api: Run only the API server in Docker
run-api:
	@echo "Starting API server in Docker container..."
	@echo "API server: http://localhost:$${PORT:-8080}"
	@echo "Press Ctrl+C to stop the server"
	@echo "Cleaning up any existing dev containers that may be holding the API port..."
	@docker compose -f docker-compose.dev.yml down --remove-orphans 2>/dev/null || true
	@docker rm -f professors-research-api-dev 2>/dev/null || true
	@PORT=$${PORT:-8080}; \
	if command -v lsof >/dev/null 2>&1; then \
		FOUND=$$(lsof -nP -tiTCP:$$PORT -sTCP:LISTEN 2>/dev/null || true); \
		if [ -n "$$FOUND" ]; then \
			if [ "$${FORCE_KILL_PORTS:-0}" = "1" ]; then \
				echo "Port $$PORT in use; FORCE_KILL_PORTS=1 set, killing: $$FOUND"; \
				kill $$FOUND 2>/dev/null || true; \
				sleep 1; \
			else \
				echo "Error: port $$PORT is in use."; \
				echo "Set FORCE_KILL_PORTS=1 to auto-kill local listeners, or change PORT (e.g. PORT=18080 make run-api)."; \
				lsof -nP -iTCP:$$PORT -sTCP:LISTEN 2>/dev/null || true; \
				exit 1; \
			fi; \
		fi; \
	fi; \
	PORT=$$PORT docker compose -f docker-compose.dev.yml up --build api

## test: Run all Go tests in Docker
test:
	@echo "Running Go tests..."
	@if docker info >/dev/null 2>&1; then \
		echo "Using Docker..."; \
		docker compose -f docker-compose.dev.yml run --rm --no-deps api go test -v ./internal/... ./pkg/...; \
	else \
		echo "Docker not available; running tests locally (go test)."; \
		go test -v ./internal/... ./pkg/...; \
	fi

## test-local: Run all Go tests locally (no Docker required)
test-local:
	@echo "Running Go tests locally..."
	@go test -v ./internal/... ./pkg/...

## test-coverage: Run tests with coverage in Docker
test-coverage:
	@echo "Running tests with coverage..."
	@if docker info >/dev/null 2>&1; then \
		echo "Using Docker..."; \
		docker compose -f docker-compose.dev.yml run --rm --no-deps api sh -c "go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html"; \
	else \
		echo "Docker not available; running coverage locally."; \
		go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html; \
	fi
	@echo "Coverage report generated: coverage.html"

## test-ui: Run UI acceptance tests in Docker (starts API and web servers automatically)
test-ui:
	@if docker info >/dev/null 2>&1; then \
		echo "Running UI acceptance tests in Docker containers..."; \
		echo "Cleaning up any existing test containers and networks..."; \
		docker compose -f docker-compose.test.yml down --remove-orphans --volumes 2>/dev/null || true; \
		docker rm -f professors-research-api-test professors-research-web-test professors-research-test-runner 2>/dev/null || true; \
		EXIT_CODE=0; \
		PORT=$${PORT:-8080} docker compose -f docker-compose.test.yml --profile test up --build --abort-on-container-exit --exit-code-from test || EXIT_CODE=$$?; \
		docker compose -f docker-compose.test.yml down --remove-orphans --volumes 2>/dev/null || true; \
		docker rm -f professors-research-api-test professors-research-web-test professors-research-test-runner 2>/dev/null || true; \
		echo "Test containers and networks cleaned up"; \
		exit $$EXIT_CODE; \
	else \
		echo "Docker not available; running UI acceptance tests locally (make test-ui-local)."; \
		$(MAKE) test-ui-local; \
	fi

## test-ui-local: Run UI acceptance tests without Docker (starts API locally; Playwright starts Vite dev server)
test-ui-local:
	@echo "Running UI acceptance tests locally (no Docker)..."
	@echo "Building and starting API server on :18080 (override with API_PORT=...)..."
	@set -e; \
	API_PORT=$${API_PORT:-18080}; \
	API_URL="http://localhost:$$API_PORT"; \
	go build -o ./bin/api-server ./cmd/api-server; \
	(PORT=$$API_PORT ./bin/api-server >/tmp/professors-research-api.log 2>&1 & echo $$! > /tmp/professors-research-api.pid); \
	trap 'kill $$(cat /tmp/professors-research-api.pid) 2>/dev/null || true; rm -f /tmp/professors-research-api.pid' EXIT INT TERM; \
	echo "Waiting for API to finish loading card cache..."; \
	for i in $$(seq 1 120); do \
		if curl -s "$$API_URL/api/health" | grep -q '"status":"ok"'; then \
			break; \
		fi; \
		sleep 1; \
	done; \
	curl -s "$$API_URL/api/health" | grep -q '"status":"ok"' || (echo "API did not become ready. See /tmp/professors-research-api.log" && exit 1); \
	echo "Ensuring Playwright browsers are installed (chromium)..."; \
	cd web; \
	npx playwright install chromium; \
	echo "Running Playwright tests (chromium only)..."; \
	VITE_API_URL=$$API_URL CI= npm test -- --project=chromium

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

