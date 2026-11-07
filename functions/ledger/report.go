package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/m0ssador/learning-go/functions/ledger/internal/cache"
	"github.com/m0ssador/learning-go/functions/ledger/internal/db"
)

const reportCacheTTL = 30 * time.Second

type SummaryItem struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

func GetReportSummary(ctx context.Context, from, to time.Time) ([]SummaryItem, error) {
	sqlDB := db.SQL()
	if sqlDB == nil {
		return nil, errNoDatabase
	}

	key := fmt.Sprintf("report:summary:%s:%s", from.Format(time.DateOnly), to.Format(time.DateOnly))
	if raw, ok, err := cache.Get(ctx, key); err != nil {
		return nil, err
	} else if ok {
		var items []SummaryItem
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			return nil, err
		}
		log.Printf("cache hit %s", key)
		return items, nil
	}

	log.Printf("cache miss %s", key)
	rows, err := sqlDB.QueryContext(ctx, `
		SELECT category, COALESCE(SUM(amount), 0)
		FROM expenses
		WHERE date >= $1 AND date <= $2
		GROUP BY category
		ORDER BY category
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SummaryItem, 0)
	for rows.Next() {
		var item SummaryItem
		if err := rows.Scan(&item.Category, &item.Total); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	raw, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	if err := cache.Set(ctx, key, string(raw), reportCacheTTL); err != nil {
		return nil, err
	}
	return items, nil
}
