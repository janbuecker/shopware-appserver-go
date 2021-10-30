.PHONY: test

test:
	go test -v ./...
fmt:
	go fmt ./...
lint:		## Run static code analysis
ifeq (, $(shell which golangci-lint))
	@echo Local binary not found, using docker.
	@echo Download local binary at https://golangci-lint.run/usage/install/#local-installation
	docker run -v $$(pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run --timeout 5m --fix
else
	golangci-lint run --timeout 5m --fix
endif