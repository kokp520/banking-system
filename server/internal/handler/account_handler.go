package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kokp520/banking-system/server/internal/model"
	"github.com/kokp520/banking-system/server/internal/service"
	"github.com/kokp520/banking-system/server/pkg/response"
)

type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

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
	var req model.CreateAccountRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	account, err := h.accountService.CreateAccount(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, account)
}
