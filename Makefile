.PHONY: up-all down-all flush-all

up-all:
	docker compose up \
		--build \
		--abort-on-container-exit apiservice contacts-worker notification-service rebalancer-service sender-service react-app \
		--exit-code-from apiservice contacts-worker notification-service rebalancer-service sender-service react-app; \
	DOWN_EXIT=$$?; \
	docker compose down; \
	exit $$DOWN_EXIT

down-all:
	docker compose down

flush-all:
	docker compose down --volumes --remove-orphans