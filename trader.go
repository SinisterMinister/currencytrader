package currencytrader

import (
	"sync"

	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/svc"
)

type Trader struct {
	provider  types.Provider
	marketSvc types.MarketSvc
	tickerSvc types.TickerSvc
	walletSvc types.WalletSvc
	orderSvc  types.OrderSvc

	mutex   sync.RWMutex
	stop    chan bool
	running bool
}

func New(provider types.Provider) (t types.Trader) {
	t = &Trader{
		provider:  provider,
		marketSvc: svc.NewMarket(provider),
		tickerSvc: svc.NewTicker(provider),
		stop:      make(chan bool),
	}
	return t
}

func (t *Trader) Start() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.startServices()
}

func (t *Trader) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.stopServices()
}

func (t *Trader) startServices() {
	if !t.running {
		t.tickerSvc.Start()
		t.marketSvc.Start()

		t.running = true
	}
}

func (t *Trader) stopServices() {
	if t.running {
		t.tickerSvc.Stop()
		t.marketSvc.Stop()

		t.running = false
	}
}

func (t *Trader) OrderSvc() types.OrderSvc {
	return t.orderSvc
}

func (t *Trader) WalletSvc() types.WalletSvc {
	return t.walletSvc
}

func (t *Trader) MarketSvc() types.MarketSvc {
	return t.marketSvc
}

func (t *Trader) TickerSvc() types.TickerSvc {
	return t.tickerSvc
}
