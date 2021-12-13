package types

type Service interface {
	Subscriptions() Subscriptions
	Subscribe(req Subscribe) (err error)
	Unsubscribe(req Subscribe) (err error)
}
