package db

import (
	"context"
	"database/sql"
	"net"
	"net/url"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(ctx context.Context) (*sql.DB, error) {
	dsn, err := dsnFromEnv()
	if err != nil {
		return nil, err
	}

	opened, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	opened.SetMaxOpenConns(10)
	opened.SetMaxIdleConns(5)
	opened.SetConnMaxLifetime(30 * time.Minute)
	if err := opened.PingContext(ctx); err != nil {
		_ = opened.Close()
		return nil, err
	}
	return opened, nil
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
