package api

import "github.com/m0ssador/learning-go/functions/ledger/service"

type CreateTransactionRequest struct {
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type TransactionResponse struct {
	ID          int     `json:"id"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type CreateBudgetRequest struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

type BudgetResponse struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period"`
}

func (r CreateTransactionRequest) ToTransaction() service.Transaction {
	return service.Transaction{
		Amount:      r.Amount,
		Category:    r.Category,
		Description: r.Description,
		Date:        r.Date,
	}
}

func TransactionResponseFrom(tx service.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          tx.ID,
		Amount:      tx.Amount,
		Category:    tx.Category,
		Description: tx.Description,
		Date:        tx.Date,
	}
}

func (r CreateBudgetRequest) ToBudget() service.Budget {
	return service.Budget{
		Category: r.Category,
		Limit:    r.Limit,
	}
}

type BulkImportError struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

type BulkImportResponse struct {
	Accepted int               `json:"accepted"`
	Rejected int               `json:"rejected"`
	Errors   []BulkImportError `json:"errors"`
}

func bulkResponseFrom(summary service.ImportSummary) BulkImportResponse {
	errs := make([]BulkImportError, 0, len(summary.Errors))
	for _, item := range summary.Errors {
		errs = append(errs, BulkImportError{Index: item.Index, Error: item.Message})
	}
	return BulkImportResponse{
		Accepted: summary.Accepted,
		Rejected: summary.Rejected,
		Errors:   errs,
	}
}

func BudgetResponseFrom(b service.Budget) BudgetResponse {
	return BudgetResponse{
		Category: b.Category,
		Limit:    b.Limit,
		Period:   b.Period,
	}
}
