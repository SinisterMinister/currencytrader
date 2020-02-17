package coinbase

import (
	"sync"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/thoas/go-funk"
)

type streamSvc struct {
	log           log.Entry
	wsSvc         *websocketSvc
	tickerHandler *tickerHandler
	stop          <-chan bool

	orderMtx     sync.RWMutex
	orderStreams map[types.OrderDTO]chan types.OrderDTO

	tickerMtx     sync.RWMutex
	tickerStreams map[types.MarketDTO]chan types.TickerDTO
}

func newStreamService(stop <-chan bool, wsSvc *websocketSvc) (svc *streamSvc) {
	svc = &streamSvc{
		stop:          stop,
		wsSvc:         wsSvc,
		orderStreams:  make(map[types.OrderDTO]chan types.OrderDTO),
		tickerStreams: make(map[types.MarketDTO]chan types.TickerDTO),
		log:           log.WithField("source", "coinbase.streamSvc"),
	}

	svc.registerTickerHandler()

	go svc.tickerStreamSink()
	return
}

func (svc *streamSvc) registerTickerHandler() {
	svc.tickerHandler = newTickerHandler(svc.stop)
	svc.wsSvc.RegisterMessageHandler("ticker", svc.tickerHandler)
}

func (svc *streamSvc) TickerStream(stop <-chan bool, market types.MarketDTO) (stream <-chan types.TickerDTO, err error) {
	// Create the stream
	svc.log.Debugf("ticker stream request for %s", market.Name)
	rawStream := make(chan types.TickerDTO)
	stream = rawStream
	svc.tickerMtx.Lock()
	svc.tickerStreams[market] = rawStream
	svc.tickerMtx.Unlock()

	// Update the subscriptions
	svc.updateWebsocketSubscriptions()

	// Handle stop
	go func() {
		select {
		case <-stop:
			svc.tickerMtx.Lock()
			delete(svc.tickerStreams, market)
			svc.tickerMtx.Unlock()

			svc.updateWebsocketSubscriptions()
		}
	}()

	return
}

func (svc *streamSvc) OrderStream(stop <-chan bool, order types.OrderDTO) (stream <-chan types.OrderDTO, err error) {
	// Create the stream
	rawStream := make(chan types.OrderDTO)
	stream = rawStream
	svc.orderMtx.Lock()
	svc.orderStreams[order] = rawStream
	svc.orderMtx.Unlock()

	// Update the subscriptions
	svc.updateWebsocketSubscriptions()

	// Handle stop
	go func() {
		select {
		case <-stop:
			// Remove the stream from the list of streams
			svc.orderMtx.Lock()
			delete(svc.orderStreams, order)
			svc.orderMtx.Unlock()

			// Update the subscriptions
			svc.updateWebsocketSubscriptions()
		}
	}()

	return
}

func (svc *streamSvc) updateWebsocketSubscriptions() {
	var tickerSubs, fullSubs []string
	subs := svc.wsSvc.Subscriptions()

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
				svc.orderMtx.RLock()
				for order := range svc.orderStreams {
					if order.Market.Name == id {
						watched = true
						break
					}
				}
				svc.orderMtx.RUnlock()

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
				svc.tickerMtx.RLock()
				for market := range svc.tickerStreams {
					if market.Name == id {
						watched = true
						break
					}
				}
				svc.tickerMtx.RUnlock()

				// If it's not watched, unsubscribe
				if !watched {
					svc.unsubscribe(channel.Name, id)
				}
			}
		default:
			svc.log.Warnf("unexpected channel type %s", channel.Name)
		}
	}

	// Add any missing ticker subscriptions
	svc.tickerMtx.RLock()
	for market := range svc.tickerStreams {
		if !funk.Contains(tickerSubs, market.Name) {
			svc.subscribe("ticker", market.Name)
		}
	}
	svc.tickerMtx.RUnlock()

	// Add missing full subscriptions
	svc.orderMtx.RLock()
	for order := range svc.orderStreams {
		if !funk.Contains(fullSubs, order.Market.Name) {
			svc.subscribe("full", order.Market.Name)
		}
	}
	svc.orderMtx.RUnlock()
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
	svc.wsSvc.Unsubscribe(req)
}

func (svc *streamSvc) subscribe(channel string, productID string) {
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
	svc.wsSvc.Subscribe(req)
}

func (svc *streamSvc) tickerStreamSink() {
	for {
		select {
		case <-svc.stop:
			return
		case ticker := <-svc.tickerHandler.Output():
			svc.tickerMtx.RLock()
			svc.log.Debug("sending ticker data to streams")
			for market, stream := range svc.tickerStreams {
				if market.Name == ticker.ProductID {
					stream <- types.TickerDTO{
						Ask:       ticker.BestAsk,
						Bid:       ticker.BestBid,
						Price:     ticker.Price,
						Quantity:  ticker.LastSize,
						Timestamp: ticker.Time,
					}
				}
			}
			svc.tickerMtx.RUnlock()
		}
	}
}
