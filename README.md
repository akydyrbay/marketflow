# marketflow
To load and run images using docker

```bash
docker load -i exchange1_amd64.tar
docker run -p 40101:40101 --name exchange1-amd64 -d exchange1:latest
nc 127.0.0.1 <port>
```
postman request:
```txt
http://localhost:8080/prices/latest/BTCUSDT
```
```graphql
marketflow/
├─ cmd/
│  └─ marketflow/
│     ├── main.go           # ③ bootstraps config & logger
│     └── shutdown.go       # ⑫ signal handling & graceful shutdown
│
├─ internal/
│  ├─ domain/
│  │  ├── models.go         # core structs: PriceUpdate, ModeResult…
│  │  └── ports.go          # interfaces: Exchange, Repository, Cache
│  │
│  ├─ app/
│  │  ├─ aggregation/
│  │  │   └── aggregator.go  # ⑦ reads from Exchange ports, writes to DB/Cache
│  │  │
│  │  ├─ mode/
│  │  │   └── mode.go        # ⑦ computes high/low/average modes
│  │  │
│  │  └─ fan/
│  │      ├── fan.go         # ⑦ fan-out orchestration
│  │      └── workpool.go    # ⑦ worker pool implementation
│
├─ adapters/
│  ├─ exchange/
│  │   └── adapter.go        # ⑥ TCP client → domain.Exchange
│  │
│  ├─ cache/
│  │   └── redis.go          # ⑤ implements domain.Cache on Redis
│  │
│  └─ db/
│      └── postgres.go       # ④ implements domain.Repository on Postgres
│
├─ api/
│  ├── server.go             # ⑩ HTTP server setup & router
│  └── handlers.go           # ⑪ REST endpoints: /prices/latest, /aggregate, /mode, /health
│
└─ pkg/
   └─ logger/
       └── logger.go         # ③ configures Go’s slog logger (level, format, context)
```

Step by step instructions 

