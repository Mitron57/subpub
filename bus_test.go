package subpub

import (
	"context"
	"errors"
	"testing"
	"time"
)

func dummy(_ any) {}

func TestBus_Publish(t *testing.T) {
	wait := make(chan struct{})
	handler := func(_ any) {
		wait <- struct{}{}
	}

	b := NewPubSub()
	_, err := b.Subscribe("subject", handler)
	if err != nil {
		t.Errorf("failed to subscribe: %v", err)
	}

	err = b.Publish("subject", "message")
	if err != nil {
		t.Errorf("failed to publish: %v", err)
	}

	select {
	case <-wait:
		// good
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for subscriber's handler to be called")
	}
}

func TestBus_PublishToClosedBus(t *testing.T) {
	b := NewPubSub()
	_ = b.Close(context.Background())

	err := b.Publish("subject", "message")
	if err == nil || !errors.Is(err, ErrClosedBus) {
		t.Errorf("should have failed to publish with error: %v, got: %v", ErrClosedBus, err)
	}
}

func TestBus_PublishToAbsentSubject(t *testing.T) {
	b := NewPubSub()

	err := b.Publish("subject", "message")
	if err == nil || !errors.Is(err, ErrNoSuchSubject) {
		t.Errorf("should have failed to publish with error: %v, got: %v", ErrNoSuchSubject, err)
	}
}

func TestBus_SubscribeToClosedBus(t *testing.T) {
	b := NewPubSub()

	err := b.Close(context.Background())
	if err != nil {
		t.Errorf("failed to close bus: %v", err)
	}

	_, err = b.Subscribe("subject", dummy)
	if err == nil || !errors.Is(err, ErrClosedBus) {
		t.Errorf("should have failed to subscribe with error: %v, got: %v", ErrClosedBus, err)
	}
}

func TestBus_SubscribeWithNilHandler(t *testing.T) {
	b := NewPubSub()
	_, err := b.Subscribe("subject", nil)
	if err == nil || !errors.Is(err, ErrNilHandler) {
		t.Errorf("should have failed to subscribe with error: %v, got: %v", ErrNilHandler, err)
	}
}

func TestBus_Close(t *testing.T) {
	b := NewPubSub()

	_, err := b.Subscribe("subject", dummy)
	if err != nil {
		t.Errorf("failed to subscribe: %v", err)
	}

	err = b.Close(context.Background())
	if err != nil {
		t.Errorf("failed to close bus: %v", err)
	}
}

func TestBus_CloseWithCancellation(t *testing.T) {
	b := NewPubSub()

	for _, topic := range []string{"topic1", "topic2", "topic3"} {
		_, err := b.Subscribe(topic, dummy)
		if err != nil {
			t.Errorf("failed to subscribe: %v", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := b.Close(ctx)

	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("should not have failed to cancel closing with error or fall with: %v, got: %v", context.Canceled, err)
	}

	err = b.Publish("topic1", "message")
	if err != nil && !errors.Is(err, ErrNoSuchSubject) {
		t.Errorf("should not have failed to publish with error or fall with: %v, got: %v", ErrNoSuchSubject, err)
	}
}
