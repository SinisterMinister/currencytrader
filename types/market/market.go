package market

import (
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

// market is where you can trade one currency for another.
type market struct {
	name             string
	baseCurrency     types.Currency
	quoteCurrency    types.Currency
	minPrice         decimal.Decimal
	maxPrice         decimal.Decimal
	priceIncrement   decimal.Decimal
	minQuantity      decimal.Decimal
	maxQuantity      decimal.Decimal
	quantityStepSize decimal.Decimal
	tickerSvc        types.TickerSvc
}

type MarketConfig struct {
	Name             string
	BaseCurrency     types.Currency
	QuoteCurrency    types.Currency
	MinPrice         decimal.Decimal
	MaxPrice         decimal.Decimal
	PriceIncrement   decimal.Decimal
	MinQuantity      decimal.Decimal
	MaxQuantity      decimal.Decimal
	QuantityStepSize decimal.Decimal
	TickerSvc        types.TickerSvc
}

func New(c MarketConfig) types.Market {
	mkt := &market{
		name:             c.Name,
		baseCurrency:     c.BaseCurrency,
		quoteCurrency:    c.QuoteCurrency,
		minPrice:         c.MinPrice,
		maxPrice:         c.MaxPrice,
		priceIncrement:   c.PriceIncrement,
		minQuantity:      c.MinQuantity,
		maxQuantity:      c.MaxQuantity,
		quantityStepSize: c.QuantityStepSize,
		tickerSvc:        c.TickerSvc,
	}

	return mkt
}

func (m *market) Name() string {
	return m.name
}

func (m *market) BaseCurrency() types.Currency {
	return m.baseCurrency
}

func (m *market) QuoteCurrency() types.Currency {
	return m.quoteCurrency
}

func (m *market) MinPrice() decimal.Decimal {
	return m.minPrice
}

func (m *market) MaxPrice() decimal.Decimal {
	return m.maxPrice
}

func (m *market) PriceIncrement() decimal.Decimal {
	return m.priceIncrement
}

func (m *market) MinQuantity() decimal.Decimal {
	return m.minQuantity
}

func (m *market) MaxQuantity() decimal.Decimal {
	return m.maxQuantity
}

func (m *market) QuantityStepSize() decimal.Decimal {
	return m.quantityStepSize
}

func (m *market) Ticker() (types.Ticker, error) {
	return m.tickerSvc.Ticker(m)
}

func (m *market) TickerStream(stop <-chan bool) <-chan types.Ticker {
	return m.tickerSvc.TickerStream(stop, m)
}

func (m *market) CandlestickStream(stop <-chan bool, interval string) <-chan types.Candlestick {
	return nil
}
