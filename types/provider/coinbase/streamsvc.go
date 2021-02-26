package coinbase

import (
	"sync"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/spf13/viper"
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
	workingOrders map[string]*workingOrder
	idMapper      map[string]string

	tickerMtx     sync.RWMutex
	tickerStreams map[types.MarketDTO]chan types.TickerDTO
}

type workingOrder struct {
	id          string
	expireTimer *time.Timer
	updates     []interface{}
}

func newStreamService(stop <-chan bool, wsSvc *websocketSvc) (svc *streamSvc) {
	svc = &streamSvc{
		stop:          stop,
		wsSvc:         wsSvc,
		orderStreams:  make(map[<-chan bool]*orderStreamWrapper),
		tickerStreams: make(map[types.MarketDTO]chan types.TickerDTO),
		workingOrders: make(map[string]*workingOrder),
		log:           log.WithField("source", "coinbase.streamSvc"),
		idMapper:      map[string]string{},
	}

	svc.registerTickerHandler()
	svc.registerOrderReceivedHandler()
	svc.registerOrderOpenHandler()
	svc.registerOrderDoneHandler()
	svc.registerOrderMatchHandler()
	svc.registerOrderChangeHandler()

	go svc.tickerStreamSink()
	go svc.orderReceivedStreamSink()
	go svc.orderOpenStreamSink()
	go svc.orderDoneStreamSink()
	go svc.orderMatchStreamSink()
	go svc.orderChangeStreamSink()

	return
}

func (svc *streamSvc) registerTickerHandler() {
	svc.tickerHandler = newTickerHandler(svc.stop)
	svc.wsSvc.RegisterMessageHandler(svc.tickerHandler)
}

func (svc *streamSvc) registerOrderReceivedHandler() {
	svc.orderReceivedHandler = newOrderReceivedHandler(svc.stop)
	svc.wsSvc.RegisterMessageHandler(svc.orderReceivedHandler)
}

func (svc *streamSvc) registerOrderOpenHandler() {
	svc.orderOpenHandler = newOrderOpenHandler(svc.stop)
	svc.wsSvc.RegisterMessageHandler(svc.orderOpenHandler)
}

func (svc *streamSvc) registerOrderDoneHandler() {
	svc.orderDoneHandler = newOrderDoneHandler(svc.stop)
	svc.wsSvc.RegisterMessageHandler(svc.orderDoneHandler)
}

func (svc *streamSvc) registerOrderMatchHandler() {
	svc.orderMatchHandler = newOrderMatchHandler(svc.stop)
	svc.wsSvc.RegisterMessageHandler(svc.orderMatchHandler)
}

func (svc *streamSvc) registerOrderChangeHandler() {
	svc.orderChangeHandler = newOrderChangeHandler(svc.stop)
	svc.wsSvc.RegisterMessageHandler(svc.orderChangeHandler)
}

func (svc *streamSvc) TickerStream(stop <-chan bool, market types.MarketDTO) (stream <-chan types.TickerDTO, err error) {
	// Create the stream
	svc.log.Debugf("ticker stream request for %s", market.Name)
	rawStream := make(chan types.TickerDTO, viper.GetInt("coinbase.streams.tickerStreamBufferSize"))
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
		stream: make(chan types.OrderDTO, viper.GetInt("coinbase.streams.orderStreamBufferSize")),
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
					select {
					case wrapper.stream <- v.ToDTO(wrapper.dto):
					default:
						log.Warn("received wrapper stream is blocked")
					}
				case Open:
					select {
					case wrapper.stream <- v.ToDTO(wrapper.dto):
					default:
						log.Warn("open wrapper stream is blocked")
					}
				case Done:
					select {
					case wrapper.stream <- v.ToDTO(wrapper.dto):
					default:
						log.Warn("done wrapper stream is blocked")
					}
				case Match:
					select {
					case wrapper.stream <- v.ToDTO(wrapper.dto):
					default:
						log.Warn("match wrapper stream is blocked")
					}
				case Change:
					select {
					case wrapper.stream <- v.ToDTO(wrapper.dto):
					default:
						log.Warn("change wrapper stream is blocked")
					}
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
	// Build the subscribe request
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
		svc.workingOrders[id] = &workingOrder{id: id, updates: []interface{}{}, expireTimer: time.NewTimer(viper.GetDuration("coinbase.websocket.workingOrderExpiration"))}
		go func(ord *workingOrder) {
			<-ord.expireTimer.C
			svc.orderMtx.Lock()
			delete(svc.workingOrders, ord.id)
			svc.orderMtx.Unlock()
		}(svc.workingOrders[id])
	}
	order := svc.workingOrders[id]
	order.updates = append(order.updates, data)
	order.expireTimer.Reset(viper.GetDuration("coinbase.websocket.workingOrderExpiration"))
	svc.workingOrders[id] = order
}

func (svc *streamSvc) registerClientId(orderId string, clientId string) {
	// Setting the id mapping
	svc.orderMtx.Lock()
	svc.idMapper[orderId] = clientId
	svc.orderMtx.Unlock()
}

func (svc *streamSvc) GetClientOrderIDFromOrderID(orderID string) string {
	svc.orderMtx.RLock()
	defer svc.orderMtx.RUnlock()

	return svc.idMapper[orderID]
}