```text
2) Configuration & env
Create configs/config.yaml (example) and allow overriding with env vars.
Example configs/config.yaml:

http:
  host: "0.0.0.0"
  port: 8080

postgres:
  host: "postgres"
  port: 5432
  user: "marketflow"
  password: "marketflow"
  dbname: "marketflow"
  sslmode: "disable"

redis:
  host: "redis"
  port: 6379
  password: ""

exchanges:
  - name: "exchange1"
    address: "127.0.0.1:40101"
  - name: "exchange2"
    address: "127.0.0.1:40102"
  - name: "exchange3"
    address: "127.0.0.1:40103"

app:
  db_batch_size: 200
  workers_per_exchange: 5
  redis_ttl_seconds: 70
Load config in main.go (simple YAML -> struct). Print help and exit if --help. Use os.Getenv for overrides.
3) Domain layer (models & ports)
Implement domain types and interfaces in internal/domain.
models.go (sketch):

package domain

import "time"

type PriceUpdate struct {
    Pair     string    // e.g. BTCUSDT
    Exchange string
    Price    float64
    TS       time.Time // timestamp of update
}

type ModeResult struct {
    Pair        string
    Exchange    string
    Timestamp   time.Time
    AvgPrice    float64
    MinPrice    float64
    MaxPrice    float64
}
ports.go (interfaces):
package domain

type Exchange interface {
    Listen(ctx context.Context) (<-chan PriceUpdate, error)
    Name() string
}

type Repository interface {
    InsertAggregates(ctx context.Context, results []ModeResult) error
}

type Cache interface {
    SetLatest(pair, exchange string, p PriceUpdate) error
    GetLatest(pair, exchange string) (PriceUpdate, error)
    AddToSlidingWindow(pair, exchange string, p PriceUpdate) error
    GetWindow(pair, exchange string, since time.Time) ([]PriceUpdate, error)
    Ping() error
}
Note: sliding window operations are implemented on top of Redis lists / sorted sets (see later).
4) Adapters
Postgres adapter adapters/db/postgres.go
Provide NewPostgresRepository(dsn string) domain.Repository.
Implement InsertAggregates to write batch rows.
Ensure batched insert with COPY or INSERT ... VALUES in tx. Keep it simple: INSERT ... VALUES with tx.Prepare and partition batches.
SQL schema (migrations/0001_create_aggregates.sql):
CREATE TABLE IF NOT EXISTS aggregates (
  id SERIAL PRIMARY KEY,
  pair_name TEXT NOT NULL,
  exchange TEXT NOT NULL,
  timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
  average_price DOUBLE PRECISION NOT NULL,
  min_price DOUBLE PRECISION NOT NULL,
  max_price DOUBLE PRECISION NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_aggregates_pair_ts ON aggregates(pair_name, timestamp DESC);
You can run this at startup (execute migration SQL if table not exists).
Redis adapter adapters/cache/redis.go
Implement domain.Cache.
Suggested Redis structures:

HSET latest:{exchange}:{pair} -> fields: price, ts (JSON or fields).
ZADD window:{exchange}:{pair} -> score = unix timestamp, value = price (or JSON for more fields). Keep TTL: set key expiry to 70s after each write.
For range queries: ZRANGEBYSCORE window:key min_ts max_ts WITHSCORES.
Important: if Redis is down, app should still write to Postgres (fallback). Log Redis errors but continue.
Exchange adapter adapters/exchange/tcp_adapter.go
TCP client that connects to provided host:port and reads newline-delimited JSON or lines (use protocol specified by provided programs; test with nc). Each exchange should produce pair,price,timestamp or JSON — parse accordingly.
Implement reconnect logic with exponential backoff. On failure, attempt reconnect every 1s → 2s → ... up to a cap, and log attempts. Provide Listen(ctx) returning a channel of domain.PriceUpdate that closes on ctx.Done().
If you need a simulator for test mode, implement generator below.
5) App layer — aggregator, fan, worker pool, batching
Flow overview
Each Exchange adapter Listen() returns a chan PriceUpdate.
Use a fan-in to combine channels into a single inCh.
Fan-out: a distributor that reads inCh and sends updates to a worker pool.
Worker pool: N workers process updates concurrently:
For each update: write to Redis sliding window, update latest, and enqueue the update into an in-memory buffer keyed by (exchange, pair).
Every 1 minute (ticker), the aggregator computes ModeResult for each (exchange,pair):
Read last 60s from Redis (preferred) or from the in-memory buffer as fallback.
Compute avg, min, max.
Save ModeResult to Postgres via Repository in a batch.
Optionally write the aggregated result to Redis (as latest aggregate).
Batching: collect db_batch_size results or flush every minute.
Worker pool specifics (internal/app/fan/workpool.go)
Parameterize workers_per_exchange from config; for 3 exchanges and 5 workers each, create len(exchanges)*workers_per_exchange workers.
Workers take PriceUpdate tasks and:
call cache.AddToSlidingWindow(...)
call cache.SetLatest(...)
optionally emit processed result to a collector channel for auditing/logging.
Ensure workers respect ctx cancellation.
Aggregator (internal/app/aggr/aggregator.go)
Runs minute ticker.
For each pair+exchange:
Query cache.GetWindow(pair, exchange, since=time.Now().Add(-60s)) to get prices.
Compute avg/min/max.
Append to a []domain.ModeResult.
Push results to repository using Repository.InsertAggregates(ctx, results).
Clear window in Redis older than 60s using ZREMRANGEBYSCORE with -inf to cutoff.
Resilience:
If Redis fails while computing window, fallback to an in-memory buffer (populated by workers).
If DB fails, log error and retry with backoff; but do not crash.
6) Generator for Test Mode
Implement a small CLI or package cmd/generator or tools/generator that:
Emits PriceUpdate lines on TCP ports 40101..40103 to simulate exchanges, or exposes an API for your app to connect.
Pattern: randomly vary price around a base with jitter; send updates at a configurable frequency (e.g., 10–100 msg/s).
Use same message format as live exchange adapters expect.
You can also integrate generator inside main process and switch modes with a command (POST /mode/test) to spawn generator goroutines instead of TCP listeners.
7) API (api/server.go and api/handlers.go)
Implement endpoints listed in README.
Handler responsibilities:

GET /prices/latest/{symbol}: query Redis HGET ALL across all exchanges (or from DB): prefer Redis.
GET /prices/latest/{exchange}/{symbol}: query Redis HGET latest:{exchange}:{symbol}.
GET /prices/highest/{symbol}?period=1m: for the period, for each exchange query Redis ZREVRANGEBYSCORE or DB if longer range. Compute highest.
POST /mode/test and /mode/live: switch inner app/mode state; start/stop generator/listeners accordingly. Return new mode and basic status JSON.
GET /health: return JSON with components status (Postgres ping, Redis ping, number of active exchange connections, uptime).
Remember to validate period param; build helper parseDuration that accepts 1s, 1m, etc.
8) Logging & graceful shutdown
Use pkg/logger/logger.go which configures slog.
All goroutines should accept context.Context. Use the shutdown.go in cmd/marketflow to set up a cancellable context that is canceled on SIGINT / SIGTERM.
On shutdown:
Stop accepting new TCP connections and generator emissions.
Wait for workers to finish processing pending items (use sync.WaitGroup).
Flush any buffered batches to Postgres.
Close DB & Redis connections.
Exit.
9) Docker + docker-compose
Dockerfile (app)
Dockerfile:
FROM golang:1.20-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/marketflow ./cmd/marketflow

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/marketflow /usr/local/bin/marketflow
COPY configs /etc/marketflow
ENV CONFIG_FILE=/etc/marketflow/config.yaml
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/marketflow"]
docker/docker-compose.yml
version: "3.8"
services:
  postgres:
    image: postgres:15
    restart: unless-stopped
    environment:
      POSTGRES_USER: marketflow
      POSTGRES_PASSWORD: marketflow
      POSTGRES_DB: marketflow
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7
    restart: unless-stopped
    ports:
      - "6379:6379"

  app:
    build: ..
    depends_on:
      - postgres
      - redis
    environment:
      - CONFIG_FILE=/etc/marketflow/config.yaml
      - DB_HOST=postgres
      - REDIS_HOST=redis
    ports:
      - "8080:8080"
    restart: on-failure

volumes:
  pgdata:
To include provided exchange images: either docker load them locally and run as described in README, or add them to compose if you have the image names.
Load and run provided images (as README instructs). Example:

docker load -i exchange1_amd64.tar
docker run -p 40101:40101 --name exchange1 -d <image-name>
Run compose:
cd docker
docker-compose up --build
10) Build & run locally
From project root:
# build
go build -o marketflow .

# run using local config
./marketflow --config configs/config.yaml
or with docker-compose:
docker-compose -f docker/docker-compose.yml up --build
11) SQL migrations at startup
Simple approach: at app start open a DB connection and run migrations/0001_create_aggregates.sql. If you prefer a migration tool, add it, but note README mentions only DB/cache external packages allowed, so keep it manual.
Example at startup:

b, _ := os.ReadFile("migrations/0001_create_aggregates.sql")
_, err = db.ExecContext(ctx, string(b))
12) Redis sliding window design (implementation details)
ZADD window:{exchange}:{pair} <unix_ts> "<price>|<ts>" (value can be JSON).
EXPIRE window:{exchange}:{pair} 70 each time you write.
To compute last 60s: ZRANGEBYSCORE window:key (now-60) now WITHSCORES then parse values for min/avg/max.
For latest price: HSET latest:{exchange}:{pair} price <p> ts <ts> and EXPIRE latest:{exchange}:{pair} 70.
This gives O(logN) writes and O(logN + N) reads. Keep windows short to limit size.
13) Testing & health checks
Use curl to test endpoints.
Use nc 127.0.0.1 40101 to confirm exchange simulators are emitting.
Load test generator: create many synthetic updates and ensure workers don't block.
Implement GET /health to return:
{"postgres": "ok", "redis": "ok", "exchanges": {"e1": "connected"}, "mode":"live"}
14) Checklist — what to implement first (author’s suggested incremental approach)
Implement domain models & ports. (fast)
Implement a minimal exchange adapter that reads from a local generator (so you can iterate quickly).
Implement Redis adapter with SetLatest and AddToSlidingWindow.
Implement workers and fan-in/out pipeline to call cache methods.
Implement aggregator to compute averages every minute and a dummy DB repo that logs results (no DB yet).
Add Postgres adapter and migration; wire repository into aggregator and test.
Implement HTTP API endpoints (start with /prices/latest and /health).
Implement TCP exchange adapter and reconnect logic; run with provided docker exchange images.
Harden logging, graceful shutdown, and retry/backoff logic.
Add docker-compose and test end-to-end.

```
