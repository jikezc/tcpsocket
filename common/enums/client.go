package enums

type ClientStatusEM int

const (
	ClientStatusConnecting ClientStatusEM = iota
	ClientStatusConnected
	ClientStatusDisConnected
	ClientStatusError
	ClientStatusWaiting
)
