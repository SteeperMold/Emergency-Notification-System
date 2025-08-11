.PHONY: run-dev down flush e2e-test

run-dev:
	docker compose up --build
	docker compose down;

down:
	docker compose down --remove-orphans

flush:
	docker compose down --volumes --remove-orphans


e2e-test:
	docker compose -f docker-compose.e2e-test.yaml up --build \
		--abort-on-container-exit e2e-tests \
		--exit-code-from e2e-tests; \
	DOWN_EXIT=$$?; \
	docker compose -f docker-compose.e2e-test.yaml down --volumes --remove-orphans; \
	exit $$DOWN_EXIT
