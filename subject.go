package subpub

import (
	"sync"
)

type subject struct {
	subject      string
	messageQueue chan any
	busNotifier  chan<- string
	mx           sync.RWMutex
	handlers     map[int]MessageHandler
	handlerId    int
	closed       <-chan struct{}
}

func newTopic(topic string, notify chan<- string, closed <-chan struct{}) *subject {
	return &subject{
		subject:      topic,
		messageQueue: make(chan any, 256),
		busNotifier:  notify,
		handlers:     make(map[int]MessageHandler),
		closed:       closed,
	}
}

// add performs registering handler to its topic
func (t *subject) add(handler MessageHandler) int {
	t.mx.Lock()
	defer t.mx.Unlock()
	id := t.handlerId
	t.handlerId++
	t.handlers[id] = handler
	return id
}

// remove performs detaching the handler from topic by handlerId
func (t *subject) remove(handlerId int) {
	t.mx.Lock()
	defer t.mx.Unlock()
	delete(t.handlers, handlerId)
}

func (t *subject) empty() bool {
	t.mx.RLock()
	defer t.mx.RUnlock()
	return len(t.handlers) == 0
}

// notifyAllHandlers asynchronously executes all handlers in a map.
func (t *subject) notifyAllHandlers(msg any) {
	var wg sync.WaitGroup
	t.mx.RLock()
	wg.Add(len(t.handlers))
	for _, handler := range t.handlers {
		go func() {
			defer wg.Done()
			handler(msg)
		}()
	}
	wg.Wait()
	t.mx.RUnlock()
}

// listen waits for a message from bus.
// It should be called in a new goroutine due to its blocking code.
func (t *subject) listen() {
	defer func() {
		t.busNotifier <- t.subject
		close(t.messageQueue)
	}()

	for {
		select {
		case msg, ok := <-t.messageQueue:
			if !ok {
				return
			}
			t.notifyAllHandlers(msg)
		case <-t.closed:
			return
		}
	}
}
