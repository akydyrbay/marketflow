package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Postgres         PostgresConfig
	Redis            RedisConfig
	Exchanges        []Exchange
	APIAddr          string
	AggregatorWindow time.Duration
	RedisTTL         time.Duration
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type Exchange struct {
	Name    string
	Address string
}

func Load() (*Config, error) {
	requiredEnv := map[string]string{
		"PG_HOST":           os.Getenv("PG_HOST"),
		"PG_PORT":           os.Getenv("PG_PORT"),
		"PG_USER":           os.Getenv("PG_USER"),
		"PG_PASSWORD":       os.Getenv("PG_PASSWORD"),
		"PG_DB":             os.Getenv("PG_DB"),
		"PG_SSLMODE":        os.Getenv("PG_SSLMODE"),
		"REDIS_HOST":        os.Getenv("REDIS_HOST"),
		"REDIS_PORT":        os.Getenv("REDIS_PORT"),
		"REDIS_DB":          os.Getenv("REDIS_DB"),
		"EXCHANGE1_ADDR":    os.Getenv("EXCHANGE1_ADDR"),
		"EXCHANGE2_ADDR":    os.Getenv("EXCHANGE2_ADDR"),
		"EXCHANGE3_ADDR":    os.Getenv("EXCHANGE3_ADDR"),
		"API_ADDR":          os.Getenv("API_ADDR"),
		"AGGREGATOR_WINDOW": os.Getenv("AGGREGATOR_WINDOW"),
		"REDIS_TTL":         os.Getenv("REDIS_TTL"),
	}
	for key, value := range requiredEnv {
		if value == "" {
			return nil, fmt.Errorf("missing required env variable: %s", key)
		}
	}

	pgPort, err := strconv.Atoi(os.Getenv("PG_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid PG_PORT: %w", err)
	}

	redisPort, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT: %w", err)
	}

	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	aggregatorWindow, err := time.ParseDuration(os.Getenv("AGGREGATOR_WINDOW"))
	if err != nil {
		return nil, fmt.Errorf("invalid AGGREGATOR_WINDOW: %w", err)
	}

	redisTTL, err := time.ParseDuration(os.Getenv("REDIS_TTL"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_TTL: %w", err)
	}

	cfg := &Config{
		Postgres: PostgresConfig{
			Host:     os.Getenv("PG_HOST"),
			Port:     pgPort,
			User:     os.Getenv("PG_USER"),
			Password: os.Getenv("PG_PASSWORD"),
			DBName:   os.Getenv("PG_DB"),
			SSLMode:  os.Getenv("PG_SSLMODE"),
		},
		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     redisPort,
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisDB,
		},
		Exchanges: []Exchange{
			{Name: "exchange1", Address: os.Getenv("EXCHANGE1_ADDR")},
			{Name: "exchange2", Address: os.Getenv("EXCHANGE2_ADDR")},
			{Name: "exchange3", Address: os.Getenv("EXCHANGE3_ADDR")},
		},
		APIAddr:          os.Getenv("API_ADDR"),
		AggregatorWindow: aggregatorWindow,
		RedisTTL:         redisTTL,
	}

	return cfg, nil
}
