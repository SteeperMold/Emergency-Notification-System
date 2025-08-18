.PHONY: run-dev down flush e2e-test e2e-test-load

run-dev:
	docker compose up --build
	docker compose down;

E2E_COMMON_COMPOSE = -f docker-compose.yaml -f docker-compose.override.yaml --env-file ./../.env

down:
	docker compose down --remove-orphans
	cd e2e; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e down --remove-orphans; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e-load down --remove-orphans

flush:
	docker compose down --volumes --remove-orphans
	cd e2e; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e down --remove-orphans --volumes; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e-load down --remove-orphans --volumes

e2e-test:
	cd e2e; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e up \
		--build \
		--abort-on-container-exit e2e-tests \
		--exit-code-from e2e-tests; \
	DOWN_EXIT=$$?; \
	cd ../; \
	make flush; \
	exit $$DOWN_EXIT

e2e-test-load:
	cd e2e; \
	docker compose $(E2E_COMMON_COMPOSE) --profile e2e-load up \
		--build \
		--abort-on-container-exit e2e-tests-load \
		--exit-code-from e2e-tests-load; \
	DOWN_EXIT=$$?; \
	cd ../; \
	make flush; \
	exit $$DOWN_EXIT
