package coinbase

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/go-playground/log/v7"
	ws "github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

type websocketSvc struct {
	log                   log.Entry
	stop                  <-chan bool
	incomingData          chan DataPackage
	incomingSubscriptions chan DataPackage
	messagesReceived      int
	messagesProcessed     int

	connRMtx   sync.Mutex
	connWMtx   sync.Mutex
	connection *ws.Conn

	subsMtx       sync.RWMutex
	subscriptions Subscriptions

	messageMtx      sync.RWMutex
	messageHandlers map[string]MessageHandler
}

func newWebsocketSvc(stop <-chan bool) (svc *websocketSvc, err error) {
	svc = &websocketSvc{
		stop:                  stop,
		incomingData:          make(chan DataPackage, viper.GetInt("coinbase.websocket.incomingDataBufferSize")),
		incomingSubscriptions: make(chan DataPackage, viper.GetInt("coinbase.websocket.incomingSubscriptionBufferSize")),
		log:                   log.WithField("source", "coinbase.websocketSvc"),
		messageHandlers:       make(map[string]MessageHandler),
	}

	// Initialize the connection
	if err = svc.initializeConnection(); err != nil {
		return
	}

	// Register self as subscriptions handler
	svc.log.Debug("registering subscriptions handler")
	svc.RegisterMessageHandler(svc)

	// Start subscriptions handler
	go svc.handleSubscriptions()

	// Start message processor
	go svc.processMessages()

	// Start connection reader
	go svc.readConnection()

	return
}

func (svc *websocketSvc) Subscriptions() Subscriptions {
	svc.subsMtx.RLock()
	defer svc.subsMtx.RUnlock()
	return svc.subscriptions
}

func (svc *websocketSvc) Subscribe(req Subscribe) (err error) {
	// Make sure the message type is correct
	req.Type = "subscribe"

	return svc.processSub(req)
}

func (svc *websocketSvc) Unsubscribe(req Subscribe) (err error) {
	// Make sure the message type is correct
	req.Type = "unsubscribe"

	return svc.processSub(req)
}

func (svc *websocketSvc) Input() chan<- DataPackage {
	return svc.incomingSubscriptions
}

func (svc *websocketSvc) Name() string {
	return "subscriptions"
}

func (svc *websocketSvc) RegisterMessageHandler(handler MessageHandler) (err error) {
	svc.messageMtx.Lock()
	defer svc.messageMtx.Unlock()

	if _, ok := svc.messageHandlers[handler.Name()]; ok {
		// Already registered
		return errors.New("handler for type already registered")
	}
	svc.log.Debugf("registering message handler for type '%s'", handler.Name())
	svc.messageHandlers[handler.Name()] = handler
	return
}

func (svc *websocketSvc) processSub(sub Subscribe) (err error) {
	svc.connWMtx.Lock()
	defer svc.connWMtx.Unlock()

	svc.log.WithField("request", sub).Debug("sending subscription request")
	return svc.connection.WriteJSON(sub)
}

func (svc *websocketSvc) handleSubscriptions() {
	for {
		select {
		// Kill switch flipped
		case <-svc.stop:
			return

		// Handle incoming subscriptions
		case pkg := <-svc.incomingSubscriptions:
			// Capture subscription responses
			var subs Subscriptions
			svc.log.Debug("handling subscription payload")
			e := json.Unmarshal(pkg.Data, &subs)
			if e != nil {
				svc.log.Error("could not parse subscriptions!")
				continue
			}

			svc.log.Debug("updating subscriptions")
			svc.subsMtx.Lock()
			svc.subscriptions = subs
			svc.subsMtx.Unlock()
		}
	}
}

func (svc *websocketSvc) initializeConnection() (err error) {
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
		req := Subscribe{Channels: svc.subscriptions.Channels}
		err = svc.Subscribe(req)
	}

	return
}

func (svc *websocketSvc) readConnection() {
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
			var message Message
			err = json.Unmarshal(data, &message)
			if err != nil {
				log.WithError(err).WithField("payload", string(data)).Warn("could not parse message from server")
				continue
			}

			// Send message
			select {
			case svc.incomingData <- DataPackage{
				Data:    data,
				Message: message,
			}:
				svc.messagesReceived++
			default:
				log.Warn("incoming data channel blocked")
			}
		}
	}
}

func (svc *websocketSvc) processMessages() {
	svc.log.Debug("starting message processor")
	for {
		select {
		// Time to exit
		case <-svc.stop:
			svc.log.Debug("stopping message processor")
			return

		// Process the message
		case pkg := <-svc.incomingData:
			svc.log.Debug("looking up handler to send data")
			svc.messageMtx.RLock()
			handler, ok := svc.messageHandlers[pkg.Type]
			svc.messageMtx.RUnlock()

			if !ok {
				svc.log.Warnf("unregistered message type %s receieved", pkg.Type)
				continue
			}

			svc.log.Debugf("sending message to '%s' handler", pkg.Type)
			select {
			case handler.Input() <- pkg:
			default:
				log.Warnf("%s handler input channel blocked", handler.Name())
			}
		}
	}
}

func (svc *websocketSvc) processWebsocketStream(stop <-chan bool) (incoming chan<- DataPackage) {
	// Setup the channels
	incoming = make(chan DataPackage, viper.GetInt("coinbase.websocket.streamProcessorBufferSize"))

	// Wrap everything in a single goroutine so we don't need mutexes
	go func(stop <-chan bool, incoming <-chan DataPackage) {
		var tickerChan, receivedChan, openChan, doneChan, matchChan chan<- DataPackage
		for {
			// Bail on stop first. We do it by itself in case incoming could be selected over stop because they're
			// both ready to return and we don't want it to parse anymore.
			select {
			case <-stop:
				return
			default:
			}

			select {
			// Bail on stop
			case <-stop:
				return

			// Process the incoming data
			case pkg := <-incoming:
				svc.log.Debugf("sending message to '%s' handler", pkg.Type)
				switch pkg.Type {
				case Subscriptions:
				case Ticker:
				case Received:
				case Open:
				case Done:
				case Match:
				default:
					svc.log.Warnf("unregistered message type %s receieved", pkg.Type)
					continue
				}

				log.Debug("looking up handler to send data")
				svc.messageMtx.RLock()
				handler, ok := svc.messageHandlers[pkg.Type]
				svc.messageMtx.RUnlock()

				if !ok {
					svc.log.Warnf("unregistered message type %s receieved", pkg.Type)
					continue
				}

				select {
				case handler.Input() <- pkg:
				default:
					log.Warnf("%s handler input channel blocked", handler.Name())
				}
			}
		}

	}(stop, incoming)
}

func (svc *websocketSvc) processSubscriptionsStream(stop <-chan bool) (incoming chan<- DataPackage) {
	incoming = make(chan DataPackage, viper.GetInt("coinbase.websocket.subscriptionsStreamProcessorBufferSize"))

	go func(stop <-chan bool, incoming chan DataPackage) {
		handler := new
	})(stop, incoming)
}