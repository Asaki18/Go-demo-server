# go-orders-demo

Минимальный демонстрационный сервис: Go + Kafka + Postgres + in-memory cache + web UI.

## Структура
- `cmd/app` — точка входа
- `internal/api` — HTTP слой
- `internal/kafka` — producer/consumer
- `internal/db` — адаптер базы (интерфейс + реализация)
- `internal/cache` — in-memory cache (реализует интерфейс)
- `internal/models` — модели заказа
- `sql/migrations` — SQL миграции

## Быстрый старт (локально)
1. Прописать `docker compose` (в корне).
2. Запустить:
```bash
docker compose up -d --build
