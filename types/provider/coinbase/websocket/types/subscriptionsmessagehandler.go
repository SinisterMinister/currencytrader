package types

type SubscriptionsMessageHandler interface {
	MessageHandler
	Output() <-chan Subscriptions
}
