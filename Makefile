GO ?= go
GOFMT ?= gofmt

.PHONY: fmt check-fmt vet test build run check docker-build

fmt:
	$(GO) fmt ./...

check-fmt:
	@files="$$($(GOFMT) -l cmd internal)" || exit $$?; \
	if [ -n "$$files" ]; then \
		printf 'Run make fmt:\n%s\n' "$$files"; \
		exit 1; \
	fi

vet:
	$(GO) vet ./...

test:
	$(GO) test -race ./...

build:
	mkdir -p bin
	$(GO) build -trimpath -o bin/mafia-api ./cmd/mafia-api

run:
	$(GO) run ./cmd/mafia-api

check: check-fmt vet test build

docker-build:
	docker build -t mafia-api:local .
