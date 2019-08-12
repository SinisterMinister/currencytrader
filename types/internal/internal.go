package internal

import "github.com/sinisterminister/currencytrader/types"

type MarketSvc interface {
	types.MarketSvc
	types.Administerable
}

type TickerSvc interface {
	types.TickerSvc
	types.Administerable
}

type Trader interface {
	types.Administerable
	OrderSvc() types.OrderSvc
	WalletSvc() types.WalletSvc
	MarketSvc() types.MarketSvc
	TickerSvc() types.TickerSvc
	Provider() types.Provider
}
