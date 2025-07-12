package service

import (
	"context"
	"github.com/kokp520/banking-system/server/internal/model"
	"github.com/kokp520/banking-system/server/internal/storage"
	"github.com/kokp520/banking-system/server/pkg/logger"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// AccountService
// 可切換成mysql 實作
type AccountService struct {
	storage *storage.MemoryStorage
}

func NewAccountService(storage *storage.MemoryStorage) *AccountService {
	return &AccountService{
		storage: storage,
	}
}

type CreateAccountInput struct {
	Name           string
	InitialBalance decimal.Decimal
}

func (s *AccountService) CreateAccount(ctx context.Context, in CreateAccountInput) (*model.Account, error) {
	account := &model.Account{
		Name:    in.Name,
		Balance: in.InitialBalance,
	}

	if err := s.storage.CreateAccount(account); err != nil {
		logger.Error("failed to create account", zap.Error(err), zap.String("name", in.Name))
		return nil, err
	}

	logger.Info("account created successfully",
		zap.Uint("accountId", account.ID),
		zap.String("name", account.Name),
		zap.String("initialBalance", account.Balance.String()),
	)

	return account, nil
}

// GetAccount
// id: accountId
// @Return: model.Account
func (s *AccountService) GetAccount(ctx context.Context, id uint) (*model.Account, error) {
	return s.storage.GetAccountByID(id)
}

// Deposit
//func (s *AccountService) Deposit() error {
//}

//
// Withdraw
// func (s *AccountService) Withdraw() error {
// }
//
// Transfer
// func (s *AccountService) Transfer() error {
// }
