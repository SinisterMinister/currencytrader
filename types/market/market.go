package market

import (
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/currency"
	"github.com/sinisterminister/currencytrader/types/internal"
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

	trader internal.Trader
}

func (m *market) ToDTO() types.MarketDTO {
	return types.MarketDTO{
		Name:             m.Name(),
		BaseCurrency:     m.BaseCurrency().ToDTO(),
		QuoteCurrency:    m.QuoteCurrency().ToDTO(),
		MinPrice:         m.MinPrice(),
		MaxPrice:         m.MaxPrice(),
		PriceIncrement:   m.PriceIncrement(),
		MinQuantity:      m.MinQuantity(),
		MaxQuantity:      m.MaxQuantity(),
		QuantityStepSize: m.QuantityStepSize(),
	}
}

func New(trader internal.Trader, m types.MarketDTO) types.Market {
	mkt := &market{
		name:             m.Name,
		baseCurrency:     currency.New(m.BaseCurrency),
		quoteCurrency:    currency.New(m.QuoteCurrency),
		minPrice:         m.MinPrice,
		maxPrice:         m.MaxPrice,
		priceIncrement:   m.PriceIncrement,
		minQuantity:      m.MinQuantity,
		maxQuantity:      m.MaxQuantity,
		quantityStepSize: m.QuantityStepSize,
		trader:           trader,
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
	return m.trader.TickerSvc().Ticker(m)
}

func (m *market) TickerStream(stop <-chan bool) <-chan types.Ticker {
	return m.trader.TickerSvc().TickerStream(stop, m)
}

func (m *market) CandlestickStream(stop <-chan bool, interval string) <-chan types.Candlestick {
	return nil
}

func (m *market) AttemptOrder(req types.OrderRequest) (types.Order, error) {
	return m.trader.OrderSvc().AttemptOrder(m, req)
}
