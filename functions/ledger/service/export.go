package service

import "github.com/m0ssador/learning-go/functions/ledger/internal/domain"

type (
	Budget          = domain.Budget
	Transaction     = domain.Transaction
	SummaryItem     = domain.SummaryItem
	ValidationError = domain.ValidationError
)

var ErrBudgetExceeded = domain.ErrBudgetExceeded
