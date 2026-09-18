.PHONY: test
test: lint
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run ./...
