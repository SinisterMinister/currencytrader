package handlers

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type done struct {
	log      log.Entry
	stop     <-chan bool
	incoming chan types.DataPackage
	outgoing chan types.Done
}

func Done(stop <-chan bool) (msgHandler types.DoneMessageHandler, err error) {
	var handler done = done{
		incoming: make(chan types.DataPackage, viper.GetInt("coinbase.websocket.handlers.done.incomingBufferSize")),
		outgoing: make(chan types.Done, viper.GetInt("coinbase.websocket.handlers.done.outgoingBufferSize")),
		stop:     stop,
		log:      log.WithField("source", "coinbase.websocket.handlers.done"),
	}

	// Start handler
	go handler.start()

	// Return the handler
	return &handler, err
}

func (handler *done) Name() string {
	return DONE_HANDLER_NAME
}

func (handler *done) Input() chan<- types.DataPackage {
	return handler.incoming
}

func (handler *done) Output() <-chan types.Done {
	return handler.outgoing
}

func (handler *done) start() {
	for {
		select {
		// Kill switch flipped
		case <-handler.stop:
			return

		// Handle incoming data
		case pkg := <-handler.incoming:
			// Capture data
			var data types.Done
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
