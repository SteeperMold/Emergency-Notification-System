.PHONY: test lint

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
