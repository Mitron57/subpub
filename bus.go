package subpub

import (
	"context"
	"sync"
)

// MessageHandler is a callback function that processes messages delivered to subscribers.
type MessageHandler func(msg any)

type SubPub interface {
	// Subscribe creates an asynchronous queue subscriber on the given object.
	Subscribe(subject string, handler MessageHandler) (Subscription, error)
	// Publish publishes the msg argument to the given subject.
	Publish(subject string, msg any) error
	// Close will shut down sub-pub system.
	// May be blocked by data delivery until the context is cancelled.
	Close(ctx context.Context) error
}

type bus struct {
	mx        sync.RWMutex
	topics    map[string]*subject
	listener  chan string
	closeChan chan struct{}
	closed    bool
}

func NewPubSub() SubPub {
	return &bus{
		topics:    make(map[string]*subject),
		listener:  make(chan string),
		closeChan: make(chan struct{}),
	}
}

// Detaches subscriber from bus.
func (b *bus) detach(subject string, subscriptionId int) {
	b.mx.Lock()
	defer b.mx.Unlock()
	topic, ok := b.topics[subject]

	if !ok {
		return
	}
	topic.remove(subscriptionId)
	if topic.empty() {
		delete(b.topics, subject)
	}
}

// Subscribe registers new subscriber to the subject.
// If there's no such subject, then the new topic will be created.
// ErrNilHandler and ErrClosedBus are returned respectively to their reason.
func (b *bus) Subscribe(subject string, handler MessageHandler) (Subscription, error) {
	if handler == nil {
		return nil, ErrNilHandler
	}

	b.mx.Lock()
	defer b.mx.Unlock()
	if b.closed {
		return nil, ErrClosedBus
	}

	topic, ok := b.topics[subject]
	if !ok {
		topic = newTopic(subject, b.listener, b.closeChan)
		b.topics[subject] = topic
		go topic.listen()
	}

	id := topic.add(handler)
	return newSubscription(b, subject, id), nil
}

// Publish method asynchronously delivers message to the subject's handlers
// ErrNoSuchSubject is returned, when there's no such subject (interesting, isn't it?),
// ErrClosedBus is returned when there's attempt to publish on a closed (or closing) bus.
func (b *bus) Publish(subject string, msg any) error {
	b.mx.RLock()
	defer b.mx.RUnlock()
	if b.closed {
		return ErrClosedBus
	}
	topic, ok := b.topics[subject]

	if !ok {
		return ErrNoSuchSubject
	}

	topic.messageQueue <- msg
	return nil
}

// Close performs bus closing. Topics are closed in undefined order.
// It may be cancelled via context, but be aware that the bus is in closed state during closing
// and closed topics are not restored.
func (b *bus) Close(ctx context.Context) error {
	b.mx.Lock()
	defer b.mx.Unlock()
	if b.closed {
		return ErrClosedBus
	}
	b.closed = true
	topicsToClose := len(b.topics)

	if topicsToClose == 0 {
		close(b.listener)
		close(b.closeChan)
		return nil
	}

	for range topicsToClose {
		b.closeChan <- struct{}{}

		select {
		case <-ctx.Done():
			topic := <-b.listener
			delete(b.topics, topic)
			b.closed = false
			return ctx.Err()
		case topic := <-b.listener:
			delete(b.topics, topic)
		}
	}

	close(b.listener)
	close(b.closeChan)
	return nil
}
