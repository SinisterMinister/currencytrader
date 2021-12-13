package types

type MatchMessageHandler interface {
	MessageHandler
	Output() <-chan Match
}
