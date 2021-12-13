package types

type OpenMessageHandler interface {
	MessageHandler
	Output() <-chan Open
}
