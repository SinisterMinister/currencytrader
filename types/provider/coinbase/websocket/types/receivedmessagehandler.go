package types

type ReceivedMessageHandler interface {
	MessageHandler
	Output() <-chan Received
}
