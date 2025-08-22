.PHONY: lint unit-test integration-test prepare-env

SERVICES := apiservice contacts-worker notification-service rebalancer-service sender-service

lint:
	@echo "Running linters for all services..."
	@for service in $(SERVICES); do \
		if [ -f $$service/Makefile ]; then \
			echo "==> $$service"; \
			$(MAKE) -C $$service lint || exit 1; \
		fi \
	done

unit-test:
	@echo "Running unit tests for all services..."
	@for service in $(SERVICES); do \
		if [ -f $$service/Makefile ]; then \
			echo "==> $$service"; \
			$(MAKE) -C $$service unit-test || exit 1; \
		fi \
	done

integration-test:
	@echo "Running integration tests for all services..."
	@for service in $(SERVICES); do \
		if [ -f $$service/Makefile ]; then \
			echo "==> $$service"; \
			$(MAKE) -C $$service integration-test || exit 1; \
		fi \
	done

prepare-env:
	@echo "Preparing environment files for all services..."
	@for service in $(SERVICES); do \
		if [ ! -f $$service/.env ]; then \
			echo "==> Copying .env.example to .env in $$service"; \
			cp $$service/.env.example $$service/.env; \
		else \
			echo "==> Skipping $$service (.env already exists)"; \
		fi \
	done
