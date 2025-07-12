package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	CodeSuccess         = 200
	InvalidParams       = 400
	Unauthorized        = 401
	Forbidden           = 403
	NotFound            = 404
	ServerError         = 500
	UnknownError        = 1000
	InsufficientBalance = 1001
	AccountNotFound     = 1002
	InvalidAmount       = 1003
)

var MsgFlags = map[int]string{
	CodeSuccess:         "success",
	InvalidParams:       "invalid parameters",
	Unauthorized:        "unauthorized",
	Forbidden:           "forbidden",
	NotFound:            "not found",
	ServerError:         "server error",
	UnknownError:        "unknown error",
	InsufficientBalance: "insufficient balance",
	AccountNotFound:     "account not found",
	InvalidAmount:       "invalid amount",
}

func GetMsg(code int) string {
	if msg, ok := MsgFlags[code]; ok {
		return msg
	}
	return MsgFlags[UnknownError]
}

func Result(c *gin.Context, httpCode, code int, data interface{}) {
	c.JSON(httpCode, Response{
		Code:    code,
		Message: GetMsg(code),
		Data:    data,
	})
}

func Success(c *gin.Context, data interface{}) {
	Result(c, http.StatusOK, CodeSuccess, data)
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    InvalidParams,
		Message: message,
		Data:    nil,
	})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    ServerError,
		Message: message,
		Data:    nil,
	})
}
