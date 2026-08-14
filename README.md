# Go Warehouse System

A small e-commerce backend built as a set of independent Go microservices,
communicating over HTTP, gRPC and Kafka. Built as a portfolio project to
practice production-style Go: clean architecture, mocked unit tests, an
OpenAPI contract, event-driven order notifications, and a one-command
Docker environment.

## Tech stack

- **Go 1.26**, [chi](https://github.com/go-chi/chi) router
- **PostgreSQL** (via `pgx/v5`) for persistence
- **gRPC** for synchronous service-to-service calls (Order → Warehouse)
- **Kafka** (via `segmentio/kafka-go`) for asynchronous order events
- **JWT** (HS256) for authentication
- **golang-migrate** for schema migrations
- **Swagger / OpenAPI 3** for API documentation
- **Docker Compose** for local orchestration

## Architecture

Four independent services, each with its own `main.go`, sharing one
PostgreSQL database (separate tables) and one Kafka topic:

```
                        ┌─────────────┐
                        │   Client    │
                        └──────┬──────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                 │
              ▼                ▼                 ▼
        ┌───────────┐   ┌────────────┐   ┌───────────────┐
        │   auth    │   │   order    │   │  swagger UI    │
        │  :8081    │   │  :8082     │   │   :8085        │
        └─────┬─────┘   └──┬──────┬──┘   └────────────────┘
              │            │      │
              │ (JWT,      │gRPC  │ produces
              │ same       │      │ order_events
              │ secret)    ▼      ▼
              │       ┌──────────┐  ┌───────────┐
              │       │warehouse │  │   kafka    │
              │       │ :50051   │  │  :29092    │
              │       └────┬─────┘  └─────┬─────┘
              │            │              │ consumes
              ▼            ▼              ▼
        ┌─────────────────────────┐  ┌───────────────┐
        │        postgres         │  │ notification   │
        │        :5433            │  │    :8083       │
        └─────────────────────────┘  └───────────────┘
```

- **auth** — registration, login, JWT issuing/refresh.
- **order** — validates the caller's JWT locally (shared `JWT_SECRET`, no
  network call to `auth`), reserves stock via gRPC to **warehouse**, saves
  the order, and publishes an `order_events` message to Kafka.
- **warehouse** — owns stock and exposes it over gRPC only.
- **notification** — consumes `order_events` from Kafka and logs order
  updates (stand-in for a real notification channel, e.g. email/push).
- **swagger** — serves the OpenAPI spec and Swagger UI for the HTTP APIs.

## Running it

**Requirements:** Docker and Docker Compose.

```bash
cp .env.example .env
# edit .env if you want non-default credentials/secrets
docker compose up -d --build
# or: make up
```

This will:
1. Start Postgres and Kafka and wait for them to become healthy.
2. Run all pending SQL migrations in a one-off `migrate` container.
3. Build and start `auth`, `order`, `warehouse`, `notification`, and the
   Swagger UI.

Check everything is up:

```bash
docker compose ps
docker compose logs -f notification   # watch order events arrive
```

Stop everything:

```bash
docker compose down       # keep the DB volume
docker compose down -v    # also wipe Postgres data
```

### Services & ports

| Service      | Protocol | Port  |
|--------------|----------|-------|
| auth         | HTTP     | 8081  |
| order        | HTTP     | 8082  |
| notification | HTTP     | 8083  |
| kafka-ui     | HTTP     | 8080  |
| swagger      | HTTP     | 8085  |
| warehouse    | gRPC     | 50051 |
| postgres     | TCP      | 5433  |
| kafka        | TCP      | 29092 |

Swagger UI: **http://localhost:8085/swagger**

### Running locally without Docker

Each service also runs directly on the host (see `Makefile` targets
`run-auth`, `run-order`, `run-warehouse`, `run-notification`,
`run-swagger`), reading configuration from a local `.env` file with
`localhost`-based defaults. Start Postgres/Kafka with
`docker compose up -d postgres kafka` first, then run migrations with
`make migrate-up`.

## API walkthrough

```bash
# 1. Register a user
curl -s -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "P@ssw0rd!"}'
# -> {"AccessToken": "...", "RefreshToken": "..."}

# 2. Log in (same payload shape)
curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "P@ssw0rd!"}'

TOKEN="<paste AccessToken here>"

# 3. Create an order (requires the JWT from step 1/2)
curl -s -X POST http://localhost:8082/api/v1/orders/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"items": [{"sku": "SKU-001", "quantity": 2}]}'
```

Then check the notification service's logs — it should show the
`order_events` message being consumed:

```bash
docker compose logs -f notification
```

Full request/response schemas are in the Swagger UI at
`http://localhost:8085/swagger`, backed by `api/openapi/openapi.yaml`.

## Configuration

All services read configuration from environment variables (see
`.env.example`). In Docker Compose these are wired to the in-network
service names automatically; the `.env` values are only the defaults used
for local, non-Docker runs and for seeding the Postgres container.

| Variable              | Used by                     | Docker Compose value                |
|------------------------|------------------------------|----------------------------------------|
| `DATABASE_URL`          | auth, order, warehouse       | `postgres://...@postgres:5432/...`     |
| `JWT_SECRET`            | auth, order                  | shared secret from `.env`              |
| `KAFKA_BROKERS`         | order, notification          | `kafka:9092`                           |
| `WAREHOUSE_GRPC_ADDR`   | order                        | `warehouse:50051`                      |

## Testing

```bash
make test
# or
go test ./...
```

Unit tests use mocked repositories/clients to verify business logic in
isolation, including the order status transitions (`PENDING` →
`PLACED`/`FAILED`).

## Project layout

```
internal/
  auth/          # registration, login, JWT issuance
  order/         # order creation, gRPC client to warehouse, Kafka producer
  warehouse/     # stock, gRPC server
  notification/  # Kafka consumer
pkg/             # shared packages (jwt, kafka, logger, http helpers, ...)
api/
  openapi/       # OpenAPI spec + Swagger UI server
  proto/         # warehouse.proto (gRPC contract)
migrations/      # golang-migrate SQL migrations
docker/          # per-service Dockerfiles
docker-compose.yml
```
