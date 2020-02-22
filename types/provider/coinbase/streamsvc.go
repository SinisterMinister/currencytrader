package coinbase

import (
	"sync"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/thoas/go-funk"
)

type streamSvc struct {
	log                  log.Entry
	wsSvc                *websocketSvc
	tickerHandler        *tickerHandler
	orderReceivedHandler *orderReceivedHandler
	orderOpenHandler     *orderOpenHandler
	orderDoneHandler     *orderDoneHandler
	orderMatchHandler    *orderMatchHandler
	orderChangeHandler   *orderChangeHandler
	stop                 <-chan bool

	orderMtx      sync.RWMutex
	orderStreams  map[<-chan bool]*orderStreamWrapper
	workingOrders map[string]workingOrder

	tickerMtx     sync.RWMutex
	tickerStreams map[types.MarketDTO]chan types.TickerDTO
}

type workingOrder struct {
	id      string
	doneAt  time.Time
	updates []interface{}
}

func newStreamService(stop <-chan bool, wsSvc *websocketSvc) (svc *streamSvc) {
	svc = &streamSvc{
		stop:          stop,
		wsSvc:         wsSvc,
		orderStreams:  make(map[<-chan bool]*orderStreamWrapper),
		tickerStreams: make(map[types.MarketDTO]chan types.TickerDTO),
		workingOrders: make(map[string]workingOrder),
		log:           log.WithField("source", "coinbase.streamSvc"),
	}

	svc.registerTickerHandler()
	svc.registerOrderReceivedHandler()
	svc.registerOrderOpenHandler()
	svc.registerOrderDoneHandler()
	svc.registerOrderMatchHandler()

	go svc.tickerStreamSink()
	go svc.orderReceivedStreamSink()
	go svc.orderOpenStreamSink()
	go svc.orderDoneStreamSink()
	go svc.orderMatchStreamSink()

	return
}

func (svc *streamSvc) registerTickerHandler() {
	svc.tickerHandler = newTickerHandler(svc.stop)
	svc.wsSvc.RegisterMessageHandler("ticker", svc.tickerHandler)
}

func (svc *streamSvc) registerOrderReceivedHandler() {
	svc.orderReceivedHandler = newOrderReceivedHandler(svc.stop)
	svc.wsSvc.RegisterMessageHandler("received", svc.orderReceivedHandler)
}

func (svc *streamSvc) registerOrderOpenHandler() {
	svc.orderOpenHandler = newOrderOpenHandler(svc.stop)
	svc.wsSvc.RegisterMessageHandler("open", svc.orderOpenHandler)
}

func (svc *streamSvc) registerOrderDoneHandler() {
	svc.orderDoneHandler = newOrderDoneHandler(svc.stop)
	svc.wsSvc.RegisterMessageHandler("done", svc.orderDoneHandler)
}

