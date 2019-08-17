package trader

import (
	"sync"

	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
	"github.com/sinisterminister/currencytrader/types/svc"
)

type trader struct {
	provider  types.Provider
	marketSvc internal.MarketSvc
	tickerSvc internal.TickerSvc
	walletSvc internal.WalletSvc
	orderSvc  types.OrderSvc

	mutex   sync.RWMutex
	stop    chan bool
	running bool
}

func New(provider types.Provider) internal.Trader {
	t := &trader{
		provider: provider,
		stop:     make(chan bool),
	}

	t.marketSvc = svc.NewMarket(t)
	t.tickerSvc = svc.NewTicker(t)
	t.walletSvc = svc.(provider)
	return t
}

func (t *trader) Start() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.startServices()
}

func (t *trader) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.stopServices()
}

func (t *trader) startServices() {
	if !t.running {
		t.tickerSvc.Start()
		t.marketSvc.Start()
		t.walletSvc.Start()

		t.running = true
	}
}

func (t *trader) stopServices() {
	if t.running {
		t.tickerSvc.Stop()
		t.marketSvc.Stop()

		t.running = false
	}
}

func (t *trader) OrderSvc() types.OrderSvc {
	return t.orderSvc
}

func (t *trader) WalletSvc() types.WalletSvc {
	return t.walletSvc
}

func (t *trader) MarketSvc() types.MarketSvc {
	return t.marketSvc
}

func (t *trader) TickerSvc() types.TickerSvc {
	return t.tickerSvc
}

func (t *trader) Provider() types.Provider {
	return t.provider
}
