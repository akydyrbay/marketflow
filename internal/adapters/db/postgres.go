package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
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
	var stats domain.PriceStats
	err := r.db.QueryRowContext(ctx, query, pair, exchange).Scan(
		&stats.Pair, &stats.Exchange, &stats.Timestamp, &stats.AveragePrice, &stats.MinPrice, &stats.MaxPrice,
	)
	if err == sql.ErrNoRows {
		logger.Warn("no latest price found", "pair", pair, "exchange", exchange)
		return domain.PriceStats{}, fmt.Errorf("no latest price for %s:%s", exchange, pair)
	}
	if err != nil {
		logger.Error("failed to get latest price", "pair", pair, "exchange", exchange, "error", err)
		return domain.PriceStats{}, fmt.Errorf("failed to get latest price: %w", err)
	}

	logger.Info("got latest price", "pair", pair, "exchange", exchange, "price", stats.AveragePrice)
	return stats, nil
}
