package trader

import (
	"sync"

	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
	"github.com/sinisterminister/currencytrader/types/svc"
)

type trader struct {
	provider   types.Provider
	marketSvc  internal.MarketSvc
	tickerSvc  internal.TickerSvc
	accountSvc internal.AccountSvc
	orderSvc   types.OrderSvc

	mutex   sync.RWMutex
	stop    chan bool
	running bool
}

func New(provider types.Provider) internal.Trader {
	t := &trader{
		provider: provider,
		stop:     make(chan bool),
	}

	t.accountSvc = svc.NewAccount(t)
	t.marketSvc = svc.NewMarket(t)
	t.tickerSvc = svc.NewTicker(t)
	t.orderSvc = svc.NewOrder(t)
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

		t.running = true
	}
}

func (t *trader) stopServices() {
	if t.running {
		t.tickerSvc.Stop()

		t.running = false
	}
}

func (t *trader) OrderSvc() types.OrderSvc {
	return t.orderSvc
}

func (t *trader) AccountSvc() types.AccountSvc {
	return t.accountSvc
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
