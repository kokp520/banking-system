package model

import (
	"github.com/shopspring/decimal"
	"time"
)

type Account struct {
	ID        uint            `json:"id"`         // autoincr
	Name      string          `json:"name"`       // 用戶名
	Balance   decimal.Decimal `json:"balance"`    // 餘額
	CreatedAt time.Time       `json:"created_at"` // 創建時間
	UpdatedAt time.Time       `json:"updated_at"` // 最近更新時間
}
