.PHONY: run-dev prepare-env down flush e2e-test e2e-test-load

run-dev:
	@echo "Starting development environment..."
	docker compose up --build
	@echo "Shutting down development environment..."
	docker compose down

prepare-env:
	@echo "Preparing .env files"
	@if [ ! -f .env ]; then \
		echo "==> Copying .env.example to .env"; \
		cp .env.example .env; \
	else \
		echo "==> Skipping, .env already exists"; \
	fi
	@echo "Starting E2E tests..."
	cd services; \
	$(MAKE) prepare-env

E2E_COMMON_COMPOSE = -f docker-compose.yaml -f docker-compose.override.yaml --env-file ./../.env

down:
	@echo "Stopping top-level containers..."
	docker compose down --remove-orphans
	cd e2e; \
	echo "Stopping e2e containers..."; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e down --remove-orphans; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e-load down --remove-orphans

flush:
	@echo "Flushing top-level containers and volumes..."
	docker compose down --volumes --remove-orphans
	cd e2e; \
	echo "Flushing e2e containers and volumes..."; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e down --remove-orphans --volumes; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e-load down --remove-orphans --volumes

e2e-test:
	@echo "Running E2E tests..."
	cd e2e; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e up \
		--build \
		--abort-on-container-exit e2e-tests \
		--exit-code-from e2e-tests; \
	DOWN_EXIT=$$?; \
	cd ../; \
	make flush; \
	echo "E2E tests finished with exit code $$DOWN_EXIT"; \
	exit $$DOWN_EXIT

e2e-test-load:
	@echo "Running E2E load tests..."
	cd e2e; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e-load up \
		--build \
		--abort-on-container-exit e2e-tests-load \
		--exit-code-from e2e-tests-load; \
	DOWN_EXIT=$$?; \
	cd ../; \
	make flush; \
	echo "E2E load tests finished with exit code $$DOWN_EXIT"; \
	exit $$DOWN_EXIT
