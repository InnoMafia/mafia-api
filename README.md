# mafia-api

Go API платформы спортивной мафии InnoMafia. Один Go-сервис принимает вызовы
gRPC и HTTP/JSON. Для запуска с вебом и PostgreSQL используйте
[mafia-ops](https://github.com/InnoMafia/mafia-ops).

## Локальная работа

Для сборки нужны Go 1.27.4 и Make, для генерации и проверок — Buf 1.75.0.
При запуске укажите подключение к PostgreSQL 18.

```bash
make generate
make check
DATABASE_URL='postgresql://localhost/mafia' go run ./cmd/api
```

`HTTP_ADDR` по умолчанию `127.0.0.1:8080`, `GRPC_ADDR` —
`127.0.0.1:9090`. В Docker HTTP и gRPC доступны внутри сети Compose,
а Caddy принимает оба протокола на общем внешнем адресе.

| Маршрут | Ответ при доступной базе | Ответ при недоступной базе |
| --- | --- | --- |
| `GET /healthz` | 204 | 204 |
| `GET /readyz` | 204 | 503 |
| `GET /api/v1/status` | 200, `{"status":"ready"}` | 503 с кодом gRPC `Unavailable` |
| `PlatformService/GetServiceStatus` | `status: "ready"` | gRPC `Unavailable` |

## Контракт

Файл сервиса [platform.proto](api/mafia/platform/v1/platform.proto) объявляет
RPC и их HTTP-привязки `google.api.http`. Типы запроса и ответа каждого
метода лежат в своём `.proto`, а реализация метода — в своём Go-файле.
Например, у `GetServiceStatus` это [get_service_status.proto](api/mafia/platform/v1/get_service_status.proto)
и [get_service_status.go](internal/handler/get_service_status.go).

`make generate` запускает Buf и обновляет:

- Go-сообщения и gRPC-клиент/сервер в `gen/go`;
- HTTP/JSON-шлюз на gRPC-Gateway в `gen/go`;
- OpenAPI 3.1 в `internal/handler/openapi`.

Сервер отдаёт документ по `GET /api/openapi.json`. Веб-клиент генерирует
TypeScript-типы из этого артефакта. CI регенерирует файлы и сравнивает их
с коммитом.

```bash
curl --fail http://127.0.0.1:8080/api/v1/status
curl --fail http://127.0.0.1:8080/api/openapi.json
buf curl --reflect=false --schema . --protocol grpc \
  http://127.0.0.1:9090/mafia.platform.v1.PlatformService/GetServiceStatus
```

Первый продуктовый сценарий и границы клубных данных описаны в
[архитектуре](docs/architecture.md).
