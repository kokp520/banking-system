package storage

import (
	"github.com/kokp520/banking-system/server/internal/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateAccount(t *testing.T) {
	storage := NewMemoryStorage()

	account := &model.Account{
		Name:    "Test",
		Balance: decimal.NewFromFloat(100),
	}

	err := storage.CreateAccount(account)

	assert.NoError(t, err)
	assert.Equal(t, uint64(1), account.ID)
	assert.False(t, account.CreatedAt.IsZero())
	assert.False(t, account.UpdatedAt.IsZero())

	retrieve, err := storage.GetAccountByID(1)
	assert.NoError(t, err)
	assert.Equal(t, account.Name, retrieve.Name)
	assert.True(t, account.Balance.Equal(retrieve.Balance))
}

func TestCreateMultipleAccounts(t *testing.T) {
	storage := NewMemoryStorage()

	account1 := &model.Account{Name: "test 1", Balance: decimal.NewFromFloat(100)}
	account2 := &model.Account{Name: "test 2", Balance: decimal.NewFromFloat(200)}

	err1 := storage.CreateAccount(account1)
	err2 := storage.CreateAccount(account2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, uint64(1), account1.ID)
	assert.Equal(t, uint64(2), account2.ID)
}

func TestGetAccountByID(t *testing.T) {
	storage := NewMemoryStorage()

	// 測試不存在account
	_, err := storage.GetAccountByID(999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account not found")

	account := &model.Account{
		Name:    "test",
		Balance: decimal.NewFromFloat(150.5),
	}
	storage.CreateAccount(account)

	retrievedAccount, err := storage.GetAccountByID(account.ID)
	assert.NoError(t, err)
	assert.Equal(t, account.Name, retrievedAccount.Name)
	assert.True(t, account.Balance.Equal(retrievedAccount.Balance))

	// 這邊原本漏掉需要特別測試這塊，需要返回copy，修改不影響原始數據
	// update memory
	retrievedAccount.Balance = decimal.NewFromFloat(999.9)
	// 重新獲取
	originalAccount, _ := storage.GetAccountByID(account.ID)
	// assert
	assert.True(t, account.Balance.Equal(originalAccount.Balance))
}

func TestDeposit(t *testing.T) {
	storage := NewMemoryStorage()

	account := &model.Account{
		Name:    "test",
		Balance: decimal.NewFromFloat(100),
	}
	storage.CreateAccount(account)

	// 測試正常存款
	err := storage.Deposit(account.ID, decimal.NewFromFloat(50))
	assert.NoError(t, err)

	updatedAccount, _ := storage.GetAccountByID(account.ID)
	expected := decimal.NewFromFloat(150)
	assert.True(t, expected.Equal(updatedAccount.Balance))

	// 測試負數存款 or 0
	err = storage.Deposit(account.ID, decimal.NewFromFloat(-10))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deposit amount cannot be negative")

	err = storage.Deposit(account.ID, decimal.Zero)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deposit amount cannot be negative")

	// 不存在account
	err = storage.Deposit(999, decimal.NewFromFloat(10.0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account not found")
}

func TestWithdraw(t *testing.T) {
	storage := NewMemoryStorage()

	account := &model.Account{
		Name:    "test",
		Balance: decimal.NewFromFloat(100),
	}
	storage.CreateAccount(account)

	// 測試正常withdraw
	err := storage.Withdraw(account.ID, decimal.NewFromFloat(30))
	assert.NoError(t, err)

	updatedAccount, _ := storage.GetAccountByID(account.ID)
	expected := decimal.NewFromFloat(70)
	assert.True(t, expected.Equal(updatedAccount.Balance))

	// balance < req
	err = storage.Withdraw(account.ID, decimal.NewFromFloat(100.0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")

	// 負數提款
	err = storage.Withdraw(account.ID, decimal.NewFromFloat(-10.0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "withdraw amount cannot be negative")

	// 不存在account
	err = storage.Withdraw(999, decimal.NewFromFloat(10.0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account not found")
}

func TestTransfer(t *testing.T) {
	storage := NewMemoryStorage()

	fromAccount := &model.Account{Name: "from user", Balance: decimal.NewFromFloat(100)}
	toAccount := &model.Account{Name: "to user", Balance: decimal.NewFromFloat(50)}

	storage.CreateAccount(fromAccount)
	storage.CreateAccount(toAccount)

	err := storage.Transfer(fromAccount.ID, toAccount.ID, decimal.NewFromFloat(30.0))
	assert.NoError(t, err)

	updatedFromAccount, _ := storage.GetAccountByID(fromAccount.ID)
	updatedToAccount, _ := storage.GetAccountByID(toAccount.ID)

	expectedFrom := decimal.NewFromFloat(70)
	expectedTo := decimal.NewFromFloat(80)

	assert.True(t, expectedFrom.Equal(updatedFromAccount.Balance))
	assert.True(t, expectedTo.Equal(updatedToAccount.Balance))

	// 餘額不足
	err = storage.Transfer(fromAccount.ID, toAccount.ID, decimal.NewFromFloat(100))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")

	// 轉給自己
	err = storage.Transfer(fromAccount.ID, fromAccount.ID, decimal.NewFromFloat(10.0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot transfer to the same account")

	// 負數轉帳
	err = storage.Transfer(fromAccount.ID, toAccount.ID, decimal.NewFromFloat(-10.0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transfer amount must be positive")

	// 測試不存在from test
	err = storage.Transfer(999, toAccount.ID, decimal.NewFromFloat(10.0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source account not found")

	// 測試不存在的目標帳戶
	err = storage.Transfer(fromAccount.ID, 999, decimal.NewFromFloat(10.0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "destination account not found")
}

func TestTransferDeadlockPrevention(t *testing.T) {
	storage := NewMemoryStorage()

	// 創建兩個帳戶
	account1 := &model.Account{Name: "User 1", Balance: decimal.NewFromFloat(100.0)}
	account2 := &model.Account{Name: "User 2", Balance: decimal.NewFromFloat(100.0)}

	storage.CreateAccount(account1)
	storage.CreateAccount(account2)

	// check順序性
	err1 := storage.Transfer(account1.ID, account2.ID, decimal.NewFromFloat(10.0))
	err2 := storage.Transfer(account2.ID, account1.ID, decimal.NewFromFloat(5.0))

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	finalAccount1, _ := storage.GetAccountByID(account1.ID)
	finalAccount2, _ := storage.GetAccountByID(account2.ID)

	expected1 := decimal.NewFromFloat(95.0)  // 100 - 10 + 5
	expected2 := decimal.NewFromFloat(105.0) // 100 + 10 - 5

	assert.True(t, expected1.Equal(finalAccount1.Balance), "account1 != 95")
	assert.True(t, expected2.Equal(finalAccount2.Balance), "account2 != 105")
}

func TestGetAccountLock(t *testing.T) {
	storage := NewMemoryStorage()

	// 測試獲取鎖
	lock1a := storage.getAccountLock(1)
	lock1b := storage.getAccountLock(1)
	lock2 := storage.getAccountLock(2)

	assert.Same(t, lock1a, lock1b, "Same account should return same lock, lock1a != lock1b")
	assert.NotSame(t, lock1a, lock2, "Different accounts should return different locks, lock2 == lock1")
}
