package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kokp520/banking-system/server/internal/service"
	"github.com/kokp520/banking-system/server/pkg/response"
	"github.com/shopspring/decimal"
	"strconv"
)

type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

// REQ

type CreateAccountRequest struct {
	Name           string          `json:"name" binding:"required"`
	InitialBalance decimal.Decimal `json:"initial_balance"`
}

type GetAccountRequest struct {
	ID uint64 `uri:"id" binging:"required"`
}

type DepositRequest struct {
	Amount decimal.Decimal `json:"amount" binding:"required"`
}

type WithdrawRequest struct {
	Amount decimal.Decimal `json:"amount" binding:"required"`
}

type TransferRequest struct {
	ToAccountID uint64          `json:"to_account_id" binding:"required"`
	Amount      decimal.Decimal `json:"amount" binding:"required"`
}

// API

// CreateAccount 創建帳戶 API
// @Summary 創建銀行帳戶
// @Description 創建一個新的銀行帳戶，初始餘額不能為負數
// @Tags accounts
// @Accept json
// @Produce json
// @Param account body model.CreateAccountRequest true "帳戶信息"
// @Success 200 {object} model.Account
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/accounts [post]
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req CreateAccountRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	account, err := h.accountService.CreateAccount(c.Request.Context(), service.CreateAccountInput{
		Name:           req.Name,
		InitialBalance: req.InitialBalance,
	})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, account)
}

func (h *AccountHandler) GetAccount(c *gin.Context) {
	var req GetAccountRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	//id := uint64(req.ID)
	// id定義為uint64
	//id, err := strconv.ParseUint(req.ID, 10, 64)
	//if err != nil {
	//	response.BadRequest(c, "parse uint failed", err)
	//}

	account, err := h.accountService.GetAccount(c.Request.Context(), req.ID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, account)
}

func (h *AccountHandler) Deposit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		response.BadRequest(c, "amount must be greater than 0")
		return
	}

	err = h.accountService.Deposit(c.Request.Context(), id, service.DepositInput{Amount: req.Amount})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "deposit successful"})
}

// Withdraw 提款 API
// @Summary 從帳戶提款
// @Description 從指定帳戶提款，需要檢查餘額是否足夠
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path uint64 true "帳戶ID"
// @Param withdraw body WithdrawRequest true "提款信息"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/accounts/{id}/withdraw [post]
func (h *AccountHandler) Withdraw(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 手動驗證金額必須大於0
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		response.BadRequest(c, "amount must be greater than 0")
		return
	}

	err = h.accountService.Withdraw(c.Request.Context(), id, service.WithdrawInput{Amount: req.Amount})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "withdraw successful"})
}

// Transfer 轉帳 API
// @Summary 轉帳
// @Description 從一個帳戶轉帳到另一個帳戶
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path uint64 true "來源帳戶ID"
// @Param transfer body TransferRequest true "轉帳信息"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/accounts/{id}/transfer [post]
func (h *AccountHandler) Transfer(c *gin.Context) {
	fromIDStr := c.Param("id")
	fromID, err := strconv.ParseUint(fromIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid from account id")
		return
	}

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 手動驗證金額必須大於0
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		response.BadRequest(c, "amount must be greater than 0")
		return
	}

	// 檢查不能轉給自己
	if fromID == req.ToAccountID {
		response.BadRequest(c, "cannot transfer to the same account")
		return
	}

	err = h.accountService.Transfer(c.Request.Context(), service.TransferInput{
		FromAccountID: fromID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message":      "transfer successful",
		"from_account": fromID,
		"to_account":   req.ToAccountID,
		"amount":       req.Amount.String(),
	})
}

func (h *AccountHandler) GetTransactions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid account id")
		return
	}

	transactions, err := h.accountService.GetTransactions(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, transactions)
}
