package types

type ChangeMessageHandler interface {
	MessageHandler
	Output() <-chan Change
}
