package market

import (
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/moneytrader/pkg"
)

// Market is where you can trade one currency for another.
type Market struct {
	name             string
	baseCurrency     pkg.Currency
	quoteCurrency    pkg.Currency
	minPrice         decimal.Decimal
	maxPrice         decimal.Decimal
	priceIncrement   decimal.Decimal
	minQuantity      decimal.Decimal
	maxQuantity      decimal.Decimal
	quantityStepSize decimal.Decimal
}

func New(name string, baseCur pkg.Currency, quoteCur pkg.Currency,
	minPrice decimal.Decimal, maxPrice decimal.Decimal, priceIncr decimal.Decimal,
	minQty decimal.Decimal, maxQty decimal.Decimal, stepsize decimal.Decimal) pkg.Market {
	market := &Market{
		name: name, baseCurrency: baseCur, quoteCurrency: quoteCur,
		minPrice: minPrice, maxPrice: maxPrice, priceIncrement: priceIncr,
		minQuantity: minQty, maxQuantity: maxQty, quantityStepSize: stepsize,
	}

	return market
}

func (m *Market) Name() string {
	return m.name
}

func (m *Market) BaseCurrency() pkg.Currency {
	return m.baseCurrency
}

func (m *Market) QuoteCurrency() pkg.Currency {
	return m.quoteCurrency
}

func (m *Market) MinPrice() decimal.Decimal {
	return m.minPrice
}

func (m *Market) MaxPrice() decimal.Decimal {
	return m.maxPrice
}

func (m *Market) PriceIncrement() decimal.Decimal {
	return m.priceIncrement
}

func (m *Market) MinQuantity() decimal.Decimal {
	return m.minQuantity
}

func (m *Market) MaxQuantity() decimal.Decimal {
	return m.maxQuantity
}

func (m *Market) QuantityStepSize() decimal.Decimal {
	return m.quantityStepSize
}

func (m *Market) TickerStream(stop <-chan bool) <-chan pkg.Ticker {
	return nil
}

func (m *Market) CandlestickStream(stop <-chan bool, interval string) <-chan pkg.Candlestick {
	return nil
}
