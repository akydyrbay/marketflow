package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"marketflow/internal/domain"
	"marketflow/pkg/logger"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres() (*PostgresRepository, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		logger.Error("failed to ping postgres", "error", err)
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	logger.Info("postgres connection established")
	return &PostgresRepository{db: db}, nil
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
