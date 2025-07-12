package service

import (
	"context"
	"github.com/kokp520/banking-system/server/internal/model"
	"github.com/kokp520/banking-system/server/pkg/logger"
	"github.com/kokp520/banking-system/server/pkg/storage"
	"go.uber.org/zap"
)

type AccountService struct {
	storage *storage.MemoryStorage
}

func NewAccountService(storage *storage.MemoryStorage) *AccountService {
	return &AccountService{
		storage: storage,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, req *model.CreateAccountRequest) (*model.Account, error) {
	account := &model.Account{
		Name:    req.Name,
		Balance: req.InitialBalance,
	}

	if err := s.storage.CreateAccount(account); err != nil {
		logger.Error("failed to create account", zap.Error(err), zap.String("name", req.Name))
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
// func (s *AccountService) Deposit() error {
// }
//
// Withdraw
// func (s *AccountService) Withdraw() error {
// }
//
// Transfer
// func (s *AccountService) Transfer() error {
// }
