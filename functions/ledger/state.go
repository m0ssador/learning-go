package ledger

import (
	"context"
	"log"

	"github.com/m0ssador/learning-go/functions/ledger/internal/cache"
	"github.com/m0ssador/learning-go/functions/ledger/internal/db"
)

func Reset() {
	sqlDB := db.SQL()
	if sqlDB == nil {
		return
	}
	if _, err := sqlDB.Exec(`TRUNCATE TABLE expenses, budgets RESTART IDENTITY`); err != nil {
		log.Printf("reset database: %v", err)
	}

	ctx := context.Background()
	if err := cache.Del(ctx, budgetsCacheKey); err != nil {
		log.Printf("reset cache: %v", err)
	}
	if err := cache.DelByPrefix(ctx, "report:summary:"); err != nil {
		log.Printf("reset report cache: %v", err)
	}
}
