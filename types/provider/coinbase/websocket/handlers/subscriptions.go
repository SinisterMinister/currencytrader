package handlers

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type subscriptions struct {
	log      log.Entry
	stop     <-chan bool
	incoming chan types.DataPackage
	outgoing chan types.Subscriptions
}

func Subscriptions(stop <-chan bool) (msgHandler types.SubscriptionsMessageHandler, err error) {
	var handler subscriptions = subscriptions{
		incoming: make(chan types.DataPackage, viper.GetInt("coinbase.websocket.handlers.subscriptions.incomingBufferSize")),
		outgoing: make(chan types.Subscriptions, viper.GetInt("coinbase.websocket.handlers.subscriptions.outgoingBufferSize")),
		stop:     stop,
		log:      log.WithField("source", "coinbase.websocket.handlers.subscriptions"),
	}

	// Start handler
	go handler.start()

	// Return the handler
	return &handler, err
}

func (handler *subscriptions) Name() string {
	return SUBSCRIPTION_HANDLER_NAME
}

func (handler *subscriptions) Input() chan<- types.DataPackage {
	return handler.incoming
}

func (handler *subscriptions) Output() <-chan types.Subscriptions {
	return handler.outgoing
}

func (handler *subscriptions) start() {
	for {
		select {
		// Kill switch flipped
		case <-handler.stop:
			return

		// Handle incoming data
		case pkg := <-handler.incoming:
			// Capture data
			var data types.Subscriptions
			handler.log.Debugf("handling %s payload", handler.Name())
			e := json.Unmarshal(pkg.Data, &data)
			if e != nil {
				handler.log.Errorf("could not parse data for %s", handler.Name())
				continue
			}
			handler.outgoing <- data
		}
	}
}
