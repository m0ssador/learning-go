package app

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/m0ssador/learning-go/functions/ledger/internal/cache"
	"github.com/m0ssador/learning-go/functions/ledger/internal/db"
	"github.com/m0ssador/learning-go/functions/ledger/internal/repository/cached"
	"github.com/m0ssador/learning-go/functions/ledger/internal/repository/pg"
	"github.com/m0ssador/learning-go/functions/ledger/service"
)

var sqlDB *sql.DB

func New(ctx context.Context) (service.Service, func() error, error) {
	opened, err := db.Open(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := cache.Connect(ctx); err != nil {
		_ = opened.Close()
		return nil, nil, err
	}

	sqlDB = opened
	budgets := cached.NewBudgets(pg.NewBudgetRepository(opened))
	expenses := cached.NewExpenses(pg.NewExpenseRepository(opened))
	svc := service.New(budgets, expenses)
	log.Printf("подключение к PostgreSQL установлено")

	closeFn := func() error {
		sqlDB = nil
		return errors.Join(cache.Close(), opened.Close())
	}
	return svc, closeFn, nil
}

func Reset(ctx context.Context) {
	if sqlDB == nil {
		return
	}
	if _, err := sqlDB.ExecContext(ctx, `TRUNCATE TABLE expenses, budgets RESTART IDENTITY`); err != nil {
		log.Printf("reset database: %v", err)
	}
	if err := cached.Clear(ctx); err != nil {
		log.Printf("reset cache: %v", err)
	}
}
