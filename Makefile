.PHONY: test

test:
	go test -v ./...
fmt:
	go fmt ./...
lint:		## Run static code analysis
	golangci-lint run --timeout 5m --fix
