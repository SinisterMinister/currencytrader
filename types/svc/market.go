package svc

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
	"github.com/sinisterminister/currencytrader/types/market"
)

// Market service
type Market struct {
	trader internal.Trader

	mutex          sync.RWMutex
	marketsRefresh *time.Timer
	markets        []types.Market
}

func NewMarket(trader internal.Trader) internal.MarketSvc {
	svc := &Market{
		trader: trader,
	}

	return svc
}

func (m *Market) Market(cur0 types.Currency, cur1 types.Currency) (market types.Market, err error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.updateMarkets()
	for _, mkt := range m.markets {
		if mkt.BaseCurrency().Name() == cur0.Name() && mkt.QuoteCurrency().Name() == cur1.Name() {
			market = mkt
			return
		}

		if mkt.BaseCurrency() == cur1 && mkt.QuoteCurrency() == cur0 {
			market = mkt
			return
		}
	}

	return market, fmt.Errorf("Could not find market for currencies '%s', '%s'", cur0.Name(), cur1.Name())
}

func (m *Market) Markets() []types.Market {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.updateMarkets()
	return m.markets
}

func (m *Market) updateMarkets() {
	if m.marketsRefresh != nil {
		select {
		default:
			// Bail out if it's not time to update
			return
		case <-m.marketsRefresh.C:
			// Time to update
		}
	}

	rawMarkets, err := m.trader.Provider().Markets()
	if err != nil {
		log.WithError(err).Error("Could not get markets from provider!")
	}
	markets := make([]types.Market, 0, len(rawMarkets))
	for _, dto := range rawMarkets {
		mkt := market.New(m.trader, dto)
		markets = append(markets, mkt)
	}

	m.markets = markets
}
