package types

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	ord "github.com/sinisterminister/currencytrader/types/order"
)

type Received struct {
	Message
	ProductID     string          `json:"product_id"`
	Time          time.Time       `json:"time"`
	Sequence      int             `json:"sequence"`
	OrderID       string          `json:"order_id"`
	Size          decimal.Decimal `json:"size"`
	Funds         decimal.Decimal `json:"funds"`
	Price         decimal.Decimal `json:"price"`
	Side          string          `json:"side"`
	OrderType     string          `json:"order_type"`
	ClientOrderID string          `json:"client_oid"`
}

func (r *Received) ToDTO(order types.OrderDTO) types.OrderDTO {
	return types.OrderDTO{
		Market:       order.Market,
		CreationTime: r.Time,
		Filled:       decimal.Zero,
		ID:           order.ID,
		Request:      order.Request,
		Status:       ord.Pending,
		Fees:         order.Fees,
		FeesSide:     order.FeesSide,
		Paid:         order.Paid,
	}
}
