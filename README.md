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
http://localhost:8081
