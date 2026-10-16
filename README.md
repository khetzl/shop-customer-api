# shop-customer-api

A Go service that exposes the web shop's catalogue to customers.

## Endpoints

| Method | Path          | Description                                                |
| ------ | ------------- | ---------------------------------------------------------- |
| GET    | `/v1/items`   | Every item in the shop.                                    |
| GET    | `/v1/forYou`  | Items recommended for the caller. Not personalised yet, so for now it returns every item. |

Both respond with `200` and `{"items": [...], "error": null}`, or `500` if the
database can't be read. Items are ordered by id:

```json
{
  "id": "i-tee-1",
  "name": "Tee, limited",
  "parentItemId": "p-tee",
  "category": "tops",
  "description": "Own description",
  "careGuide": "Machine wash at 30°C",
  "material": "cotton",
  "colour": "white",
  "tailoringId": "slim",
  "price": 2500,
  "sizes": ["S", "M", "L"],
  "photoUrls": ["https://example.com/front.jpg"],
  "releaseDate": "2026-10-01",
  "limitedEdition": { "max": 100, "available": 42 },
  "relatedItemIds": ["i-cap-1"],
  "relatedParentItemIds": ["p-cap"]
}
```

`price` is in pennies. `limitedEdition` is left out for items that aren't a
limited edition. Missing text fields are `""`, and missing lists are `[]`.

## Storage

All data is read from **PostgreSQL through the
[`shop-db-models`](https://github.com/khetzl/shop-db-models) library**, using
its generated SQLBoiler models (`pkg/storage/dbmodel`) over a `database/sql`
connection (lib/pq driver). This service does **not** define its own tables;
the schema is owned by `shop-db-models` (`db-scripts/0000_init_up.sql`).

`itemstore` reads the `item_details` view, which already resolves what an item
inherits from its parent item (category, material, colour, sizes, base price,
and the description, care guide and tailoring fallbacks) and lists its photos
and related items.

### Configuration

| Variable       | Default                                                   | Purpose                  |
| -------------- | --------------------------------------------------------- | ------------------------ |
| `DATABASE_URL` | `postgres://madman@localhost:5433/postgres?sslmode=disable` | lib/pq connection string |
| `API_ADDRESS`  | `localhost:8080`                                          | HTTP listen address      |
| `ENVIRONMENT`  | `development`                                             | Deployment environment   |

The default `DATABASE_URL` points at the local database that `shop-db-models`'
Makefile runs (`make init_db`, `make start_db`, then `make push_model` to apply
the schema).

## Running

```bash
go run ./cmd/api
```

## Layout

```
cmd/api/                    entry point
cmd/api/config/             environment configuration
cmd/api/bootstrap/          wires the DB, stores, handler, router and server
internal/api/handler/       HTTP handlers
internal/api/router/        routes and CORS
internal/api/server/        HTTP server lifecycle
internal/domain/models/     API models
internal/services/itemstore/ item persistence via shop-db-models
internal/storage/postgres/  database connection
test/integration/           full-stack HTTP tests
```

## Running the tests

Standard `go test`. From the repo root:

```bash
# Everything
go test ./...

# Just the handler tests (no database needed)
go test ./internal/api/handler

# Just the item store tests
go test ./internal/services/itemstore

# Just the API integration test
go test ./test/integration

# Race detector (recommended)
go test -race ./...
```

### Prerequisite: a PostgreSQL database for the store/integration tests

The item store tests and the integration test run against real
`shop-db-models` persistence, so they need a PostgreSQL database with the
`shop-db-models` schema (`db-scripts/0000_init_up.sql`) applied. Point them at
one with `TEST_DATABASE_URL` (falling back to `DATABASE_URL`). When neither is
set, the database-backed tests **skip** rather than fail, so `go test ./...`
still runs cleanly without a database.

```bash
export TEST_DATABASE_URL='postgres://madman@localhost:5433/shop_test?sslmode=disable'
go test ./...
```

Each test truncates the item tables (everything except `tailoring`) before
running, so **don't point it at a database whose items you want to keep**.

### Prerequisite: private module access

The module depends on the private `github.com/khetzl/shop-db-models` and
`github.com/askthechain/atcutils` packages, so Go needs to fetch them directly
instead of through the public proxy and checksum database:

```bash
export GOPRIVATE=github.com/khetzl,github.com/askthechain

# and make sure git can authenticate to GitHub, e.g.:
git config --global url."ssh://git@github.com/".insteadOf "https://github.com/"
# (or a PAT via ~/.netrc)
```
