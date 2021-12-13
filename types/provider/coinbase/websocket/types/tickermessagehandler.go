package types

type TickerMessageHandler interface {
	MessageHandler
	Output() <-chan Ticker
}
