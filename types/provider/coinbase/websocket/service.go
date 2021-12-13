package websocket

import (
	"encoding/json"
	"sync"

	"github.com/go-playground/log/v7"
	ws "github.com/gorilla/websocket"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/handlers"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/processors"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type service struct {
	log  log.Entry
	stop <-chan bool

	changeHandler        types.ChangeMessageHandler
	doneHandler          types.DoneMessageHandler
	matchHandler         types.MatchMessageHandler
	openHandler          types.OpenMessageHandler
	receivedHandler      types.ReceivedMessageHandler
	subscriptionsHandler types.SubscriptionsMessageHandler
	tickerHandler        types.TickerMessageHandler

	changeProcessor        processors.Change
	doneProcessor          processors.Done
	matchProcessor         processors.Match
	openProcessor          processors.Open
	receivedProcessor      processors.Received
	subscriptionsProcessor processors.Subscriptions
	tickerProcessor        processors.Ticker

	incomingData chan types.DataPackage

	connRMtx   sync.Mutex
	connWMtx   sync.Mutex
	connection *ws.Conn

	subsMtx       sync.RWMutex
	subscriptions types.Subscriptions
}

func NewService(stop <-chan bool) (s types.Service, err error) {
	// Setup the service
	svc := &service{
		incomingData: make(chan types.DataPackage, viper.GetInt("coinbase.websocket.service.incomingDataBufferSize")),
		log:          log.WithField("source", "coinbase.websocket.service"),
		stop:         stop,
	}

	// Initialize the handlers
	if err = svc.initializeHandlers(); err != nil {
		return
	}

	// Initialize the processors
	if err = svc.initializeProcessors(); err != nil {
		return
	}

	// Initialize the connection
	if err = svc.initializeConnection(); err != nil {
		return
	}

	// Start message processor
	go svc.processMessages()

	// Start connection reader
	go svc.readConnection()

	// Start subscriptions processor
	go svc.processSignals()

	return svc, err
}

func (svc *service) Subscriptions() types.Subscriptions {
	svc.subsMtx.RLock()
	defer svc.subsMtx.RUnlock()
	return svc.subscriptions
}

func (svc *service) Subscribe(req types.Subscribe) (err error) {
	// Make sure the message type is correct
	req.Type = "subscribe"

	return svc.sendSubRequest(req)
}

func (svc *service) Unsubscribe(req types.Subscribe) (err error) {
	// Make sure the message type is correct
	req.Type = "unsubscribe"

	return svc.sendSubRequest(req)
}

func (svc *service) UpdateSubscriptions(subs types.Subscriptions) {
	svc.log.Debug("updating subscriptions")
	svc.subsMtx.Lock()
	svc.subscriptions = subs
	svc.subsMtx.Unlock()
}

func (svc *service) sendSubRequest(sub types.Subscribe) (err error) {
	svc.connWMtx.Lock()
	defer svc.connWMtx.Unlock()

	svc.log.WithField("request", sub).Debug("sending subscription request")
	return svc.connection.WriteJSON(sub)
}

func (svc *service) initializeHandlers() (err error) {
	// Setup the change handler
	changeHandler, err := handlers.Change(svc.stop)
	if err != nil {
		return
	}

	// Setup the done handler
	doneHandler, err := handlers.Done(svc.stop)
	if err != nil {
		return
	}

	// Setup the match handler
	matchHandler, err := handlers.Match(svc.stop)
	if err != nil {
		return
	}

	// Setup the open handler
	openHandler, err := handlers.Open(svc.stop)
	if err != nil {
		return
	}

	// Setup the received handler
	receivedHandler, err := handlers.Received(svc.stop)
	if err != nil {
		return
	}

	// Setup the subs handler
	subsHandler, err := handlers.Subscriptions(svc.stop)
	if err != nil {
		return
	}

	// Setup the ticker handler
	tickerHandler, err := handlers.Ticker(svc.stop)
	if err != nil {
		return
	}

	// Set the handlers
	svc.changeHandler = changeHandler
	svc.doneHandler = doneHandler
	svc.matchHandler = matchHandler
	svc.openHandler = openHandler
	svc.receivedHandler = receivedHandler
	svc.subscriptionsHandler = subsHandler
	svc.tickerHandler = tickerHandler

	return
}

func (svc *service) initializeProcessors() (err error) {
	// Setup the change processor
	change, err := processors.NewChange(svc.stop)
	if err != nil {
		return
	}

	// Setup the done processor
	done, err := processors.NewDone(svc.stop)
	if err != nil {
		return
	}

	// Setup the match processor
	match, err := processors.NewMatch(svc.stop)
	if err != nil {
		return
	}

	// Setup the open processor
	open, err := processors.NewOpen(svc.stop)
	if err != nil {
		return
	}

	// Setup the received processor
	received, err := processors.NewReceived(svc.stop)
	if err != nil {
		return
	}

	// Setup the subscriptions processor
	subscriptions, err := processors.NewSubscriptions(svc.stop, svc)
	if err != nil {
		return
	}

	// Setup the ticker processor
	ticker, err := processors.NewTicker(svc.stop)
	if err != nil {
		return
	}

	// Set the processors
	svc.changeProcessor = change
	svc.doneProcessor = done
	svc.matchProcessor = match
	svc.openProcessor = open
	svc.receivedProcessor = received
	svc.subscriptionsProcessor = subscriptions
	svc.tickerProcessor = ticker

	// Process the data
	go svc.processChangeHandlerData()
	go svc.processDoneHandlerData()
	go svc.processMatchHandlerData()
	go svc.processOpenHandlerData()
	go svc.processReceivedHandlerData()
	go svc.processSubscriptionsHandlerData()
	go svc.processTickerHandlerData()

	return
}

func (svc *service) initializeConnection() (err error) {
	url := viper.GetString("coinbase.websocketURL")
	svc.log.Debugf("connecting to %s", url)

	svc.connRMtx.Lock()
	svc.connWMtx.Lock()
	if svc.connection != nil {
		svc.connection.Close()
	}
	svc.connRMtx.Unlock()
	svc.connWMtx.Unlock()
	svc.connection, _, err = ws.DefaultDialer.Dial(url, nil)

	// Resubscribe to any previous subscriptions
	if len(svc.subscriptions.Channels) > 0 {
		// Build the subscribe request
		req := types.Subscribe{Channels: svc.subscriptions.Channels}
		err = svc.Subscribe(req)
	}

	return
}

func (svc *service) readConnection() {
	for {
		select {
		// Time to stop
		case <-svc.stop:
			return

		// Read the next message
		default:
			svc.log.Debug("reading message from websocket")
			svc.connRMtx.Lock()
			_, data, err := svc.connection.ReadMessage()
			svc.connRMtx.Unlock()
			if err != nil {
				svc.log.WithError(err).WithTrace().Error("error readding message from socket. restarting connection")
				svc.initializeConnection()
				continue
			}

			// Try to parse the data into a message
			svc.log.Debug("parsing message from websocket")
			var message types.Message
			err = json.Unmarshal(data, &message)
			if err != nil {
				log.WithError(err).WithField("payload", string(data)).Warn("could not parse message from server")
				continue
			}

			// Send message
			select {
			case svc.incomingData <- types.DataPackage{
				Data:    data,
				Message: message,
			}:
			default:
				log.Warn("incoming data channel blocked")
			}
		}
	}
}

func (svc *service) processMessages() {
	svc.log.Debug("starting message processor")
	for {
		select {
		// Time to exit
		case <-svc.stop:
			svc.log.Debug("stopping message processor")
			return

		// Send incoming messages to their respective handlers
		case pkg := <-svc.incomingData:
			var handler types.MessageHandler

			svc.log.Debug("looking up handler to send data")
			switch pkg.Type {
			case svc.changeHandler.Name():
				handler = svc.changeHandler
			case svc.doneHandler.Name():
				handler = svc.doneHandler
			case svc.matchHandler.Name():
				handler = svc.matchHandler
			case svc.openHandler.Name():
				handler = svc.openHandler
			case svc.receivedHandler.Name():
				handler = svc.receivedHandler
			case svc.subscriptionsHandler.Name():
				handler = svc.subscriptionsHandler
			case svc.tickerHandler.Name():
				handler = svc.subscriptionsHandler

			// Bail out for unregistered types
			default:
				svc.log.Warnf("unregistered message type %s receieved", pkg.Type)
				continue
			}

			svc.log.Debugf("sending message to '%s' handler", handler.Name())
			select {
			case handler.Input() <- pkg:
			default:
				log.Warnf("%s handler input channel blocked", handler.Name())
			}
		}
	}
}

func (svc *service) processSignals() {
	svc.log.Debug("starting signal processor")
	for {
		select {
		// Kill switch flipped
		case <-svc.stop:
			svc.log.Debug("stopping signal processor")
			return

		case <-svc.subscriptionsHandler.Output():
		}
	}
}

func (svc *service) processChangeHandlerData() {
	for {
		select {
		// Kill switch flipped
		case <-svc.stop:
			return
		case out := <-svc.changeHandler.Output():
			select {
			case svc.changeProcessor.Input() <- out:
			default:
				log.Warn("change processor backed up")
				svc.changeProcessor.Input() <- out
				log.Warn("change processor freed up")
			}
		}
	}
}

func (svc *service) processDoneHandlerData() {
	for {
		select {
		// Kill switch flipped
		case <-svc.stop:
			return
		case out := <-svc.doneHandler.Output():
			select {
			case svc.doneProcessor.Input() <- out:
			default:
				log.Warn("done processor backed up")
				svc.doneProcessor.Input() <- out
				log.Warn("done processor freed up")
			}
		}
	}
}

func (svc *service) processMatchHandlerData() {
	for {
		select {
		// Kill switch flipped
		case <-svc.stop:
			return
		case out := <-svc.matchHandler.Output():
			select {
			case svc.matchProcessor.Input() <- out:
			default:
				log.Warn("match processor backed up")
				svc.matchProcessor.Input() <- out
				log.Warn("match processor freed up")
			}
		}
	}
}

func (svc *service) processOpenHandlerData() {
	for {
		select {
		// Kill switch flipped
		case <-svc.stop:
			return
		case out := <-svc.openHandler.Output():
			select {
			case svc.openProcessor.Input() <- out:
			default:
				log.Warn("open processor backed up")
				svc.openProcessor.Input() <- out
				log.Warn("open processor freed up")
			}
		}
	}
}

func (svc *service) processReceivedHandlerData() {
	for {
		select {
		// Kill switch flipped
		case <-svc.stop:
			return
		case out := <-svc.receivedHandler.Output():
			select {
			case svc.receivedProcessor.Input() <- out:
			default:
				log.Warn("received processor backed up")
				svc.receivedProcessor.Input() <- out
				log.Warn("received processor freed up")
			}
		}
	}
}

func (svc *service) processSubscriptionsHandlerData() {
	for {
		select {
		// Kill switch flipped
		case <-svc.stop:
			return
		case out := <-svc.subscriptionsHandler.Output():
			select {
			case svc.subscriptionsProcessor.Input() <- out:
			default:
				log.Warn("subscriptions processor backed up")
				svc.subscriptionsProcessor.Input() <- out
				log.Warn("subscriptions processor freed up")
			}
		}
	}
}

func (svc *service) processTickerHandlerData() {
	for {
		select {
		// Kill switch flipped
		case <-svc.stop:
			return
		case out := <-svc.tickerHandler.Output():
			select {
			case svc.tickerProcessor.Input() <- out:
			default:
				log.Warn("ticker processor backed up")
				svc.tickerProcessor.Input() <- out
				log.Warn("ticker processor freed up")
			}
		}
	}
}
