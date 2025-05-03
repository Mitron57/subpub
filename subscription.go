package subPub

type Subscription interface {
	// Unsubscribe will remove interest in the current subject subscription is for.
	Unsubscribe()
}

type subscriber struct {
	bus     *bus
	subject string
	id      int
}

func newSubscription(bus *bus, subject string, id int) Subscription {
	return &subscriber{bus: bus, subject: subject, id: id}
}

// Unsubscribe detaches subscriber from the bus
func (s *subscriber) Unsubscribe() {
	s.bus.detach(s.subject, s.id)
}
