package db

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var handle *sql.DB

func Open() error {
	dsn, err := dsnFromEnv()
	if err != nil {
		return err
	}

	opened, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	opened.SetMaxOpenConns(10)
	opened.SetMaxIdleConns(5)
	opened.SetConnMaxLifetime(30 * time.Minute)
	if err := opened.Ping(); err != nil {
		_ = opened.Close()
		return err
	}

	handle = opened
	return nil
}

func SQL() *sql.DB {
	return handle
}

func dsnFromEnv() (string, error) {
	if value := os.Getenv("DATABASE_URL"); value != "" {
		return value, nil
	}

	host := envOr("DB_HOST", "localhost")
	port := envOr("DB_PORT", "5432")
	user := envOr("DB_USER", "postgres")
	pass := envOr("DB_PASS", "postgres")
	name := envOr("DB_NAME", "cashapp")
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, pass),
		Host:     net.JoinHostPort(host, port),
		Path:     name,
		RawQuery: "sslmode=disable",
	}
	return u.String(), nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func Close() error {
	if handle == nil {
		return nil
	}
	err := handle.Close()
	handle = nil
	if err != nil {
		return fmt.Errorf("close postgres: %w", err)
	}
	return nil
}
