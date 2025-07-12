package model

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"time"
)

type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypeTransfer TransactionType = "transfer"
)

type Transaction struct {
	ID            uint64            `json:"id"`
	Type          TransactionType   `json:"type"`
	FromAccountID *uint64           `json:"from_account_id"`
	ToAccountID   uint64            `json:"to_account_id"`
	Amount        decimal.Decimal   `json:"amount"`
	Description   string            `json:"description"`
	CreatedAt     time.Time         `json:"created_at"`
	TraceID       string            `json:"trace_id"`
}

func (t Transaction) MarshalJSON() ([]byte, error) {
	type Alias Transaction
	return json.Marshal(&struct {
		Amount string `json:"amount"`
		*Alias
	}{
		Amount: t.Amount.StringFixed(2),
		Alias:  (*Alias)(&t),
	})
}

func NewDeposit(accountID uint64, amount decimal.Decimal, traceID string) *Transaction {
	return &Transaction{
		Type:        TransactionTypeDeposit,
		ToAccountID: accountID,
		Amount:      amount,
		Description: "Deposit to account",
		CreatedAt:   time.Now(),
		TraceID:     traceID,
	}
}

func NewWithdraw(accountID uint64, amount decimal.Decimal, traceID string) *Transaction {
	return &Transaction{
		Type:        TransactionTypeWithdraw,
		ToAccountID: accountID,
		Amount:      amount,
		Description: "Withdraw from account",
		CreatedAt:   time.Now(),
		TraceID:     traceID,
	}
}

func NewTransfer(fromAccountID, toAccountID uint64, amount decimal.Decimal, traceID string) *Transaction {
	return &Transaction{
		Type:          TransactionTypeTransfer,
		FromAccountID: &fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
		Description:   "Transfer between accounts",
		CreatedAt:     time.Now(),
		TraceID:       traceID,
	}
}