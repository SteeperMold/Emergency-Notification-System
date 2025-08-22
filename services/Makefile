.PHONY: lint unit-test integration-test

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
