package model

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"time"
)

type Account struct {
	ID        uint64          `json:"id"`         // autoincr
	Name      string          `json:"name"`       // 用戶名
	Balance   decimal.Decimal `json:"balance"`    // 餘額
	CreatedAt time.Time       `json:"created_at"` // 創建時間
	UpdatedAt time.Time       `json:"updated_at"` // 最近更新時間
}

func (a Account) MarshalJSON() ([]byte, error) {
	type Alias Account
	return json.Marshal(&struct {
		Balance string `json:"balance"`
		*Alias
	}{
		Balance: a.Balance.StringFixed(2),
		Alias:   (*Alias)(&a),
	})
}
