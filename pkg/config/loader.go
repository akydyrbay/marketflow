package config

import (
	"errors"
	"os"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type ExchangeConfig struct {
	Ports     []string
	ExchHosts []string
}

func LoadRedisConfig() (*RedisConfig, error) {
	host := os.Getenv("CACHE_HOST")
	port := os.Getenv("CACHE_PORT")
	pass := os.Getenv("CACHE_PASSWORD")

	if host == "" || port == "" || pass == "" {
		return nil, errors.New("CACHE_HOST or CACHE_PORT or CACHE_PASSWORD not set")
	}

	return &RedisConfig{
		Host:     host,
		Port:     port,
		Password: pass,
	}, nil
}

func LoadDBConfig() (*DBConfig, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	if host == "" || port == "" || user == "" || pass == "" || name == "" {
		return nil, errors.New("one or more DB_* env vars are missing")
	}

	return &DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: pass,
		Name:     name,
	}, nil
}

func LoadExchangeConfig() (*ExchangeConfig, error) {
	port1 := os.Getenv("EXCHANGE1_PORT")
	port2 := os.Getenv("EXCHANGE2_PORT")
	port3 := os.Getenv("EXCHANGE3_PORT")

	host1 := os.Getenv("EXCHANGE1_NAME")
	host2 := os.Getenv("EXCHANGE2_NAME")
	host3 := os.Getenv("EXCHANGE3_NAME")

	if port1 == "" || port2 == "" || port3 == "" || host1 == "" || host2 == "" || host3 == "" {
		return nil, errors.New("one or more EXCHANGE_* env vars are missing")
	}

	return &ExchangeConfig{
		Ports:     []string{port1, port2, port3},
		ExchHosts: []string{host1, host2, host3},
	}, nil
}