func (svc *streamSvc) registerOrderMatchHandler() {
	svc.orderMatchHandler = newOrderMatchHandler(svc.stop)
	svc.wsSvc.RegisterMessageHandler("match", svc.orderMatchHandler)
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

type orderStreamWrapper struct {
	dto    types.OrderDTO
	stream chan types.OrderDTO
}

func (svc *streamSvc) OrderStream(stop <-chan bool, order types.OrderDTO) (stream <-chan types.OrderDTO, err error) {
	// Create the stream
	wrapper := &orderStreamWrapper{
		dto:    order,
		stream: make(chan types.OrderDTO),
	}
	stream = wrapper.stream
	svc.orderMtx.Lock()
	svc.orderStreams[stop] = wrapper
	svc.orderMtx.Unlock()

	// Update the subscriptions
	svc.updateWebsocketSubscriptions()

	// Handle updating the stream with working data if any
	go func() {
		svc.orderMtx.RLock()
		defer svc.orderMtx.RUnlock()
		if data, ok := svc.workingOrders[order.ID]; ok {
			for _, d := range data.updates {
				switch v := d.(type) {
				case Received:
					wrapper.stream <- v.ToDTO(wrapper.dto)
				case Open:
					wrapper.stream <- v.ToDTO(wrapper.dto)
				case Done:
					wrapper.stream <- v.ToDTO(wrapper.dto)
				case Match:
					wrapper.stream <- v.ToDTO(wrapper.dto)
				case Change:
					wrapper.stream <- v.ToDTO(wrapper.dto)
				}
			}
		}
	}()

	// Handle stop
	go func() {
		select {
		case <-stop:
			// Remove the stream from the list of streams
			svc.orderMtx.Lock()
			delete(svc.orderStreams, stop)
			svc.orderMtx.Unlock()

			// Update the subscriptions
			svc.updateWebsocketSubscriptions()
		}
	}()

	return
}

func (svc *streamSvc) updateStreamWithWorkingData(id string, wrapper *orderStreamWrapper) {
	svc.orderMtx.RLock()
	if data, ok := svc.workingOrders[id]; ok {
		for _, d := range data.updates {
			switch v := d.(type) {
			case Received:
				wrapper.stream <- v.ToDTO(wrapper.dto)
			}
		}
	}
	svc.orderMtx.RUnlock()
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
				for _, wrapper := range svc.orderStreams {
					if wrapper.dto.Market.Name == id {
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
	for _, wrapper := range svc.orderStreams {
		if !funk.Contains(fullSubs, wrapper.dto.Market.Name) {
			svc.subscribe("full", wrapper.dto.Market.Name)
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

func (svc *streamSvc) updateWorkingOrders(id string, data interface{}) {
	svc.orderMtx.Lock()
	defer svc.orderMtx.Unlock()
	if _, ok := svc.workingOrders[id]; !ok {
		svc.workingOrders[id] = workingOrder{id: id, updates: []interface{}{}}
	}
	order := svc.workingOrders[id]
	order.updates = append(order.updates, data)
	svc.workingOrders[id] = order
}

func (svc *streamSvc) orderReceivedStreamSink() {
	for {
		select {
		case <-svc.stop:
			return
		case orderData := <-svc.orderReceivedHandler.Output():
			svc.orderMtx.RLock()
			svc.log.Debug("sending order received data to streams")
			for _, wrapper := range svc.orderStreams {
				if wrapper.dto.ID == orderData.OrderID {
					wrapper.stream <- orderData.ToDTO(wrapper.dto)
				}
			}
			svc.orderMtx.RUnlock()

			// Update working orders
			svc.log.Debug("adding order received data to working orders")
			svc.updateWorkingOrders(orderData.OrderID, orderData)
		}
	}
}

func (svc *streamSvc) orderOpenStreamSink() {
	for {
		select {
		case <-svc.stop:
			return
		case orderData := <-svc.orderOpenHandler.Output():
			svc.orderMtx.RLock()
			svc.log.Debug("sending order open data to streams")
			for _, wrapper := range svc.orderStreams {
				if wrapper.dto.ID == orderData.OrderID {
					wrapper.stream <- orderData.ToDTO(wrapper.dto)
				}
			}
			svc.orderMtx.RUnlock()

			// Update working orders
			svc.log.Debug("adding order open data to working orders")
			svc.updateWorkingOrders(orderData.OrderID, orderData)
		}
	}
}

func (svc *streamSvc) orderDoneStreamSink() {
	for {
		select {
		case <-svc.stop:
			return
		case orderData := <-svc.orderDoneHandler.Output():
			svc.orderMtx.RLock()
			svc.log.Debug("sending order done data to streams")
			for _, wrapper := range svc.orderStreams {
				if wrapper.dto.ID == orderData.OrderID {
					wrapper.stream <- orderData.ToDTO(wrapper.dto)
				}
			}
			svc.orderMtx.RUnlock()

			// Update working orders
			svc.log.Debug("adding order done data to working orders")
			svc.updateWorkingOrders(orderData.OrderID, orderData)
		}
	}
}

func (svc *streamSvc) orderMatchStreamSink() {
	for {
		select {
		case <-svc.stop:
			return
		case orderData := <-svc.orderMatchHandler.Output():
			svc.orderMtx.RLock()
			svc.log.Debug("sending order match data to streams")
			for _, wrapper := range svc.orderStreams {
				if wrapper.dto.ID == orderData.MakerOrderID || wrapper.dto.ID == orderData.TakerOrderID {
					wrapper.stream <- orderData.ToDTO(wrapper.dto)
				}
			}
			svc.orderMtx.RUnlock()

			// Update working orders
			svc.log.Debug("adding order match data to working orders")
			svc.updateWorkingOrders(orderData.MakerOrderID, orderData)
			svc.updateWorkingOrders(orderData.TakerOrderID, orderData)
		}
	}
}

func (svc *streamSvc) orderChangeStreamSink() {
	for {
		select {
		case <-svc.stop:
			return
		case orderData := <-svc.orderChangeHandler.Output():
			svc.orderMtx.RLock()
			svc.log.Debug("sending order change data to streams")
			for _, wrapper := range svc.orderStreams {
				if wrapper.dto.ID == orderData.OrderID {
					wrapper.stream <- orderData.ToDTO(wrapper.dto)
				}
			}
			svc.orderMtx.RUnlock()

			// Update working orders
			svc.log.Debug("adding order change data to working orders")
			svc.updateWorkingOrders(orderData.OrderID, orderData)
		}
	}
}
