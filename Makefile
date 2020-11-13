.PHONY: test

test:
	go test -v ./...
fmt:
	go fmt ./...
static:
	staticcheck -f stylish $$(go list ./...)