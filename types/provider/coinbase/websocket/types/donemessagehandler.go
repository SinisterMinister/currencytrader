package types

type DoneMessageHandler interface {
	MessageHandler
	Output() <-chan Done
}
