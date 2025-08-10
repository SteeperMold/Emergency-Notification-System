.PHONY: run-dev down flush

run-dev:
	docker compose up \
		--build \
		--abort-on-container-exit apiservice contacts-worker notification-service rebalancer-service sender-service react-app \
		--exit-code-from apiservice contacts-worker notification-service rebalancer-service sender-service react-app; \
	DOWN_EXIT=$$?; \
	docker compose down; \
	exit $$DOWN_EXIT

down:
	docker compose down

flush:
	docker compose down --volumes --remove-orphans