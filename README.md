# subpub
Simple asynchronous publisher-subscriber bus

Exposed interface
```go
type SubPub interface {
	// Subscribe creates an asynchronous queue subscriber on the given object.
	Subscribe(subject string, handler MessageHandler) (Subscription, error)
	// Publish publishes the msg argument to the given subject.
	Publish(subject string, msg any) error
	// Close will shut down sub-pub system.
	// May be blocked by data delivery until the context is cancelled.
	Close(ctx context.Context) error
}
```
Code example
```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mitron57/subpub"
)

func main() {
	// Create a new bus
	bus := subpub.NewPubSub()

	// Subscribe to "news" topic
	sub, err := bus.Subscribe("news", func(msg any) {
		fmt.Println("Received message:", msg)
	})
	if err != nil {
		panic(err)
	}

	// Publish a message to "news" topic
	err = bus.Publish("news", "Hello, World!")
	if err != nil {
		panic(err)
	}

	// Give goroutines time to process (in production you'd probably wait or block more elegantly)
	time.Sleep(500 * time.Millisecond)

	// Unsubscribe
	sub.Unsubscribe()

	// Close the bus with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := bus.Close(ctx); err != nil {
		panic(err)
	}
}
```
