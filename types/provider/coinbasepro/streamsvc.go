package coinbasepro

import (
	"encoding/json"
	"sync"

	"github.com/go-playground/log"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/thoas/go-funk"
)

type streamSvc struct {
	wsHandler *websocketHandler

	mutex         sync.Mutex
	stop          <-chan bool
	wsStream      <-chan DataPackage
	orderStreams  map[types.OrderDTO]chan types.OrderDTO
	tickerStreams map[types.MarketDTO]chan types.TickerDTO
	walletStreams map[types.WalletDTO]chan types.WalletDTO
}

func newStreamService(stop <-chan bool, wsh *websocketHandler) *streamSvc {
	svc := &streamSvc{
		stop:          stop,
		wsHandler:     wsh,
		wsStream:      wsh.GetStream(stop),
		orderStreams:  make(map[types.OrderDTO]chan types.OrderDTO),
		tickerStreams: make(map[types.MarketDTO]chan types.TickerDTO),
		walletStreams: make(map[types.WalletDTO]chan types.WalletDTO),
	}

	go svc.streamSink()

	return svc
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
			svc.mutex.Lock()
			delete(svc.tickerStreams, market)
			svc.updateWebsocketSubscriptions()
			svc.mutex.Unlock()
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

func (svc *streamSvc) streamSink() {
	for {
		select {
		case <-svc.stop:
			return
		case pkg := <-svc.wsStream:
			select {
			case <-svc.stop:
				return
			default:
			}
			switch pkg.Type {
			case "ticker":
				svc.broadcastTicker(pkg)
			default:
				log.WithField("type", pkg.Type).Error("unexpected message")
			}
		}
	}
}

func (svc *streamSvc) broadcastOrder(data interface{}) {

}

func (svc *streamSvc) broadcastTicker(data DataPackage) {
	var ticker Ticker
	err := json.Unmarshal(data.Data, &ticker)
	if err != nil {
		log.WithError(err).Error("could not unmarshal ticker")
	}
	payload := types.TickerDTO{
		Ask:       ticker.BestAsk,
		Bid:       ticker.BestBid,
		Price:     ticker.Price,
		Quantity:  ticker.LastSize,
		Timestamp: ticker.Time,
	}

	// Copy the map
	streams := make(map[types.MarketDTO]chan types.TickerDTO)
	svc.mutex.Lock()
	for market, stream := range svc.tickerStreams {
		streams[market] = stream
	}
	svc.mutex.Unlock()

	for market, stream := range streams {
		if market.Name == ticker.ProductID {
			svc.mutex.Lock()
			select {
			case stream <- payload:
			default:
				log.Warn("ticker: skipping blocked stream")
			}
			svc.mutex.Unlock()
		}
	}
}

func (svc *streamSvc) broadcastWallet(data interface{}) {

}
