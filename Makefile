SERVICES = apiservice contacts-worker notification-service rebalancer-service sender-service
E2E_COMMON_COMPOSE = -f docker-compose.yaml -f docker-compose.override.yaml --env-file ./../.env

.PHONY: all
all: lint unit-test integration-test e2e-test

.PHONY: run-dev
run-dev:
	docker compose up --build
	docker compose down

.PHONY: down
down:
	docker compose down --remove-orphans
	cd e2e; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e down --remove-orphans; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e-load down --remove-orphans

.PHONY: flush
flush:
	docker compose down --volumes --remove-orphans
	cd e2e; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e down --remove-orphans --volumes; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e-load down --remove-orphans --volumes

.PHONY: prepare-env
prepare-env:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
	else \
		echo "==> Skipping, .env already exists"; \
	fi
	@cd services; \
	for service in $(SERVICES); do \
		if [ ! -f $$service/.env ]; then \
			cp $$service/.env.example $$service/.env; \
		else \
			echo "==> Skipping $$service (.env already exists)"; \
		fi \
	done

.PHONY: lint
lint:
	@cd services; \
	for service in $(SERVICES); do \
		if [ -f $$service/Makefile ]; then \
			echo "==> $$service"; \
			$(MAKE) -C $$service lint || exit 1; \
		fi \
	done

.PHONY: unit-test
unit-test:
	@cd services; \
	for service in $(SERVICES); do \
		if [ -f $$service/Makefile ]; then \
			echo "==> $$service"; \
			$(MAKE) -C $$service unit-test || exit 1; \
		fi \
	done

.PHONY: integration-test
integration-test:
	@cd services; \
	for service in $(SERVICES); do \
		if [ -f $$service/Makefile ]; then \
			echo "==> $$service"; \
			$(MAKE) -C $$service integration-test || exit 1; \
		fi \
	done

.PHONY: coverage
coverage:
	@rm -f coverage*.out
	@for service in $(SERVICES); do \
		echo "==> $$service"; \
		cd services/$$service && go test -coverprofile=coverage.out ./... || exit 1; \
		cd ../../; \
	done; \
	gocovmerge services/*/coverage.out > coverage_total.out; \

.PHONY: e2e-test
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

.PHONY: e2e-test-load
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