func (svc *streamSvc) orderReceivedStreamSink() {
	for {
		select {
		case <-svc.stop:
			return
		case orderData := <-svc.orderReceivedHandler.Output():
			// Bail out if there's no ClientOrderID
			clientId := orderData.ClientOrderID
			if clientId == "" {
				continue
			}
			svc.registerClientId(orderData.OrderID, clientId)

			svc.orderMtx.RLock()
			svc.log.Debug("sending order received data to streams")
			for _, wrapper := range svc.orderStreams {
				if wrapper.dto.ID == clientId {
					select {
					case wrapper.stream <- orderData.ToDTO(wrapper.dto):
						log.WithField("dto", orderData.ToDTO(wrapper.dto)).Debugf("sending data for order %s", wrapper.dto.ID)
					default:
						log.WithField("dto", orderData.ToDTO(wrapper.dto)).Warn("skipping blocked order stream")
					}
				}
			}
			svc.orderMtx.RUnlock()

			// Update working orders
			svc.log.Debug("adding order received data to working orders")
			svc.updateWorkingOrders(clientId, orderData)
		}
	}
}

func (svc *streamSvc) orderOpenStreamSink() {
	for {
		select {
		case <-svc.stop:
			return
		case orderData := <-svc.orderOpenHandler.Output():
			// Find the client ID
			clientId := svc.GetClientOrderIDFromOrderID(orderData.OrderID)

			// Bail if not found
			if clientId == "" {
				continue
			}

			// Send the data
			svc.orderMtx.RLock()
			svc.log.Debug("sending order open data to streams")
			for _, wrapper := range svc.orderStreams {
				if wrapper.dto.ID == clientId {
					select {
					case wrapper.stream <- orderData.ToDTO(wrapper.dto):
						log.WithField("dto", orderData.ToDTO(wrapper.dto)).Debugf("sending data for order %s", wrapper.dto.ID)
					default:
						log.WithField("dto", orderData.ToDTO(wrapper.dto)).Warn("skipping blocked order stream")
					}
				}
			}
			svc.orderMtx.RUnlock()

			// Update working orders
			svc.log.Debug("adding order open data to working orders")
			svc.updateWorkingOrders(clientId, orderData)
		}
	}
}

func (svc *streamSvc) orderDoneStreamSink() {
	for {
		select {
		case <-svc.stop:
			return
		case orderData := <-svc.orderDoneHandler.Output():
			// Find the client ID
			clientId := svc.GetClientOrderIDFromOrderID(orderData.OrderID)

			// Bail if not found
			if clientId == "" {
				continue
			}

			// Send the data
			svc.orderMtx.RLock()
			svc.log.Debug("sending order done data to streams")
			for _, wrapper := range svc.orderStreams {
				if wrapper.dto.ID == clientId {
					select {
					case wrapper.stream <- orderData.ToDTO(wrapper.dto):
						log.WithField("dto", orderData.ToDTO(wrapper.dto)).Debugf("sending data for order %s", wrapper.dto.ID)
					default:
						log.WithField("dto", orderData.ToDTO(wrapper.dto)).Warn("skipping blocked order stream")
					}
				}
			}
			svc.orderMtx.RUnlock()

			// Update working orders
			svc.log.Debug("adding order done data to working orders")
			svc.updateWorkingOrders(clientId, orderData)
		}
	}
}

func (svc *streamSvc) orderMatchStreamSink() {
	for {
		select {
		case <-svc.stop:
			return
		case orderData := <-svc.orderMatchHandler.Output():
			// Find the client ID
			takerId := svc.GetClientOrderIDFromOrderID(orderData.TakerOrderID)
			makerId := svc.GetClientOrderIDFromOrderID(orderData.MakerOrderID)

			// Send the data
			svc.orderMtx.RLock()
			svc.log.Debug("sending order match data to streams")
			for _, wrapper := range svc.orderStreams {
				if wrapper.dto.ID == makerId || wrapper.dto.ID == takerId {
					select {
					case wrapper.stream <- orderData.ToDTO(wrapper.dto):
						log.WithField("dto", orderData.ToDTO(wrapper.dto)).Debugf("sending data for order %s", wrapper.dto.ID)
					default:
						log.WithField("dto", orderData.ToDTO(wrapper.dto)).Warn("skipping blocked order stream")
					}
				}
			}
			svc.orderMtx.RUnlock()

			// Update working orders
			svc.log.Debug("adding order match data to working orders")
			svc.updateWorkingOrders(takerId, orderData)
			svc.updateWorkingOrders(makerId, orderData)
		}
	}
}

func (svc *streamSvc) orderChangeStreamSink() {
	for {
		select {
		case <-svc.stop:
			return
		case orderData := <-svc.orderChangeHandler.Output():
			// Find the client ID
			clientId := svc.GetClientOrderIDFromOrderID(orderData.OrderID)

			// Bail if not found
			if clientId == "" {
				continue
			}

			// Send the data
			svc.orderMtx.RLock()
			svc.log.Debug("sending order change data to streams")
			for _, wrapper := range svc.orderStreams {
				if wrapper.dto.ID == clientId {
					select {
					case wrapper.stream <- orderData.ToDTO(wrapper.dto):
						log.WithField("dto", orderData.ToDTO(wrapper.dto)).Debugf("sending data for order %s", wrapper.dto.ID)
					default:
						log.WithField("dto", orderData.ToDTO(wrapper.dto)).Warn("skipping blocked order stream")
					}
				}
			}
			svc.orderMtx.RUnlock()

			// Update working orders
			svc.log.Debug("adding order change data to working orders")
			svc.updateWorkingOrders(clientId, orderData)
		}
	}
}
