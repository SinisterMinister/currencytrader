package market

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/candle"
	"github.com/sinisterminister/currencytrader/types/currency"
	"github.com/sinisterminister/currencytrader/types/internal"
)

// market is where you can trade one currency for another.
type market struct {
	dto    types.MarketDTO
	trader internal.Trader
}

func (m *market) ToDTO() types.MarketDTO {
	return m.dto
}

func New(trader internal.Trader, m types.MarketDTO) types.Market {
	mkt := &market{
		dto:    m,
		trader: trader,
	}

	return mkt
}

func (m *market) Name() string { return m.dto.Name }

func (m *market) BaseCurrency() types.Currency { return currency.New(m.trader, m.dto.BaseCurrency) }

func (m *market) QuoteCurrency() types.Currency { return currency.New(m.trader, m.dto.QuoteCurrency) }

func (m *market) MinPrice() decimal.Decimal { return m.dto.MinPrice }

func (m *market) MaxPrice() decimal.Decimal { return m.dto.MaxPrice }

func (m *market) PriceIncrement() decimal.Decimal { return m.dto.PriceIncrement }

func (m *market) MinQuantity() decimal.Decimal { return m.dto.MinQuantity }

func (m *market) MaxQuantity() decimal.Decimal { return m.dto.MaxQuantity }

func (m *market) QuantityStepSize() decimal.Decimal { return m.dto.QuantityStepSize }

func (m *market) Ticker() (types.Ticker, error) {
	return m.trader.TickerSvc().Ticker(m)
}

func (m *market) TickerStream(stop <-chan bool) <-chan types.Ticker {
	return m.trader.TickerSvc().TickerStream(stop, m)
}

func (m *market) Candles(interval types.CandleInterval, start time.Time, end time.Time) ([]types.Candle, error) {
	candles := []types.Candle{}
	dtos, err := m.trader.Provider().Candles(m.ToDTO(), interval, start, end)
	if err != nil {
		return candles, err
	}

	for _, c := range dtos {
		candles = append(candles, candle.New(c))
	}

	return candles, nil
}

func (m *market) AttemptOrder(t types.OrderType, s types.OrderSide, p decimal.Decimal, q decimal.Decimal) (types.Order, error) {
	return m.trader.OrderSvc().AttemptOrder(m, t, s, p, q)
}
