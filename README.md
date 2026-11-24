# mafia-api

Единый Go-бэкенд InnoMafia для судейского интерфейса, OBS-оверлея и обмена
данными с MafiaUniverse. Модульный монолит для self-hosted установки клуба.

Сейчас здесь только инфраструктурная основа: HTTP-сервер, проверки состояния,
конфигурация, корректное завершение работы, тесты и контейнеризация.
Игровые правила, хранилище, авторизация и WebSocket ещё не реализованы.

## Быстрый старт

Требуются Go **1.27.4** и Make.

```sh
make run
```

По умолчанию сервер доступен только локально. В другом терминале:

```sh
curl --fail http://127.0.0.1:8080/healthz
curl --fail http://127.0.0.1:8080/readyz
```

Оба эндпоинта возвращают `{"status":"ok"}`. Сейчас readiness подтверждает только
работу HTTP-процесса, а не доступность базы или внешних сервисов.
`SIGINT` и `SIGTERM` запускают graceful shutdown.

## Конфигурация

| Переменная | По умолчанию | Назначение |
| --- | --- | --- |
| `HTTP_ADDR` | `127.0.0.1:8080` | TCP-адрес HTTP-сервера |
| `SHUTDOWN_TIMEOUT` | `10s` | Время на завершение активных HTTP-запросов |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` или `error` |

```sh
HTTP_ADDR=127.0.0.1:9090 LOG_LEVEL=debug make run
```

Конфигурация читается из окружения; `.env` автоматически не загружается.
Некорректные значения останавливают запуск. Логи пишутся в stdout в JSON.

## Разработка

```sh
make fmt         # Форматирование
make check       # Форматирование без изменений, vet, тесты с race detector, сборка
make build       # bin/mafia-api
```

Внешних Go-зависимостей пока нет, поэтому `go.sum` отсутствует намеренно.
GitHub Actions выполняет проверки и сборку контейнера, но не публикует образы
и не развёртывает приложение.

```text
cmd/mafia-api/       Точка входа
internal/config/    Чтение и проверка окружения
internal/httpapi/   HTTP-обработчики
internal/httpserver/ Жизненный цикл HTTP-сервера
deploy/             Compose для API и пример окружения
docs/deployment.md  Развёртывание из нескольких репозиториев
```

## Docker

```sh
make docker-build
docker run --rm -p 127.0.0.1:8080:8080 mafia-api:local
```

Образ запускается без root и shell. В контейнере `HTTP_ADDR=:8080`,
поэтому ограничивать внешний доступ нужно публикацией порта или reverse proxy.
Проверки состояния выполняются HTTP-запросами извне контейнера.

Локальный вариант через Compose:

```sh
docker compose -f deploy/compose.yaml up -d --build
curl --fail --retry 10 --retry-connrefused --retry-delay 1 http://127.0.0.1:8080/readyz
docker compose -f deploy/compose.yaml logs api
docker compose -f deploy/compose.yaml down
```

Нужен Compose v2. Если он установлен отдельным исполняемым файлом,
используйте `docker-compose` вместо `docker compose`.

Текущий сервис не хранит данные и не требует тома. Он не заменяет `mafia-old`.
До появления авторизации не выставляйте будущие игровые эндпоинты в интернет.

Модель выпуска API, web и HUD описана в [docs/deployment.md](docs/deployment.md).
