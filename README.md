# marketflow

```graphql
marketflow/
├─ cmd/
│  └─ marketflow/
│     ├── main.go           # Application entrypoint: bootstraps config, logger, adapters & HTTP server
│     └── flags.go          # Defines CLI flags/subcommands (e.g. “test” vs “live” mode)
│
├─ internal/
│  ├─ domain/
│  │  ├── models.go         # Core data types (PriceUpdate, AggregateRecord, Mode enum)
│  │  └── symbols.go        # Definition of allowed symbols (BTCUSDT, etc.) & validation
│  │
│  ├─ app/
│  │  ├── orchestrator.go   # Coordinates adapters, cache, DB, and scheduler
│  │  ├── processor.go      # Business logic: fan-out/fan-in, filtering, dispatching
│  │  ├── aggregator.go     # Aggregates 60-sec windows, computes min/avg/max
│  │  └── testmode.go       # Synthetic data generator when in “test” mode
│  │
│  └─ adapters/
│     ├─ exchange/
│     │  ├── adapter.go      # ExchangeAdapter struct: connect(), readLoop(), reconnect logic
│     │  └── price_update.go # Unmarshalling into PriceUpdate, symbol-filter helper
│     │
│     ├─ cache/
│     │  └── redis.go        # Redis client wrapper: push updates to sorted sets, eviction
│     │
│     ├─ db/
│     │  └── postgres.go     # Postgres client wrapper: batch insert aggregates, backfill
│     │
│     └─ http/
│        ├── server.go       # HTTP server setup (router, middleware, shutdown)
│        └── handlers.go     # REST handlers: /prices/latest, /aggregate, /mode, /health
│
└─ pkg/
   ├─ config/
   │  ├── loader.go          # Loads YAML/JSON/TOML + environment vars (e.g. via Viper)
   │  └── defaults.go        # Default config values & validation rules
   │
   └─ logger/
      └── logger.go          # Initializes Go’s slog logger with level, format, context tags
```

Step by step instructions 

```text
🛠️ Project Setup & Architecture
1. Initialize the Project
Start a new Go module:

bash
Copy
Edit
go mod init marketflow
Install and use gofumpt as the required code formatter:

bash
Copy
Edit
go install mvdan.cc/gofumpt@latest
2. Directory Layout (Hexagonal/Clean Architecture)
Create the following structure:

bash
Copy
Edit
/cmd/marketflow       → main.go entrypoint, CLI flags, app bootstrap
/internal/app         → application logic: orchestration, aggregation, test-mode
/internal/domain      → core domain models (e.g. PriceUpdate), symbol definitions
/internal/adapters/   → external interface adapters (exchange, db, cache, http)
/pkg/config           → config loader from env/YAML
/pkg/logger           → structured logger setup (Go slog)
3. Configuration Loading
Use Viper or similar to support config in config.yaml, .env, or command-line flags.

Configure PostgreSQL, Redis, exchange ports (40101–40103), and app HTTP port.

🔌 Exchange Integration (Live Mode)
4. Connect to Exchange #1 (port 40101)
Implement a TCP-based Exchange Adapter that connects to the exchange simulator.

Parse each line-delimited JSON message into a PriceUpdate struct.

Only forward the following symbols: BTCUSDT, DOGEUSDT, TONUSDT, SOLUSDT, ETHUSDT.

Reconnect automatically with back-off if the connection drops.

⚙️ Concurrency Pipeline
5. Fan-out / Fan-in Architecture
Fan-out: distribute each valid price update into 5 worker goroutines.

Fan-in: aggregate processed updates into a downstream channel for batching or caching.

🧠 Caching in Redis
6. Redis Integration
Store a rolling 60-second window of updates per symbol (prices:<exchange>:<symbol>).

Use Redis sorted sets with timestamps as scores.

If Redis is down, buffer writes and retry without blocking the data pipeline.

🗃️ Storage in PostgreSQL
7. Aggregation and Database Writes
Define a schema with:

scss
Copy
Edit
(pair_name, exchange, timestamp, average_price, min_price, max_price)
Every 60 seconds:

Fetch the last 60s of data from Redis.

Compute aggregates and insert them into PostgreSQL in batches.

Recover and backfill aggregates after Redis downtime.

🌐 REST API
8. Basic Endpoints
Build a simple HTTP server with endpoints like:

GET /prices/latest/{symbol}

GET /prices/aggregate/{symbol}?period=1m

Read data from Redis (or fallback to PostgreSQL).

Return proper HTTP status codes and error messages.

🧪 Test Mode
9. Synthetic Price Generator
Create a test-mode generator that mimics live updates.

Activate it with:

POST /mode/test (enable generator)

POST /mode/live (switch back to exchange adapters)

➕ Add Remaining Exchanges
10. Run All Sources
Repeat the Exchange Adapter logic for:

40102 → Exchange #2

40103 → Exchange #3

Each gets its own worker pool and feeds into the same aggregation logic.

🔁 Advanced Concurrency Handling
11. Optimize Load Behavior
Tune worker pool sizes and channel buffers.

Handle backpressure (e.g. rate-limiting, discarding old updates).

📡 Full API Suite
12. Add More Features
Extend REST API with:

GET /prices/highest

GET /prices/lowest

GET /prices/average

GET /health

Support query parameters for ?period=, ?exchange=, etc.

⚙️ Config, Logging, and Shutdown
13. Operational Readiness
Parse config file, .env, and CLI flags cleanly.

Integrate structured logging using slog (log levels, key context).

Handle graceful shutdown on SIGINT/SIGTERM.

✅ Final Polish & Testing
14. Prepare for Presentation
Format code with gofumpt.

Add a short USAGE.md or README section with architecture diagrams.

Test live exchange failover (stop container, verify auto-reconnect).

Demonstrate both Live Mode and Test Mode with API calls.
```