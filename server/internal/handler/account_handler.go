package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kokp520/banking-system/server/internal/service"
	"github.com/kokp520/banking-system/server/pkg/response"
	"github.com/shopspring/decimal"
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
	ID uint `uri:"id" binging:"required"`
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

	// id定義為uint
	//id, err := strconv.ParseUint(req.ID, 10, 32)
	//if err != nil {
	//	response.BadRequest(c, "parse uint failed", err)
	//}

	account, err := h.accountService.GetAccount(c.Request.Context(), req.ID)
	if err != nil {
		response.Error(c, response.AccountNotFound)
		return
	}

	response.Success(c, account)
}
