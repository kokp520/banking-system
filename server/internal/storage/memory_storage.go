package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/kokp520/banking-system/server/internal/model"
	"github.com/shopspring/decimal"
)

type MemoryStorage struct {
	accounts     map[uint64]*model.Account
	accountID    uint64
	globalMutex  sync.RWMutex // 鎖accounts map
	accountLocks sync.Map     // 鎖每隔帳戶, sync.map是原子性
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		accounts:  make(map[uint64]*model.Account),
		accountID: 0,
	}
}

// getAccountLock
func (s *MemoryStorage) getAccountLock(accountID uint64) *sync.RWMutex {
	// sync.map是原子 避免同時對map做操作(get/update)造成race condition
	value, _ := s.accountLocks.LoadOrStore(accountID, &sync.RWMutex{})
	return value.(*sync.RWMutex)
}

func (s *MemoryStorage) CreateAccount(account *model.Account) error {
	s.globalMutex.Lock()
	defer s.globalMutex.Unlock()

	s.accountID++
	account.ID = s.accountID
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()

	s.accounts[account.ID] = account
	return nil
}

func (s *MemoryStorage) GetAccountByID(id uint64) (*model.Account, error) {
	// 使用帳戶級別的讀鎖
	accountLock := s.getAccountLock(id)
	accountLock.RLock()
	defer accountLock.RUnlock()

	s.globalMutex.RLock()
	account, exists := s.accounts[id]
	s.globalMutex.RUnlock()

	if !exists {
		return nil, errors.New("account not found")
	}

	// Return a copy to avoid external modifications
	accountCopy := *account
	return &accountCopy, nil
}

// Deposit
// 如果使用MySQL，這個操作可以用事務包裝：BEGIN; UPDATE accounts SET balance = balance + ? WHERE id = ?; COMMIT;
func (s *MemoryStorage) Deposit(id uint64, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("deposit amount cannot be negative")
	}

	accountLock := s.getAccountLock(id)
	accountLock.Lock()
	defer accountLock.Unlock()

	s.globalMutex.RLock()
	account, exists := s.accounts[id]
	s.globalMutex.RUnlock()

	if !exists {
		return errors.New("account not found")
	}

	account.Balance = account.Balance.Add(amount)
	account.UpdatedAt = time.Now()
	return nil
}

// Withdraw
// 如果使用MySQL，這個操作可以用事務包裝：BEGIN; UPDATE accounts SET balance = balance - ? WHERE id = ? AND balance >= ?; COMMIT;
func (s *MemoryStorage) Withdraw(id uint64, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("withdraw amount cannot be negative")
	}

	accountLock := s.getAccountLock(id)
	accountLock.Lock()
	defer accountLock.Unlock()

	s.globalMutex.RLock()
	account, exists := s.accounts[id]
	s.globalMutex.RUnlock()

	if !exists {
		return errors.New("account not found")
	}

	if account.Balance.LessThan(amount) {
		return errors.New("insufficient balance")
	}

	account.Balance = account.Balance.Sub(amount)
	account.UpdatedAt = time.Now()
	return nil
}
