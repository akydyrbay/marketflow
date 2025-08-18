package db

import (
	"database/sql"
	"fmt"
	"log"
	"marketflow/pkg/config"
	"marketflow/pkg/logger"
	"time"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres() *PostgresRepository {
	logger.Info("Starting database connection...")

	dbConfig, err := config.LoadDBConfig()
	if err != nil {
		logger.Error("Error loading DB config", "error", err)
		log.Fatal(err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Name,
	)

	var db *sql.DB

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
	logger.Info("closing postgres connection")
	return r.db.Close()
}
