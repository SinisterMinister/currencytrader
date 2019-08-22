package types

import (
	"time"

	"github.com/shopspring/decimal"
)

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
	Name() string
	Precision() int
	Symbol() string
	ToDTO() CurrencyDTO
}

type CurrencyDTO struct {
	Name      string
	Precision int
	Symbol    string
}

type Market interface {
	AttemptOrder(side OrderSide, price decimal.Decimal, quantity decimal.Decimal) (Order, error)
	BaseCurrency() Currency
	Candles(interval CandleInterval, periods int) ([]Candle, error)
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
	Filled() decimal.Decimal
	ID() string
	Request() OrderRequest
	Status() OrderStatus
	StatusStream(stop <-chan bool) <-chan OrderStatus
	ToDTO() OrderDTO
}

type OrderDTO struct {
	CreationTime time.Time
	Filled       decimal.Decimal
	ID           string
	Request      OrderRequestDTO
	Status       OrderStatus
}

type OrderRequest interface {
	Price() decimal.Decimal
	Quantity() decimal.Decimal
	Side() OrderSide
	ToDTO() OrderRequestDTO
}

type OrderRequestDTO struct {
	Price    decimal.Decimal
	Quantity decimal.Decimal
	Side     OrderSide
}

// Side represents which side the order will be placed
type OrderSide string

// Status handles the various statuses the Order can be in
type OrderStatus string

type OrderSvc interface {
	AttemptOrder(market Market, req OrderRequest) (order Order, err error)
	CancelOrder(order Order) error
	Order(id string) (Order, error)
}

type Provider interface {
	AttemptOrder(market MarketDTO, req OrderRequestDTO) (OrderDTO, error)
	CancelOrder(order OrderDTO) error
	Candles(interval CandleInterval, periods int) ([]CandleDTO, error)
	Currencies() ([]CurrencyDTO, error)
	Markets() ([]MarketDTO, error)
	Order(id string) (OrderDTO, error)
	OrderStream(stop <-chan bool, order OrderDTO) (<-chan OrderDTO, error)
	Ticker(market MarketDTO) (TickerDTO, error)
	TickerStream(stop <-chan bool, market MarketDTO) (<-chan TickerDTO, error)
	Wallet(currency CurrencyDTO) (WalletDTO, error)
	Wallets() ([]WalletDTO, error)
	WalletStream(stop <-chan bool, currency CurrencyDTO) (<-chan WalletDTO, error)
}

type Trader interface {
	Administerable

	MarketSvc() MarketSvc
	OrderSvc() OrderSvc
	TickerSvc() TickerSvc
	WalletSvc() WalletSvc
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
	Locked   decimal.Decimal
	Reserved decimal.Decimal
}

type WalletSvc interface {
	Wallet(currency Currency) (Wallet, error)
	Wallets() []Wallet
}
