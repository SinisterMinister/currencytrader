package coinbasepro

import (
	"sync"

	"github.com/go-playground/log"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/thoas/go-funk"
)

type streamSvc struct {
	wsHandler *websocketHandler

	mutex         sync.Mutex
	orderStreams  map[types.OrderDTO]chan types.OrderDTO
	tickerStreams map[types.MarketDTO]chan types.TickerDTO
	walletStreams map[types.WalletDTO]chan types.WalletDTO
}

func newStreamService(wsh *websocketHandler) *streamSvc {
	return &streamSvc{
		wsHandler:     wsh,
		orderStreams:  make(map[types.OrderDTO]chan types.OrderDTO),
		tickerStreams: make(map[types.MarketDTO]chan types.TickerDTO),
		walletStreams: make(map[types.WalletDTO]chan types.WalletDTO),
	}
}

func (svc *streamSvc) OrderStream(stop <-chan bool, order types.OrderDTO) (stream <-chan types.OrderDTO, err error) {
	svc.mutex.Lock()
	defer svc.mutex.Unlock()

	// Create the stream
	rawStream := make(chan types.OrderDTO)
	stream = rawStream
	svc.orderStreams[order] = rawStream

	// Update the subscriptions
	svc.updateWebsocketSubscriptions()

	// Handle stop
	go func() {
		select {
		case <-stop:
			delete(svc.orderStreams, order)
			svc.updateWebsocketSubscriptions()
		}
	}()

	return
}

func (svc *streamSvc) TickerStream(stop <-chan bool, market types.MarketDTO) (stream <-chan types.TickerDTO, err error) {
	svc.mutex.Lock()
	defer svc.mutex.Unlock()

	// Create the stream
	rawStream := make(chan types.TickerDTO)
	stream = rawStream
	svc.tickerStreams[market] = rawStream

	// Update the subscriptions
	svc.updateWebsocketSubscriptions()

	// Handle stop
	go func() {
		select {
		case <-stop:
			delete(svc.tickerStreams, market)
			svc.updateWebsocketSubscriptions()
		}
	}()

	return
}

func (svc *streamSvc) WalletStream(stop <-chan bool, wal types.WalletDTO) (stream <-chan types.WalletDTO, err error) {
	return
}

func (svc *streamSvc) updateWebsocketSubscriptions() {
	var tickerSubs, fullSubs []string
	subs := svc.wsHandler.Subscriptions()

	// First, remove any unneeded subscriptions
	for _, channel := range subs.Channels {
		switch channel.Name {
		case "full":
			// Store the product ids for later
			fullSubs = channel.ProductIDs

			// Loop over the product ids
			for _, id := range channel.ProductIDs {
				var watched bool
				// Check if the ID is being watched
				for order := range svc.orderStreams {
					if order.Market.Name == id {
						watched = true
						break
					}
				}
				// If it's not watched, unsubscribe
				if !watched {
					svc.unsubscribe(channel.Name, id)
				}
			}
		case "ticker":
			// Store the product ids for later
			tickerSubs = channel.ProductIDs

			// Loop over the product ids
			for _, id := range channel.ProductIDs {
				var watched bool
				// Check if the ID is being watched
				for market := range svc.tickerStreams {
					if market.Name == id {
						watched = true
						break
					}
				}
				// If it's not watched, unsubscribe
				if !watched {
					svc.unsubscribe(channel.Name, id)
				}
			}
		default:
			log.Warnf("unexpected channel type %s", channel.Name)
		}
	}

	// Add any missing ticker subscriptions
	for market := range svc.tickerStreams {
		if !funk.Contains(tickerSubs, market.Name) {
			svc.subscribe("ticker", market.Name)
		}
	}

	// Add missing full subscriptions
	for order := range svc.orderStreams {
		if !funk.Contains(fullSubs, order.Market.Name) {
			svc.subscribe("full", order.Market.Name)
		}
	}
}

func (svc *streamSvc) unsubscribe(channel string, productID string) {
	// Build the unsubscribe request
	req := Subscribe{Channels: []struct {
		Name       string   `json:"name"`
		ProductIDs []string `json:"product_ids"`
	}{
		{
			Name:       channel,
			ProductIDs: append([]string{}, productID),
		},
	}}
	svc.wsHandler.Unsubscribe(req)
}

func (svc *streamSvc) subscribe(channel string, productId string) {
	// Build the unsubscribe request
	req := Subscribe{Channels: []struct {
		Name       string   `json:"name"`
		ProductIDs []string `json:"product_ids"`
	}{
		{
			Name:       channel,
			ProductIDs: append([]string{}, productId),
		},
	}}
	svc.wsHandler.Subscribe(req)
}
