package template

import (
	"time"

	"github.com/sinisterminister/currencytrader/types"
)

type provider struct {
	trader types.Trader
}

func New() types.Provider {
	return &provider{}
}

func (p *provider) AttemptOrder(req types.OrderRequestDTO) (dto types.OrderDTO, err error) {
	return
}

func (p *provider) CancelOrder(ord types.OrderDTO) (err error) {
	return
}

func (p *provider) Candles(mkt types.MarketDTO, interval types.CandleInterval, start time.Time, end time.Time) (candles []types.CandleDTO, err error) {
	return
}

func (p *provider) Currencies() (curs []types.CurrencyDTO, err error) {
	return
}

func (p *provider) Markets() (mkts []types.MarketDTO, err error) {
	return
}

func (p *provider) Order(markets types.MarketDTO, id string) (ord types.OrderDTO, err error) {
	return
}

func (p *provider) OrderStream(stop <-chan bool, order types.OrderDTO) (stream <-chan types.OrderDTO, err error) {
	return
}

func (p *provider) Ticker(market types.MarketDTO) (tkr types.TickerDTO, err error) {
	return
}

func (p *provider) TickerStream(stop <-chan bool, market types.MarketDTO) (stream <-chan types.TickerDTO, err error) {
	return
}

func (p *provider) Wallet(id string) (wal types.WalletDTO, err error) {
	return
}

func (p *provider) Wallets() (wals []types.WalletDTO, err error) {
	return
}

func (p *provider) WalletStream(stop <-chan bool, wal types.WalletDTO) (stream <-chan types.WalletDTO, err error) {
	return
}
