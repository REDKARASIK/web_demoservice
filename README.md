# 📦 Web Demoservice

![Go](https://img.shields.io/badge/Go-1.24-blue?logo=go)
![Docker](https://img.shields.io/badge/Docker-✔-2496ED?logo=docker)
![Postgres](https://img.shields.io/badge/Postgres-17-4169E1?logo=postgresql)
![Kafka](https://img.shields.io/badge/Kafka%2FRedpanda-✔-D21F3C?logo=apache-kafka)

Мини-сервис для обработки заказов: читает события из Kafka (Redpanda), сохраняет их в PostgreSQL и отдаёт через HTTP API и простую веб-страницу.

<p align="center">
  <img src="web/img.png" alt="UI Screenshot" width="600"/>
</p>

---

## Архитектура
- Ingest: Kafka consumer (Redpanda) читает события из `orders` и преобразует DTO в доменную модель.
- Service: `OrderService` пишет заказ в БД в транзакции и обслуживает чтение.
- Storage: PostgreSQL с разнесением по схемам `orders` и `banks`.
- Cache: in-memory кэш по `order_id` с TTL, прогрев из БД за последние 24 часа.
- Transport: HTTP API `GET /api/v1/order/{order_id}` и статическая страница `web/index.html`.

## Почему так
- Kafka/Redpanda: входящие события приходят асинхронно, нужен устойчивый консьюмер.
- Postgres: нормализованные таблицы и транзакционные upsert-операции.
- Кэш: ускорение частых чтений, отдельный прогрев при старте.
- Отдельные схемы `orders` и `banks`: логическое разделение доменов.

## Структура БД
Схемы: `orders`, `banks`.

Таблицы:
- `orders.orders`: основной заказ (`order_id` UUID PK).
- `orders.delivery`: доставка, связь 1:1 по `order_id`.
- `orders.payments`: платеж, связь 1:1 по `order_id`, ссылка на `banks.banks`.
- `orders.items`: товары, уникальные по `rid`.
- `orders.order_items`: связь M:N между заказами и товарами.
- `banks.banks`: справочник банков.

Связи:
- `orders.delivery.order_id` -> `orders.orders.order_id` (1:1).
- `orders.payments.order_id` -> `orders.orders.order_id` (1:1).
- `orders.payments.bank_id` -> `banks.banks.id` (N:1).
- `orders.order_items` -> `orders.orders` и `orders.items` (M:N).

## Интерфейс
- HTTP API: `GET /api/v1/order/{order_id}` возвращает JSON заказа.
- Web UI: `web/index.html` (форма поиска `order_id`, вывод JSON).

## Порты
- `8080` — приложение (HTTP API + статика).
- `8081` — Redpanda Console (веб-интерфейс Kafka).
- `5432` — PostgreSQL.
- `19092` — Kafka внешняя точка (localhost для клиентов).
- `9092` — Kafka внутренняя точка (docker network).

##  Возможности
- Консьюмер Kafka (через [franz-go/kgo](https://github.com/twmb/franz-go))
- PostgreSQL (таблицы `orders`, `delivery`, `payment`, `item`)
- Upsert заказов и связанных сущностей
- Кэш заказов в памяти (по `order_uid`)
- HTTP API для получения заказа
- Простой фронтенд (`web/index.html`)

---

## ⚙️ Запуск

```bash
# собрать и запустить сервисы
docker compose up -d --build

# смотреть логи приложения
docker compose logs -f app

# открыть фронт в браузере
http://localhost:8080
```

## Makefile
Короткие команды-обертки над `docker compose`:

```bash
make up        # db + миграции + kafka + app
make down      # остановка всех сервисов
make logs      # логи app + db
make build     # сборка app
make restart   # перезапуск app + tail логов
```
