package svc

import (
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sirupsen/logrus"

	"github.com/sinisterminister/moneytrader/pkg"
)

// Market service
type Market struct {
	markets        []pkg.Market
	marketsRefresh *time.Timer
	provider       pkg.Provider
}

func NewMarket(provider pkg.Provider) pkg.MarketSvc {
	svc := &Market{
		provider: provider,
	}

	return svc
}

func (m *Market) GetMarket(cur0 pkg.Currency, cur1 pkg.Currency) (market pkg.Market, err error) {
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

	return market, fmt.Errorf("Could not find market for currencies '%s', '%s'", cur0.Name, cur1.Name)
}

func (m *Market) GetMarkets() []pkg.Market {
	return m.markets[:0]
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

	markets, err := m.provider.GetAllMarkets()
	if err != nil {
		logrus.WithError(err).Error("Could not get markets from provider!")
	}

	m.markets = markets
}
