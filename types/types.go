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
	ToDTO() CurrencyDTO
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
	Ticker() (Ticker, error)
	TickerStream(stop <-chan bool) <-chan Ticker
	ToDTO() MarketDTO
	AttemptOrder(side OrderSide, price decimal.Decimal, quantity decimal.Decimal) (Order, error)
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
	ToDTO() CandlestickDTO
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
	ToDTO() TickerDTO
}

type TickerDTO struct {
	Ask       decimal.Decimal
	Bid       decimal.Decimal
	Price     decimal.Decimal
	Quantity  decimal.Decimal
	Timestamp time.Time
	Volume    decimal.Decimal
}

type Wallet interface {
	Available() decimal.Decimal
	Currency() Currency
	Free() decimal.Decimal
	Locked() decimal.Decimal
	Release(amt decimal.Decimal) error
	Reserve(amt decimal.Decimal) error
	Reserved() decimal.Decimal
	Total() decimal.Decimal
	ToDTO() WalletDTO
}

type WalletDTO struct {
	Currency CurrencyDTO
	Free     decimal.Decimal
	Locked   decimal.Decimal
	Reserved decimal.Decimal
}

// Side represents which side the order will be placed
type OrderSide string

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

// Status handles the various statuses the Order can be in
type OrderStatus string

type Order interface {
	CreationTime() time.Time
	Filled() decimal.Decimal
	ID() string
	Request() OrderRequest
	Status() OrderStatus
	ToDTO() OrderDTO
}

type OrderDTO struct {
	CreationTime time.Time
	Filled       decimal.Decimal
	ID           string
	Request      OrderRequestDTO
	Status       OrderStatus
}

type OrderSvc interface {
	AttemptOrder(market Market, req OrderRequest) (order Order, err error)
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
	GetOrder(id string) (OrderDTO, error)
	AttemptOrder(market MarketDTO, req OrderRequestDTO) (OrderDTO, error)
	CancelOrder(order OrderDTO) error
	GetOrderStream(stop <-chan bool, order OrderDTO) (<-chan OrderDTO, error)
}
