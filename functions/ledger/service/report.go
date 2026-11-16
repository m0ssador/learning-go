package service

import (
	"context"
	"log"
	"sync"
	"time"
)

const reportHeartbeat = 400 * time.Millisecond

type categorySum struct {
	category string
	total    float64
	err      error
}

func (s *ledger) ReportSummary(ctx context.Context, from, to time.Time) (map[string]float64, error) {
	stop := startReportHeartbeat(ctx)
	defer stop()

	categories, err := s.expenses.Categories(ctx, from, to)
	if err != nil {
		return nil, err
	}

	totals := make([]categorySum, len(categories))
	var wg sync.WaitGroup
	for i, category := range categories {
		wg.Add(1)
		go func(i int, category string) {
			defer wg.Done()
			if err := ctx.Err(); err != nil {
				totals[i] = categorySum{err: err}
				return
			}
			total, err := s.expenses.SumBetween(ctx, category, from, to)
			totals[i] = categorySum{category: category, total: total, err: err}
		}(i, category)
	}
	wg.Wait()

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := make(map[string]float64, len(totals))
	for _, item := range totals {
		if item.err != nil {
			return nil, item.err
		}
		out[item.category] = item.total
	}
	return out, nil
}

func startReportHeartbeat(ctx context.Context) func() {
	ticker := time.NewTicker(reportHeartbeat)
	done := make(chan struct{})
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Printf("идёт расчёт отчёта")
			case <-ctx.Done():
				log.Printf("расчёт отчёта остановлен: %v", ctx.Err())
				return
			case <-done:
				if err := ctx.Err(); err != nil {
					log.Printf("расчёт отчёта остановлен: %v", err)
					return
				}
				log.Printf("расчёт отчёта завершён")
				return
			}
		}
	}()
	return func() { close(done) }
}
