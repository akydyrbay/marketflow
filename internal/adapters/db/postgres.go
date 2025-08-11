package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	// "marketflow/internal/domain"
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
	logger.Info("closing postgres connection")
	return r.db.Close()
}
