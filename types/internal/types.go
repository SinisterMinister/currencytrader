package internal

import "github.com/sinisterminister/currencytrader/types"

type MarketSvc interface {
	types.MarketSvc
}

type TickerSvc interface {
	types.TickerSvc
	types.Administerable
}

type WalletSvc interface {
	types.WalletSvc
}

type Wallet interface {
	types.Wallet
	Update(dto types.WalletDTO)
}

type Order interface {
	types.Order
	Update(dto types.OrderDTO)
}

type Trader interface {
	types.Administerable
	OrderSvc() types.OrderSvc
	WalletSvc() types.WalletSvc
	MarketSvc() types.MarketSvc
	TickerSvc() types.TickerSvc
	Provider() types.Provider
}
