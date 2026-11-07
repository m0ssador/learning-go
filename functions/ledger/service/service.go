package service

import (
	"context"
	"errors"
	"time"

	"github.com/m0ssador/learning-go/functions/ledger/internal/domain"
)

type Service interface {
	SetBudget(ctx context.Context, budget domain.Budget) error
	ListBudgets(ctx context.Context) ([]domain.Budget, error)
	AddTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error)
	ListTransactions(ctx context.Context) ([]domain.Transaction, error)
	ReportSummary(ctx context.Context, from, to time.Time) ([]domain.SummaryItem, error)
}

type ledger struct {
	budgets  domain.BudgetRepository
	expenses domain.ExpenseRepository
}

func New(budgets domain.BudgetRepository, expenses domain.ExpenseRepository) Service {
	return &ledger{budgets: budgets, expenses: expenses}
}

func (s *ledger) SetBudget(ctx context.Context, budget domain.Budget) error {
	if err := budget.Validate(); err != nil {
		return err
	}
	return s.budgets.Save(ctx, budget)
}

func (s *ledger) ListBudgets(ctx context.Context) ([]domain.Budget, error) {
	return s.budgets.List(ctx)
}

func (s *ledger) AddTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	if err := tx.Validate(); err != nil {
		return domain.Transaction{}, err
	}
	if _, err := time.Parse(time.DateOnly, tx.Date); err != nil {
		return domain.Transaction{}, errors.New("transaction date must be YYYY-MM-DD")
	}

	budget, err := s.budgets.GetByCategory(ctx, tx.Category)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return domain.Transaction{}, err
	}
	if err == nil {
		spent, err := s.expenses.SumByCategory(ctx, tx.Category)
		if err != nil {
			return domain.Transaction{}, err
		}
		if spent+tx.Amount > budget.Limit {
			return domain.Transaction{}, domain.ErrBudgetExceeded
		}
	}

	return s.expenses.Add(ctx, tx)
}

func (s *ledger) ListTransactions(ctx context.Context) ([]domain.Transaction, error) {
	return s.expenses.List(ctx)
}

func (s *ledger) ReportSummary(ctx context.Context, from, to time.Time) ([]domain.SummaryItem, error) {
	return s.expenses.Summary(ctx, from, to)
}
