package subpub

import "errors"

var (
	ErrClosedBus     = errors.New("bus is closed")
	ErrNilHandler    = errors.New("handler is nil")
	ErrNoSuchSubject = errors.New("no such subject")
)
