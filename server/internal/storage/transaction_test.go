package storage

import (
	"github.com/kokp520/banking-system/server/internal/model"
	"github.com/shopspring/decimal"
	"testing"
)

// 追加tx
func TestAddTransaction(t *testing.T) {
	storage := NewMemoryStorage()

	transaction := model.NewDeposit(1, decimal.NewFromInt(100), "trace-123")

	err := storage.AddTransaction(transaction)
	if err != nil {
		t.Fatalf("Failed to add transaction: %v", err)
	}

	if transaction.ID == 0 {
		t.Error("Transaction ID should be assigned")
	}

	if transaction.ID != 1 {
		t.Errorf("Expected transaction ID 1, got %d", transaction.ID)
	}
}

// get tx
func TestGetTransactionsByAccountID(t *testing.T) {
	storage := NewMemoryStorage()

	deposit := model.NewDeposit(1, decimal.NewFromInt(100), "trace-1")
	withdraw := model.NewWithdraw(1, decimal.NewFromInt(50), "trace-2")
	transfer1 := model.NewTransfer(1, 2, decimal.NewFromInt(25), "trace-3")
	transfer2 := model.NewTransfer(2, 1, decimal.NewFromInt(10), "trace-4")
	unrelated := model.NewDeposit(3, decimal.NewFromInt(200), "trace-5")

	storage.AddTransaction(deposit)
	storage.AddTransaction(withdraw)
	storage.AddTransaction(transfer1)
	storage.AddTransaction(transfer2)
	storage.AddTransaction(unrelated)

	transactions, err := storage.GetTransactionsByAccountID(1)
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}

	expectedCount := 4
	if len(transactions) != expectedCount {
		t.Errorf("Expected %d transactions for account 1, got %d", expectedCount, len(transactions))
	}

	foundTypes := make(map[model.TransactionType]int)
	for _, transaction := range transactions {
		foundTypes[transaction.Type]++

		// 不包含id: 1  error
		if transaction.ToAccountID != 1 &&
			(transaction.FromAccountID == nil || *transaction.FromAccountID != 1) {
			t.Errorf("Transaction %d should involve account 1", transaction.ID)
		}
	}

	// 檢查type
	if foundTypes[model.TransactionTypeDeposit] != 1 ||
		foundTypes[model.TransactionTypeWithdraw] != 1 ||
		foundTypes[model.TransactionTypeTransfer] != 2 {
		t.Errorf("Unexpected transaction type distribution: %v", foundTypes)
	}
}
