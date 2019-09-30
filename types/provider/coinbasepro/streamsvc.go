package coinbasepro

import "github.com/sinisterminister/currencytrader/types"

type streamSvc struct {
	wsHandler *websocketHandler
}

func (svc *streamSvc) OrderStream(stop <-chan bool, order types.OrderDTO) (stream <-chan types.OrderDTO, err error) {
	return
}

func (svc *streamSvc) TickerStream(stop <-chan bool, market types.MarketDTO) (stream <-chan types.TickerDTO, err error) {
	return
}

func (svc *streamSvc) WalletStream(stop <-chan bool, wal types.WalletDTO) (stream <-chan types.WalletDTO, err error) {
	return
}

func (svc *streamSvc) updateSubscriptions() {

}
