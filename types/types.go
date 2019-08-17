package types

import (
	"time"

	"github.com/shopspring/decimal"
)

// Currency TODO
type Currency interface {
	Name() string
	Precision() int
	Symbol() string
}

type CurrencyDTO struct {
	Name      string
	Precision int
	Symbol    string
}

type Market interface {
	BaseCurrency() Currency
	CandlestickStream(stop <-chan bool, interval string) <-chan Candlestick
	MaxPrice() decimal.Decimal
	MaxQuantity() decimal.Decimal
	MinPrice() decimal.Decimal
	MinQuantity() decimal.Decimal
	Name() string
	PriceIncrement() decimal.Decimal
	QuantityStepSize() decimal.Decimal
	QuoteCurrency() Currency
	TickerStream(stop <-chan bool) <-chan Ticker
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

type Candlestick interface {
	Close() decimal.Decimal
	High() decimal.Decimal
	Low() decimal.Decimal
	Open() decimal.Decimal
	Timestamp() time.Time
	Volume() decimal.Decimal
}

type CandlestickDTO struct {
	Close     decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Open      decimal.Decimal
	Timestamp time.Time
	Volume    decimal.Decimal
}

type Ticker interface {
	Ask() decimal.Decimal
	Bid() decimal.Decimal
	Price() decimal.Decimal
	Quantity() decimal.Decimal
	Timestamp() time.Time
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

// Wallet TODO
type Wallet interface {
	Available() decimal.Decimal
	Currency() Currency
	Free() decimal.Decimal
	Locked() decimal.Decimal
	Release(amt decimal.Decimal) error
	Reserve(amt decimal.Decimal) error
	Reserved() decimal.Decimal
	Total() decimal.Decimal
}

type WalletDTO struct {
	Currency CurrencyDTO
	Free     decimal.Decimal
	Locked   decimal.Decimal
	Reserved decimal.Decimal
}

// Side represents which side the order will be placed
type Side int

const (
	// BuySide represents a buy sided order
	BuySide Side = iota

	// SellSide represents a sell sided order
	SellSide
)

type OrderRequest interface {
	Price() decimal.Decimal
	Quantity() decimal.Decimal
	Side() Side
}

type OrderRequestDTO struct {
	Price    decimal.Decimal
	Quantity decimal.Decimal
	Side     Side
}

// Status handles the various statuses the Order can be in
type Status int

const (
	// Pending is for orders still working to be fulfilled
	Pending Status = iota

	// Canceled is for orders that have been cancelled
	Canceled

	// Success is for orders that have succefully filled
	Success
)

type Order interface {
	CreationTime() time.Time
	Filled() decimal.Decimal
	ID() string
	Request() OrderRequest
	Status() Status
}

type OrderDTO struct {
	CreationTime time.Time
	Filled       decimal.Decimal
	ID           string
	Request      OrderRequest
	Status       Status
}

type OrderSvc interface {
	AttemtOrder(req OrderRequest) Order
	CancelOrder(order Order) error
	GetOrder(id string) (Order, error)
}

type Administerable interface {
	Start()
	Stop()
}

type WalletSvc interface {
	GetWallet(currency Currency) (Wallet, error)
	GetWallets() []Wallet
}

type MarketSvc interface {
	GetMarket(cur0 Currency, cur1 Currency) (Market, error)
	GetMarkets() []Market
}

type TickerSvc interface {
	Ticker(market Market) (Ticker, error)
	TickerStream(stop <-chan bool, market Market) <-chan Ticker
}

type Trader interface {
	Administerable
	OrderSvc() OrderSvc
	WalletSvc() WalletSvc
	MarketSvc() MarketSvc
	TickerSvc() TickerSvc
}

type Provider interface {
	GetMarkets() ([]MarketDTO, error)
	GetCurrencies() ([]CurrencyDTO, error)
	GetWallets() ([]WalletDTO, error)
	GetWallet(currency CurrencyDTO) (WalletDTO, error)
	GetWalletStream(stop <-chan bool, currency CurrencyDTO) (<-chan WalletDTO, error)
	GetTicker(market MarketDTO) (TickerDTO, error)
	GetTickerStream(stop <-chan bool, market MarketDTO) (<-chan TickerDTO, error)
}
