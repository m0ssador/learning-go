package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

const DefaultRequestTimeout = 2 * time.Second

func RequestTimeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TimeoutFromEnv() time.Duration {
	value := os.Getenv("REQUEST_TIMEOUT")
	if value == "" {
		return DefaultRequestTimeout
	}
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		log.Printf("REQUEST_TIMEOUT %q не разобран, остаётся %s", value, DefaultRequestTimeout)
		return DefaultRequestTimeout
	}
	return d
}

func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
