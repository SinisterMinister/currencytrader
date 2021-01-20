package coinbase

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	ord "github.com/sinisterminister/currencytrader/types/order"
)

type CallRequest struct {
	Args         []interface{}
	Method       string
	ResponseChan chan []interface{}
}

type MessageHandler interface {
	Name() string
	Input() chan<- DataPackage
}

type Message struct {
	Type string `json:"type"`
}

type DataPackage struct {
	Message
	Data []byte
}

type Heartbeat struct {
	Message
	Sequence    int       `json:"sequence"`
	LastTradeID int       `json:"last_trade_id"`
	ProductID   string    `json:"product_id"`
	Time        time.Time `json:"time"`
}

type AuthenticatedSubscribe struct {
	Signature  string `json:"signature,omitempty"`
	Key        string `json:"key,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`
}

type GlobalSubscribe struct {
	Message
	AuthenticatedSubscribe
	ProductIDs []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}

type Subscribe struct {
	Message
	AuthenticatedSubscribe
	Channels []struct {
		Name       string   `json:"name"`
		ProductIDs []string `json:"product_ids"`
	} `json:"channels"`
}

type Subscriptions struct {
	Message
	Channels []struct {
		Name       string   `json:"name"`
		ProductIDs []string `json:"product_ids"`
	} `json:"channels"`
}

type Status struct {
	Message
	Products []struct {
		ID             string          `json:"id"`
		BaseCurrency   string          `json:"base_currency"`
		QuoteCurrency  string          `json:"quote_currency"`
		BaseMinSize    decimal.Decimal `json:"base_min_size"`
		BaseMaxSize    decimal.Decimal `json:"base_max_size"`
		BaseIncrement  decimal.Decimal `json:"base_increment"`
		QuoteIncrement decimal.Decimal `json:"quote_increment"`
		DisplayName    string          `json:"display_name"`
		Status         string          `json:"status"`
		StatusMessage  string          `json:"status_message"`
		MinMarketFunds decimal.Decimal `json:"min_market_funds"`
		MaxMarketFunds decimal.Decimal `json:"max_market_funds"`
		PostOnly       bool            `json:"post_only"`
		LimitOnly      bool            `json:"limit_only"`
		CancelOnly     bool            `json:"cancel_only"`
	} `json:"products"`
	Currencies []struct {
		ID            string                 `json:"id"`
		Name          string                 `json:"name"`
		MinSize       string                 `json:"min_size"`
		Status        string                 `json:"status"`
		StatusMessage string                 `json:"status_message"`
		MaxPrecision  decimal.Decimal        `json:"max_precision"`
		ConvertableTo []string               `json:"convertible_to"`
		Details       map[string]interface{} `json:"details"`
	} `json:"currencies"`
}

type Ticker struct {
	Message
	TradeID   int             `json:"trade_id"`
	Sequence  int             `json:"sequence"`
	Time      time.Time       `json:"time"`
	ProductID string          `json:"product_id"`
	Price     decimal.Decimal `json:"price"`
	Side      string          `json:"side"`
	LastSize  decimal.Decimal `json:"last_size"`
	BestBid   decimal.Decimal `json:"best_bid"`
	BestAsk   decimal.Decimal `json:"best_ask"`
}

type Snapshot struct {
	Message
	ProductID string              `json:"product_id"`
	Bids      [][]decimal.Decimal `json:"bids"`
	Asks      [][]decimal.Decimal `json:"asks"`
}

type Level2Update struct {
	Message
	ProductID string     `json:"product_id"`
	Time      time.Time  `json:"time"`
	Changes   [][]string `json:"changes"`
}

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

type Open struct {
	Message
	ProductID     string          `json:"product_id"`
	Time          time.Time       `json:"time"`
	Sequence      int             `json:"sequence"`
	OrderID       string          `json:"order_id"`
	RemainingSize decimal.Decimal `json:"remaining_size"`
	Price         decimal.Decimal `json:"price"`
	Side          string          `json:"side"`
}

func (o *Open) ToDTO(order types.OrderDTO) types.OrderDTO {
	var status types.OrderStatus
	if order.Request.Quantity.Equal(o.RemainingSize) {
		status = ord.Pending
	} else {
		status = ord.Partial
	}
	return types.OrderDTO{
		Market:       order.Market,
		CreationTime: order.CreationTime,
		Filled:       decimal.Zero, // We fill zero here since a match event will cover the actual amount(s)
		ID:           order.ID,
		Request:      order.Request,
		Status:       status,
		Fees:         order.Fees,
		FeesSide:     order.FeesSide,
		Paid:         order.Paid,
	}
}

type Done struct {
	Message
	ProductID     string          `json:"product_id"`
	Time          time.Time       `json:"time"`
	Sequence      int             `json:"sequence"`
	OrderID       string          `json:"order_id"`
	RemainingSize decimal.Decimal `json:"remaining_size"`
	Price         decimal.Decimal `json:"price"`
	Side          string          `json:"side"`
	Reason        string          `json:"reason"`
}

func (d *Done) ToDTO(order types.OrderDTO) types.OrderDTO {
	var status types.OrderStatus
	switch d.Reason {
	case "filled":
		status = ord.Filled
	case "canceled":
		status = ord.Canceled
	}
	return types.OrderDTO{
		Market:       order.Market,
		CreationTime: order.CreationTime,
		Filled:       order.Request.Quantity.Sub(d.RemainingSize),
		ID:           order.ID,
		Request:      order.Request,
		Status:       status,
		Fees:         order.Fees,
		FeesSide:     order.FeesSide,
		Paid:         order.Paid,
	}
}

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
	var status types.OrderStatus
	status = ord.Partial
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

type Change struct {
	Message
	ProductID string          `json:"product_id"`
	Time      time.Time       `json:"time"`
	Sequence  int             `json:"sequence"`
	OrderID   string          `json:"order_id"`
	Price     decimal.Decimal `json:"price"`
	Side      string          `json:"side"`
	OldFunds  decimal.Decimal `json:"old_funds"`
	NewFunds  decimal.Decimal `json:"new_funds"`
	OldSize   decimal.Decimal `json:"old_size"`
	NewSize   decimal.Decimal `json:"new_size"`
}

func (c *Change) ToDTO(order types.OrderDTO) types.OrderDTO {
	request := order.Request
	request.Price = c.Price
	request.Quantity = c.NewSize
	return types.OrderDTO{
		Market:       order.Market,
		CreationTime: order.CreationTime,
		Filled:       order.Filled,
		ID:           order.ID,
		Request:      request,
		Status:       ord.Updated,
		Fees:         order.Fees,
		FeesSide:     order.FeesSide,
		Paid:         order.Paid,
	}
}

type Activate struct {
	Message
	ProductID    string          `json:"product_id"`
	Timestamp    string          `json:"timestamp"`
	UserID       string          `json:"user_id"`
	ProfileID    string          `json:"profile_id"`
	OrderID      string          `json:"order_id"`
	StopType     string          `json:"stop_type"`
	StopPrice    decimal.Decimal `json:"stop_price"`
	Side         string          `json:"side"`
	Size         decimal.Decimal `json:"size"`
	Funds        decimal.Decimal `json:"funds"`
	TakerFeeRate decimal.Decimal `json:"taker_fee_rate"`
	Private      bool            `json:"private"`
}
