package handlers

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type open struct {
	log      log.Entry
	stop     <-chan bool
	incoming chan types.DataPackage
	outgoing chan types.Open
}

func Open(stop <-chan bool) (msgHandler types.OpenMessageHandler, err error) {
	var handler open = open{
		incoming: make(chan types.DataPackage, viper.GetInt("coinbase.websocket.handlers.open.incomingBufferSize")),
		outgoing: make(chan types.Open, viper.GetInt("coinbase.websocket.handlers.open.outgoingBufferSize")),
		stop:     stop,
		log:      log.WithField("source", "coinbase.websocket.handlers.open"),
	}

	// Start handler
	go handler.start()

	// Return the handler
	return &handler, err
}

func (handler *open) Name() string {
	return OPEN_HANDLER_NAME
}

func (handler *open) Input() chan<- types.DataPackage {
	return handler.incoming
}

func (handler *open) Output() <-chan types.Open {
	return handler.outgoing
}

func (handler *open) start() {
	for {
		select {
		// Kill switch flipped
		case <-handler.stop:
			return

		// Handle incoming data
		case pkg := <-handler.incoming:
			// Capture data
			var data types.Open
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
