package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/kokp520/banking-system/server/internal/model"
	"github.com/shopspring/decimal"
)

type MemoryStorage struct {
	accounts        map[uint64]*model.Account
	transactions map[uint64]*model.Transaction
	accountID       uint64
	transactionID   uint64
	globalMutex     sync.RWMutex // 鎖accounts map
	accountLocks    sync.Map     // 鎖每隔帳戶, sync.map是原子性
	transactionMutex sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		accounts:        make(map[uint64]*model.Account),
		transactions: make(map[uint64]*model.Transaction),
		accountID:       0,
		transactionID:   0,
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

func (s *MemoryStorage) Transfer(fromID, toID uint64, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("transfer amount must be positive")
	}

	if fromID == toID {
		return errors.New("cannot transfer to the same account")
	}

	var firstLock, secondLock *sync.RWMutex
	var firstID, secondID uint64

	if fromID < toID {
		firstID, secondID = fromID, toID
	} else {
		firstID, secondID = toID, fromID
	}

	firstLock = s.getAccountLock(firstID)
	secondLock = s.getAccountLock(secondID)

	firstLock.RLock()
	secondLock.RLock()

	s.globalMutex.RLock()
	fromAccount, fromExists := s.accounts[fromID]
	toAccount, toExists := s.accounts[toID]
	s.globalMutex.RUnlock()

	if !fromExists {
		firstLock.RUnlock()
		secondLock.RUnlock()
		return errors.New("source account not found")
	}
	if !toExists {
		firstLock.RUnlock()
		secondLock.RUnlock()
		return errors.New("destination account not found")
	}

	if fromAccount.Balance.LessThan(amount) {
		firstLock.RUnlock()
		secondLock.RUnlock()
		return errors.New("insufficient balance")
	}

	// 釋放讀鎖
	firstLock.RUnlock()
	secondLock.RUnlock()

	firstLock.Lock()
	defer firstLock.Unlock()

	secondLock.Lock()
	defer secondLock.Unlock()

	s.globalMutex.RLock()
	fromAccount, fromExists = s.accounts[fromID]
	toAccount, toExists = s.accounts[toID]
	s.globalMutex.RUnlock()

	if !fromExists {
		return errors.New("source account not found")
	}
	if !toExists {
		return errors.New("destination account not found")
	}

	if fromAccount.Balance.LessThan(amount) {
		return errors.New("insufficient balance")
	}

	fromAccount.Balance = fromAccount.Balance.Sub(amount)
	fromAccount.UpdatedAt = time.Now()

	toAccount.Balance = toAccount.Balance.Add(amount)
	toAccount.UpdatedAt = time.Now()

	return nil
}

func (s *MemoryStorage) AddTransaction(transaction *model.Transaction) error {
	s.transactionMutex.Lock()
	defer s.transactionMutex.Unlock()

	s.transactionID++
	transaction.ID = s.transactionID
	s.transactions[transaction.ID] = transaction
	return nil
}

func (s *MemoryStorage) GetTransactionsByAccountID(accountID uint64) ([]*model.Transaction, error) {
	s.transactionMutex.RLock()
	defer s.transactionMutex.RUnlock()

	var transactions []*model.Transaction
	for _, transaction := range s.transactions {
		if transaction.ToAccountID == accountID || (transaction.FromAccountID != nil && *transaction.FromAccountID == accountID) {
			transactionCopy := *transaction
			transactions = append(transactions, &transactionCopy)
		}
	}
	return transactions, nil
}

func (s *MemoryStorage) GetAllTransactions() ([]*model.Transaction, error) {
	s.transactionMutex.RLock()
	defer s.transactionMutex.RUnlock()

	var transactions []*model.Transaction
	for _, transaction := range s.transactions {
		transactionCopy := *transaction
		transactions = append(transactions, &transactionCopy)
	}
	return transactions, nil
}
