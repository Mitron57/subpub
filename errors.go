package subPub

import "errors"

var (
    ErrClosedBus     = errors.New("bus is closed")
    ErrNilHandler    = errors.New("nil handler")
    ErrNoSuchSubject = errors.New("no such subject")
)
