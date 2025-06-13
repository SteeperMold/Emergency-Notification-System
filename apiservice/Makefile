.PHONY: up test down lint flush

up:
	docker compose up \
		--build \
		--abort-on-container-exit apiservice \
		--exit-code-from apiservice; \
	DOWN_EXIT=$$?; \
	docker compose down; \
	exit $$DOWN_EXIT

down:
	docker compose down

test:
	docker compose -f docker-compose.test.yaml up \
		--build \
		--abort-on-container-exit test \
		--exit-code-from test; \
	DOWN_EXIT=$$?; \
	docker compose -f docker-compose.test.yaml down --volumes; \
	exit $$DOWN_EXIT

lint:
	golangci-lint run ./...

flush:
	docker compose down --volumes