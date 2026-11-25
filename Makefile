BUF ?= buf

.PHONY: check generate

generate:
	$(BUF) generate

check:
	$(BUF) lint
	$(BUF) format --diff --exit-code
	test -z "$$(gofmt -l openapi.go cmd internal gen)"
	go vet ./...
	go test -race ./...
	go build ./...
