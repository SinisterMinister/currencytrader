package handlers

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type received struct {
	log      log.Entry
	stop     <-chan bool
	incoming chan types.DataPackage
	outgoing chan types.Received
}

func Received(stop <-chan bool) (msgHandler types.ReceivedMessageHandler, err error) {
	var handler received = received{
		incoming: make(chan types.DataPackage, viper.GetInt("coinbase.websocket.handlers.received.incomingBufferSize")),
		outgoing: make(chan types.Received, viper.GetInt("coinbase.websocket.handlers.received.outgoingBufferSize")),
		stop:     stop,
		log:      log.WithField("source", "coinbase.websocket.handlers.received"),
	}

	// Start handler
	go handler.start()

	// Return the handler
	return &handler, err
}

func (handler *received) Name() string {
	return RECEIVED_HANDLER_NAME
}

func (handler *received) Input() chan<- types.DataPackage {
	return handler.incoming
}

func (handler *received) Output() <-chan types.Received {
	return handler.outgoing
}

func (handler *received) start() {
	for {
		select {
		// Kill switch flipped
		case <-handler.stop:
			return

		// Handle incoming data
		case pkg := <-handler.incoming:
			// Capture data
			var data types.Received
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
