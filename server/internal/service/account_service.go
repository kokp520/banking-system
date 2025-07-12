package service

import (
	"context"
	"github.com/kokp520/banking-system/server/internal/model"
	"github.com/kokp520/banking-system/server/internal/storage"
	"github.com/kokp520/banking-system/server/pkg/logger"
	"github.com/kokp520/banking-system/server/pkg/trace"
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
		logger.WithTraceID(ctx).Error("failed to create account", zap.Error(err), zap.String("name", in.Name))
		return nil, err
	}

	logger.WithTraceID(ctx).Info("account created successfully",
		zap.Uint64("accountId", account.ID),
		zap.String("name", account.Name),
		zap.String("initialBalance", account.Balance.String()),
	)

	return account, nil
}

// GetAccount
// id: accountId
// @Return: model.Account
func (s *AccountService) GetAccount(ctx context.Context, id uint64) (*model.Account, error) {
	return s.storage.GetAccountByID(id)
}

type DepositInput struct {
	Amount decimal.Decimal
}

type WithdrawInput struct {
	Amount decimal.Decimal
}

type TransferInput struct {
	FromAccountID uint64
	ToAccountID   uint64
	Amount        decimal.Decimal
}

// Deposit 存款操作
func (s *AccountService) Deposit(ctx context.Context, id uint64, in DepositInput) error {
	if err := s.storage.Deposit(id, in.Amount); err != nil {
		logger.WithTraceID(ctx).Error("failed to deposit",
			zap.Error(err),
			zap.Uint64("accountId", id),
			zap.String("amount", in.Amount.String()),
		)
		return err
	}

	traceID := trace.GetTraceID(ctx)
	deposit := model.NewDeposit(id, in.Amount, traceID)
	if err := s.storage.AddTransaction(deposit); err != nil {
		logger.WithTraceID(ctx).Error("failed to add deposit transaction",
			zap.Error(err),
			zap.Uint64("accountId", id),
			zap.String("amount", in.Amount.String()),
		)
	}

	logger.WithTraceID(ctx).Info("deposit successful",
		zap.Uint64("accountId", id),
		zap.String("amount", in.Amount.String()),
	)

	return nil
}

// Withdraw 提款操作
func (s *AccountService) Withdraw(ctx context.Context, id uint64, in WithdrawInput) error {
	if err := s.storage.Withdraw(id, in.Amount); err != nil {
		logger.WithTraceID(ctx).Error("failed to withdraw",
			zap.Error(err),
			zap.Uint64("accountId", id),
			zap.String("amount", in.Amount.String()),
		)
		return err
	}

	traceID := trace.GetTraceID(ctx)
	withdraw := model.NewWithdraw(id, in.Amount, traceID)
	if err := s.storage.AddTransaction(withdraw); err != nil {
		logger.WithTraceID(ctx).Error("failed to add withdraw transaction",
			zap.Error(err),
			zap.Uint64("accountId", id),
			zap.String("amount", in.Amount.String()),
		)
	}

	logger.WithTraceID(ctx).Info("withdraw successful",
		zap.Uint64("accountId", id),
		zap.String("amount", in.Amount.String()),
	)

	return nil
}

// Transfer 轉帳操作
func (s *AccountService) Transfer(ctx context.Context, in TransferInput) error {
	if err := s.storage.Transfer(in.FromAccountID, in.ToAccountID, in.Amount); err != nil {
		logger.WithTraceID(ctx).Error("failed to transfer",
			zap.Error(err),
			zap.Uint64("fromAccountId", in.FromAccountID),
			zap.Uint64("toAccountId", in.ToAccountID),
			zap.String("amount", in.Amount.String()),
		)
		return err
	}

	traceID := trace.GetTraceID(ctx)
	transfer := model.NewTransfer(in.FromAccountID, in.ToAccountID, in.Amount, traceID)
	if err := s.storage.AddTransaction(transfer); err != nil {
		logger.WithTraceID(ctx).Error("failed to add transfer transaction",
			zap.Error(err),
			zap.Uint64("fromAccountId", in.FromAccountID),
			zap.Uint64("toAccountId", in.ToAccountID),
			zap.String("amount", in.Amount.String()),
		)
	}

	logger.WithTraceID(ctx).Info("transfer successful",
		zap.Uint64("fromAccountId", in.FromAccountID),
		zap.Uint64("toAccountId", in.ToAccountID),
		zap.String("amount", in.Amount.String()),
	)

	return nil
}

func (s *AccountService) GetTransactions(ctx context.Context, accountID uint64) ([]*model.Transaction, error) {
	transactions, err := s.storage.GetTransactionsByAccountID(accountID)
	if err != nil {
		logger.WithTraceID(ctx).Error("failed to get transactions",
			zap.Error(err),
			zap.Uint64("accountId", accountID),
		)
		return nil, err
	}

	logger.WithTraceID(ctx).Info("transactions retrieved successfully",
		zap.Uint64("accountId", accountID),
		zap.Int("transactionCount", len(transactions)),
	)

	return transactions, nil
}
