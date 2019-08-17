package svc

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
	"github.com/sinisterminister/currencytrader/types/market"

	"github.com/sirupsen/logrus"
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

func (m *Market) GetMarket(cur0 types.Currency, cur1 types.Currency) (market types.Market, err error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	for _, mkt := range m.markets {
		if cmp.Equal(mkt.BaseCurrency, cur0) && cmp.Equal(mkt.QuoteCurrency, cur1) {
			market = mkt
			return
		}

		if cmp.Equal(mkt.BaseCurrency, cur1) && cmp.Equal(mkt.QuoteCurrency, cur0) {
			market = mkt
			return
		}
	}

	return market, fmt.Errorf("Could not find market for currencies '%s', '%s'", cur0.Name(), cur1.Name())
}

func (m *Market) GetMarkets() []types.Market {
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

	rawMarkets, err := m.trader.Provider().GetMarkets()
	if err != nil {
		logrus.WithError(err).Error("Could not get markets from provider!")
	}
	markets := make([]types.Market, 0, len(rawMarkets))
	for _, dto := range rawMarkets {
		conf := market.MarketConfig{
			MarketDTO: dto,
			TickerSvc: m.trader.TickerSvc(),
		}
		mkt := market.New(conf)
		markets = append(markets, mkt)
	}

	m.markets = markets
}
