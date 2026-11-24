FROM golang:1.27.4-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/mafia-api ./cmd/mafia-api

FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.source="https://github.com/InnoMafia/mafia-api"
LABEL org.opencontainers.image.description="InnoMafia Go API"

WORKDIR /app
COPY --from=build --chown=65532:65532 /out/mafia-api /app/mafia-api

ENV HTTP_ADDR=:8080
EXPOSE 8080
USER 65532:65532

ENTRYPOINT ["/app/mafia-api"]
