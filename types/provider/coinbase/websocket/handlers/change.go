package handlers

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type change struct {
	log      log.Entry
	stop     <-chan bool
	incoming chan types.DataPackage
	outgoing chan types.Change
}

func Change(stop <-chan bool) (msgHandler types.ChangeMessageHandler, err error) {
	var handler change = change{
		incoming: make(chan types.DataPackage, viper.GetInt("coinbase.websocket.handlers.change.incomingBufferSize")),
		outgoing: make(chan types.Change, viper.GetInt("coinbase.websocket.handlers.change.outgoingBufferSize")),
		stop:     stop,
		log:      log.WithField("source", "coinbase.websocket.handlers.change"),
	}

	// Start handler
	go handler.start()

	// Return the handler
	return &handler, err
}

func (handler *change) Name() string {
	return CHANGE_HANDLER_NAME
}

func (handler *change) Input() chan<- types.DataPackage {
	return handler.incoming
}

func (handler *change) Output() <-chan types.Change {
	return handler.outgoing
}

func (handler *change) start() {
	for {
		select {
		// Kill switch flipped
		case <-handler.stop:
			return

		// Handle incoming data
		case pkg := <-handler.incoming:
			// Capture data
			var data types.Change
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
