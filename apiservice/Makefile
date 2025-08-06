.PHONY: unit-test integration-test lint

unit-test:
	go test -v -race -cover ./...

integration-test:
	go test -v -race -cover ./... -tags=integration

lint:
	golangci-lint run ./...
