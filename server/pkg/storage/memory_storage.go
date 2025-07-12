package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/kokp520/banking-system/server/internal/model"
	"github.com/shopspring/decimal"
)

type MemoryStorage struct {
	accounts  map[uint]*model.Account
	accountID uint
	mutex     sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		accounts:  make(map[uint]*model.Account),
		accountID: 0,
	}
}

func (s *MemoryStorage) CreateAccount(account *model.Account) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.accountID++
	account.ID = s.accountID
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()

	s.accounts[account.ID] = account
	return nil
}

func (s *MemoryStorage) GetAccountByID(id uint) (*model.Account, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	account, exists := s.accounts[id]
	if !exists {
		return nil, errors.New("account not found")
	}

	// Return a copy to avoid external modifications
	accountCopy := *account
	return &accountCopy, nil
}

func (s *MemoryStorage) UpdateAccount(account *model.Account) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.accounts[account.ID]; !exists {
		return errors.New("account not found")
	}

	account.UpdatedAt = time.Now()
	s.accounts[account.ID] = account
	return nil
}

func (s *MemoryStorage) UpdateAccountBalance(id uint, balance decimal.Decimal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	account, exists := s.accounts[id]
	if !exists {
		return errors.New("account not found")
	}

	account.Balance = balance
	account.UpdatedAt = time.Now()
	return nil
}
