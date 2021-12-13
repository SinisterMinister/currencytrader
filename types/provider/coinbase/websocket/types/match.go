package types

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	ord "github.com/sinisterminister/currencytrader/types/order"
)

type Match struct {
	Message
	TradeID        int             `json:"trade_id"`
	ProductID      string          `json:"product_id"`
	Time           time.Time       `json:"time"`
	Sequence       int             `json:"sequence"`
	Price          decimal.Decimal `json:"price"`
	Side           string          `json:"side"`
	Size           decimal.Decimal `json:"size"`
	MakerOrderID   string          `json:"maker_order_id"`
	TakerOrderID   string          `json:"taker_order_id"`
	TakerUserID    string          `json:"taker_user_id"`
	UserID         string          `json:"user_id"`
	TakerProfileID string          `json:"taker_profile_id"`
	ProfileID      string          `json:"profile_id"`
}

func (m *Match) ToDTO(order types.OrderDTO) types.OrderDTO {
	var status types.OrderStatus = ord.Partial
	return types.OrderDTO{
		Market:       order.Market,
		CreationTime: order.CreationTime,
		Filled:       m.Size.Add(order.Filled),
		ID:           order.ID,
		Request:      order.Request,
		Status:       status,
		Fees:         order.Fees,
		FeesSide:     order.FeesSide,
		Paid:         order.Paid,
	}
}
