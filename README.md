# mafia-api

Go API платформы спортивной мафии InnoMafia. Стартовая версия поднимает
HTTP-сервис, проверяет соединение с PostgreSQL и публикует первый типизированный
маршрут. Для запуска с вебом и базой используйте [mafia-ops](https://github.com/InnoMafia/mafia-ops).

## Локальная работа

Требуются Go 1.27.4, Make и PostgreSQL 18 для запуска сервиса.

```bash
make check
make generate
DATABASE_URL='postgresql://localhost/mafia' go run ./cmd/api
```

`DATABASE_URL` задаёт подключение PostgreSQL. Добавьте локальные учётные данные
и параметры TLS согласно своей конфигурации. `HTTP_ADDR` по умолчанию
`127.0.0.1:8080`. Контейнер слушает `:8080`.

| Маршрут | Ответ при доступной базе | Ответ при недоступной базе |
| --- | --- | --- |
| `GET /healthz` | 204 | 204 |
| `GET /readyz` | 204 | 503 |
| `GET /api/v1/status` | 200, `{"status":"ready"}` | 503, `{"status":"unavailable"}` |

Редактируйте [спецификацию OpenAPI](api/openapi.yaml) и запускайте `make generate`.
CI проверяет результат генерации, тесты с race detector и сборку образа.

Первый продуктовый сценарий и границы клубных данных описаны в
[архитектуре](docs/architecture.md).
