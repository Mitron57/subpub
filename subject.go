package subpub

import (
	"sync"
)

// Limit for the channel
const messages = 256

type subject struct {
	subject      string
	handlers     map[int]MessageHandler
	handlerId    int
	messageQueue chan any
	closed       chan struct{}
}

func newTopic(topic string) *subject {
	return &subject{
		subject:      topic,
		messageQueue: make(chan any, messages), //any size is 16 bytes, so we can fit into virtual memory page, like unix pipe does
		handlers:     make(map[int]MessageHandler),
		closed:       make(chan struct{}),
	}
}

// add performs registering handler to its topic
func (t *subject) add(handler MessageHandler) int {
	id := t.handlerId
	t.handlerId++
	t.handlers[id] = handler
	return id
}

// remove performs detaching the handler from topic by handlerId
func (t *subject) remove(handlerId int) {
	delete(t.handlers, handlerId)
}

func (t *subject) empty() bool {
	return len(t.handlers) == 0
}

// notifyAllHandlers asynchronously executes all handlers in a map.
func (t *subject) notifyAllHandlers(msg any) {
	var wg sync.WaitGroup
	wg.Add(len(t.handlers))
	for _, handler := range t.handlers {
		go func() {
			defer wg.Done()
			handler(msg)
		}()
	}
	wg.Wait()
}

func (t *subject) close() {
	t.closed <- struct{}{}
}

// listen waits for a message from bus.
// It should be called in a new goroutine due to its blocking code.
func (t *subject) listen() {
	defer func() {
		close(t.messageQueue)
		close(t.closed)
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
