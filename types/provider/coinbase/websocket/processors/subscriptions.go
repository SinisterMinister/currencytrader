package processors

import (
	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type Subscriptions interface {
	Input() chan<- types.Subscriptions
}

func NewSubscriptions(stop <-chan bool, svc subscriptionsService) (Subscriptions, error) {
	instance := subscriptions{
		log:   log.WithField("source", "coinbase.websocket.processors.log"),
		input: make(chan types.Subscriptions, viper.GetInt("coinbase.websocket.processors.subscriptions.inputBufferSize")),
		svc:   svc,
	}

	return &instance, nil
}

type subscriptionsService interface {
	types.Service
	UpdateSubscriptions(subs types.Subscriptions)
}

type subscriptions struct {
	log   log.Entry
	svc   subscriptionsService
	input chan types.Subscriptions
}

func (proc *subscriptions) Input() (input chan<- types.Subscriptions) {
	return
}
