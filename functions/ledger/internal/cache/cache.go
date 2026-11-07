package cache

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

func Connect(ctx context.Context) error {
	addr := envOr("REDIS_ADDR", "localhost:6379")
	db, err := redisDB()
	if err != nil {
		return err
	}

	client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		client = nil
		return err
	}

	log.Printf("подключение к Redis установлено: %s db=%d", addr, db)
	return nil
}

func Get(ctx context.Context, key string) (string, bool, error) {
	if client == nil {
		return "", false, nil
	}
	value, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if client == nil {
		return nil
	}
	return client.Set(ctx, key, value, ttl).Err()
}

func Del(ctx context.Context, keys ...string) error {
	if client == nil || len(keys) == 0 {
		return nil
	}
	return client.Del(ctx, keys...).Err()
}

func Close() error {
	if client == nil {
		return nil
	}
	err := client.Close()
	client = nil
	return err
}

func DelByPrefix(ctx context.Context, prefix string) error {
	if client == nil {
		return nil
	}
	var cursor uint64
	for {
		keys, next, err := client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}

func redisDB() (int, error) {
	value := os.Getenv("REDIS_DB")
	if value == "" {
		return 0, nil
	}
	db, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return db, nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
