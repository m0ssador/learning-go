package service

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
)

const DefaultImportWorkers = 4

type ImportFailure struct {
	Index   int
	Message string
}

type ImportSummary struct {
	Accepted int
	Rejected int
	Errors   []ImportFailure
}

type importJob struct {
	index int
	tx    Transaction
}

type importResult struct {
	index int
	err   error
}

func (s *ledger) ImportTransactions(ctx context.Context, txs []Transaction, workers int) (ImportSummary, error) {
	if workers < 1 {
		workers = DefaultImportWorkers
	}

	jobs := make(chan importJob)
	results := make(chan importResult)

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.importWorker(ctx, jobs, results)
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	collected := make(chan ImportSummary, 1)
	go func() {
		collected <- collectImportResults(results)
	}()

	go func() {
		defer close(jobs)
		for i, tx := range txs {
			select {
			case <-ctx.Done():
				return
			case jobs <- importJob{index: i, tx: tx}:
			}
		}
	}()

	summary := <-collected
	if err := ctx.Err(); err != nil {
		return summary, err
	}
	return summary, nil
}

func (s *ledger) importWorker(ctx context.Context, jobs <-chan importJob, results chan<- importResult) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, err := s.AddTransaction(ctx, job.tx)
			results <- importResult{index: job.index, err: err}
		}
	}
}

func collectImportResults(results <-chan importResult) ImportSummary {
	var accepted atomic.Int64
	var rejected atomic.Int64
	failures := make([]ImportFailure, 0)
	for res := range results {
		if res.err != nil {
			rejected.Add(1)
			failures = append(failures, ImportFailure{Index: res.index, Message: res.err.Error()})
			continue
		}
		accepted.Add(1)
	}
	sort.Slice(failures, func(i, j int) bool {
		return failures[i].Index < failures[j].Index
	})
	return ImportSummary{
		Accepted: int(accepted.Load()),
		Rejected: int(rejected.Load()),
		Errors:   failures,
	}
}
