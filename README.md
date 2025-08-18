# MarketFlow

## Learning Objectives

- Concurrency and Parallelism
- Concurrency Patterns
- Real-time Data Processing
- Data Caching (Redis)

## Abstract

In this project, you will build a Real-Time Market Data Processing System. This system will be able to process a large volume of incoming data concurrently and efficiently.

Financial markets, especially cryptocurrency exchanges, generate vast amounts of data. Traders rely on real-time data to make decisions. This project will simulate the backend of a system that processes real-time cryptocurrency price updates from multiple sources. However, to ensure the system remains functional without external dependencies, the project will also support a test mode where prices are generated locally.

This project will give you hands-on experience in real-time data ingestion, processing, caching, and storage while ensuring data consistency and performance using Go's concurrency primitives.

## Context

Imagine you're developing a system for a financial firm that requires real-time price updates from multiple cryptocurrency exchanges. The system must efficiently handle concurrent data streams, process updates in real time, and expose an API for querying recent prices and market statistics.

This project is real (a similar project is used at the workplace of one of the graduates). It integrates into a more complex system and addresses some business challenges.

## Project Features

- Real-time data fetching (Live Mode)
- Real-time test data fetching (Test Mode)
- Concurrent data processing using channels & worker pools
- Data storage in PostgreSQL
- Redis caching for quick access to frequently requested prices
- REST API for querying aggregated price data

## System Requirements

- Go 1.23 or later
- Redis
- PostgreSQL

## Setup

1. Clone this repository.

2. **Configure env**:
    - Create your `.env` file to provide PostgreSQL and Redis connection details. Example `.env`:
    ```text
    # Database configs
    DB_HOST=db
    DB_USER=user
    DB_PASSWORD=password
    DB_NAME=marketflow
    DB_PORT=5432               

    # Cache memory configs
    CACHE_HOST=redis
    CACHE_PORT=6379
    CACHE_PASSWORD=superPassword

    # Exchange app configs
    EXCHANGE1_PORT=40101
    EXCHANGE1_NAME=exchange1

    EXCHANGE2_PORT=40102
    EXCHANGE2_NAME=exchange2

    EXCHANGE3_PORT=40103
    EXCHANGE3_NAME=exchange3

    # Aggregator
    AGGREGATOR_WINDOW=1m

    # App config
    APP_PORT=8080
    ```

3. **Running the Provided Programs**:
    - Define the CPU architecture, then load the images:

    ```bash
    docker load -i exchange1_amd64.tar
    docker load -i exchange2_amd64.tar
    docker load -i exchange3_amd64.tar
    ```
    - or use Makefile 

    ```bash
    make load
    ```

4. **Start the Application**:
    - Run the application with docker using Makefile:
    ```bash
    make up
    ```

## API Endpoints

### Market Data API

- `GET /prices/latest/{symbol}` – Get the latest price for a given symbol.
- `GET /prices/latest/{exchange}/{symbol}` – Get the latest price for a given symbol from a specific exchange.
- `GET /prices/highest/{symbol}` – Get the highest price over a period.
- `GET /prices/highest/{exchange}/{symbol}` – Get the highest price over a period from a specific exchange.
- `GET /prices/highest/{symbol}?period={duration}` – Get the highest price within the last {duration}.
- `GET /prices/lowest/{symbol}` – Get the lowest price over a period.
- `GET /prices/lowest/{exchange}/{symbol}` – Get the lowest price over a period from a specific exchange.
- `GET /prices/average/{symbol}` – Get the average price over a period.

### Data Mode API

- `POST /mode/test` – Switch to Test Mode (use generated data).
- `POST /mode/live` – Switch to Live Mode (fetch data from provided programs).

### System Health

- `GET /health` – Returns system status (e.g., connections, Redis availability).

## Data Handling

### Data Storage

- Data is stored in PostgreSQL in the `AggregatedData` table:
    - `pair_name` (string)
    - `exchange` (string)
    - `timestamp` (timestamp)
    - `average_price` (float)
    - `min_price` (float)
    - `max_price` (float)

- Latest price data is cached in Redis for quick access.

### Concurrency Implementation

- **Fan-in**: Aggregating multiple market data streams into a single channel for centralized processing.
- **Fan-out**: Distributing incoming data updates to multiple workers.
- **Worker Pool**: Managing a set of workers to process live updates efficiently.
- **Generator**: Implementing a generator to produce synthetic data for Test Mode.

## Logging

- The application uses Go’s `log/slog` package for logging throughout the application.
- Logs include contextual information such as timestamps and IDs.

## Shutdown

The application implements **graceful shutdown handling** to ensure that resources are cleaned up and the application exits cleanly when receiving a termination signal (e.g., `SIGINT`, `SIGTERM`).

## Configuration

Configuration parameters are read from a `.env` file, including:
- PostgreSQL connection details
- Redis connection details
- Exchange connection details for both live and test modes

