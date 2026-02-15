DC = docker compose

db-up:
	$(DC) up -d db

migrate: db-up
	$(DC) run --rm migrate

kafka-up:
	$(DC) up -d redpanda
	$(DC) up -d redpanda-init
	$(DC) up -d redpanda-console

app-up: db-up migrate kafka-up
	$(DC) up -d app

build:
	$(DC) build app

restart:
	$(DC) restart app
	$(DC) logs -f app

up: db-up migrate kafka-up app-up

down:
	$(DC) down

logs:
	$(DC) logs -f app db

lint:
	golangci-lint run ./...
