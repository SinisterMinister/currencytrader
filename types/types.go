package types

import (
	"time"

	"github.com/shopspring/decimal"
)

type AccountSvc interface {
	Currencies() ([]Currency, error)
	Currency(name string) (Currency, error)
	Fees() (Fees, error)
	Wallet(currency Currency) (Wallet, error)
	Wallets() ([]Wallet, error)
}

type Administerable interface {
	Start()
	Stop()
}

type Candle interface {
	Close() decimal.Decimal
	High() decimal.Decimal
	Low() decimal.Decimal
	Open() decimal.Decimal
	Timestamp() time.Time
	ToDTO() CandleDTO
	Volume() decimal.Decimal
}

type CandleDTO struct {
	Close     decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Open      decimal.Decimal
	Timestamp time.Time
	Volume    decimal.Decimal
}

type CandleInterval string

// Currency TODO
type Currency interface {
	Increment() decimal.Decimal
	Name() string
	Precision() int
	Symbol() string
	ToDTO() CurrencyDTO
	Wallet() Wallet
}

type CurrencyDTO struct {
	Increment decimal.Decimal
	Name      string
	Precision int
	Symbol    string
}

type Fees interface {
	MakerRate() decimal.Decimal
	TakerRate() decimal.Decimal
	ToDTO() FeesDTO
	Volume() decimal.Decimal
}

type FeesDTO struct {
	MakerRate decimal.Decimal
	TakerRate decimal.Decimal
	Volume    decimal.Decimal
}

type Market interface {
	AttemptOrder(req OrderRequest) (Order, error)
	AverageTradeVolume() (decimal.Decimal, error)
	BaseCurrency() Currency
	Candles(interval CandleInterval, start time.Time, end time.Time) ([]Candle, error)
	MaxPrice() decimal.Decimal
	MaxQuantity() decimal.Decimal
	MinPrice() decimal.Decimal
	MinQuantity() decimal.Decimal
	Name() string
	PriceIncrement() decimal.Decimal
	QuantityStepSize() decimal.Decimal
	QuoteCurrency() Currency
	Ticker() (Ticker, error)
	TickerStream(stop <-chan bool) <-chan Ticker
	ToDTO() MarketDTO
}

type MarketDTO struct {
	Name             string
	BaseCurrency     CurrencyDTO
	QuoteCurrency    CurrencyDTO
	MinPrice         decimal.Decimal
	MaxPrice         decimal.Decimal
	PriceIncrement   decimal.Decimal
	MinQuantity      decimal.Decimal
	MaxQuantity      decimal.Decimal
	QuantityStepSize decimal.Decimal
}

type MarketSvc interface {
	Market(cur0 Currency, cur1 Currency) (Market, error)
	Markets() []Market
}

type Order interface {
	CreationTime() time.Time
	Done() <-chan bool
	Fees() (OrderSide, decimal.Decimal)
	Filled() decimal.Decimal
	ID() string
	IsDone() bool
	Market() Market
	Paid() decimal.Decimal
	Request() OrderRequest
	Status() OrderStatus
	StatusStream(stop <-chan bool) <-chan OrderStatus
	ToDTO() OrderDTO
}

type OrderDTO struct {
	Market       MarketDTO
	CreationTime time.Time       `json:"creationTime"`
	Fees         decimal.Decimal `json:"fees"`
	FeesSide     OrderSide       `json:"feesSide"`
	Filled       decimal.Decimal `json:"filled"`
	ID           string          `json:"id"`
	Paid         decimal.Decimal `json:"paid"`
	Request      OrderRequestDTO
	Status       OrderStatus `json:"status"`
}

type OrderRequest interface {
	Market() Market
	Price() decimal.Decimal
	Quantity() decimal.Decimal
	Side() OrderSide
	ToDTO() OrderRequestDTO
	Type() OrderType
}

type OrderRequestDTO struct {
	Price    decimal.Decimal `json:"price"`
	Quantity decimal.Decimal `json:"quantity"`
	Side     OrderSide       `json:"side"`
	Type     OrderType       `json:"type"`
	Market   MarketDTO
}

// Side represents which side the order will be placed
type OrderSide string

// Status handles the various statuses the Order can be in
type OrderStatus string

type OrderType string

type OrderSvc interface {
	AttemptOrder(m Market, t OrderType, s OrderSide, price decimal.Decimal, quantity decimal.Decimal) (order Order, err error)
	CancelOrder(order Order) error
	Order(m Market, id string) (Order, error)
}

type Provider interface {
	AttemptOrder(req OrderRequestDTO) (OrderDTO, error)
	AverageTradeVolume(mkt MarketDTO) (decimal.Decimal, error)
	CancelOrder(order OrderDTO) error
	Candles(mkt MarketDTO, interval CandleInterval, start time.Time, end time.Time) ([]CandleDTO, error)
	Currencies() ([]CurrencyDTO, error)
	Fees() (FeesDTO, error)
	Markets() ([]MarketDTO, error)
	Order(markest MarketDTO, id string) (OrderDTO, error)
	OrderStream(stop <-chan bool, order OrderDTO) (<-chan OrderDTO, error)
	Ticker(market MarketDTO) (TickerDTO, error)
	TickerStream(stop <-chan bool, market MarketDTO) (<-chan TickerDTO, error)
	Wallet(currency CurrencyDTO) (WalletDTO, error)
	Wallets() ([]WalletDTO, error)
}

type Trader interface {
	Administerable

	AccountSvc() AccountSvc
	MarketSvc() MarketSvc
	OrderSvc() OrderSvc
	TickerSvc() TickerSvc
}

type Ticker interface {
	Ask() decimal.Decimal
	Bid() decimal.Decimal
	Price() decimal.Decimal
	Quantity() decimal.Decimal
	Timestamp() time.Time
	ToDTO() TickerDTO
	Volume() decimal.Decimal
}

type TickerDTO struct {
	Ask       decimal.Decimal
	Bid       decimal.Decimal
	Price     decimal.Decimal
	Quantity  decimal.Decimal
	Timestamp time.Time
	Volume    decimal.Decimal
}

type TickerSvc interface {
	Ticker(market Market) (Ticker, error)
	TickerStream(stop <-chan bool, market Market) <-chan Ticker
}

type Wallet interface {
	Available() decimal.Decimal
	Currency() Currency
	Free() decimal.Decimal
	ID() string
	Locked() decimal.Decimal
	Release(amt decimal.Decimal) error
	Reserve(amt decimal.Decimal) error
	Reserved() decimal.Decimal
	ToDTO() WalletDTO
	Total() decimal.Decimal
}

type WalletDTO struct {
	Currency CurrencyDTO
	Free     decimal.Decimal
	ID       string
	Locked   decimal.Decimal
	Reserved decimal.Decimal
}
