package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"marketflow/internal/domain"
	"marketflow/pkg/logger"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres() *PostgresRepository {
	logger.Info("Starting database connection...")

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	if host == "" || port == "" || user == "" || pass == "" || name == "" {
		logger.Error("one or more DB_* env vars are missing",
			"DB_HOST", host,
			"DB_PORT", port,
			"DB_USER", user,
			"DB_NAME", name,
		)
		log.Fatal("unable to continue without DB config")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name,
	)

	var db *sql.DB
	var err error
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			logger.Warn("failed to open database", "attempt", i+1, "error", err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = db.Ping()
		if err == nil {
			break
		}

		logger.Warn("failed to ping database", "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		logger.Error("failed to connect after retries", "error", err)
		log.Fatal(err)
	}
	if err := createTables(db); err != nil {
		logger.Error("failed to create tables", "error", err)
		log.Fatal(err)
	}
	logger.Info("postgres connection established")
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Close() error {
	logger.Info("closing postgres")
	return r.db.Close()
}

func (r *PostgresRepository) GetLatest(ctx context.Context, exchange, pair string) (domain.PriceStats, error) {
	query := `
		SELECT pair_name, exchange, timestamp, average_price, min_price, max_price
		FROM price_stats
		WHERE pair_name = $1 AND exchange = $2
		ORDER BY timestamp DESC
		LIMIT 1
	`
	var avgStr, minStr, maxStr string
	var stats domain.PriceStats
	stats.Average, _ = strconv.ParseFloat(avgStr, 64)
	stats.Min, _ = strconv.ParseFloat(minStr, 64)
	stats.Max, _ = strconv.ParseFloat(maxStr, 64)
	err := r.db.QueryRowContext(ctx, query, pair, exchange).Scan(
		&stats.Pair, &stats.Exchange, &stats.Timestamp, &stats.Average, &stats.Min, &stats.Max,
	)
	if err == sql.ErrNoRows {
		logger.Warn("no latest price found", "pair", pair, "exchange", exchange)
		return domain.PriceStats{}, fmt.Errorf("no latest price for %s:%s", exchange, pair)
	}
	if err != nil {
		logger.Error("failed to get latest price", "pair", pair, "exchange", exchange, "error", err)
		return domain.PriceStats{}, fmt.Errorf("failed to get latest price: %w", err)
	}

	logger.Info("got latest price", "pair", pair, "exchange", exchange, "price", stats.Average)
	return stats, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS price_stats (
            id SERIAL PRIMARY KEY,
            pair_name VARCHAR(20) NOT NULL,
            exchange VARCHAR(50) NOT NULL,
            timestamp TIMESTAMP NOT NULL,
            average_price DECIMAL(24,8) NOT NULL,
            min_price DECIMAL(24,8) NOT NULL,
            max_price DECIMAL(24,8) NOT NULL,
            UNIQUE(pair_name, exchange, timestamp)
        );
        
        CREATE INDEX IF NOT EXISTS idx_pair_timestamp ON price_stats(pair_name, timestamp);
        CREATE INDEX IF NOT EXISTS idx_exchange_pair_timestamp ON price_stats(exchange, pair_name, timestamp);
    `)
	logger.Info("Table is created")
	return err
}

func (r *PostgresRepository) StoreStatsBatch(ctx context.Context, stats []domain.PriceStats) error {
	if len(stats) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO price_stats (pair_name, exchange, timestamp, average_price, min_price, max_price)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (pair_name, exchange, timestamp) DO UPDATE
		SET average_price = EXCLUDED.average_price,
			min_price = EXCLUDED.min_price,
			max_price = EXCLUDED.max_price
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, stat := range stats {
		_, err := stmt.ExecContext(ctx,
			stat.Pair,
			stat.Exchange,
			stat.Timestamp,
			stat.Average,
			stat.Min,
			stat.Max,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) StorePriceUpdate(ctx context.Context, update domain.PriceUpdate) error {
	// This will be called by the cache layer
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO price_updates (pair, exchange, price, timestamp)
		VALUES ($1, $2, $3, $4)
	`, update.Pair, update.Exchange, update.Price, update.Time)
	return err
}
