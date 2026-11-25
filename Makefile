.PHONY: check generate

generate:
	go tool oapi-codegen --config api/codegen.yaml api/openapi.yaml

check:
	test -z "$$(gofmt -l cmd internal)"
	go vet ./...
	go test -race ./...
	go build ./...
