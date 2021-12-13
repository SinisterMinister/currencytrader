package types

type MessageHandler interface {
	Name() string
	Input() chan<- DataPackage
}
