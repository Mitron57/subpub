package subpub

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func dummy(_ any) {}

func TestBus_Publish(t *testing.T) {
	wait := make(chan struct{})
	handler := func(_ any) {
		wait <- struct{}{}
	}

	b := NewPubSub()
	_, err := b.Subscribe("subject", handler)
	require.NoError(t, err)

	err = b.Publish("subject", "message")
	require.NoError(t, err)

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
	require.ErrorIs(t, err, ErrClosedBus)
}

func TestBus_PublishToAbsentSubject(t *testing.T) {
	b := NewPubSub()

	err := b.Publish("subject", "message")
	require.ErrorIs(t, err, ErrNoSuchSubject)
}

func TestBus_SubscribeToClosedBus(t *testing.T) {
	b := NewPubSub()

	err := b.Close(context.Background())
	require.NoError(t, err)

	_, err = b.Subscribe("subject", dummy)
	require.ErrorIs(t, err, ErrClosedBus)
}

func TestBus_SubscribeWithNilHandler(t *testing.T) {
	b := NewPubSub()
	_, err := b.Subscribe("subject", nil)
	require.ErrorIs(t, err, ErrNilHandler)
}

func TestBus_Close(t *testing.T) {
	b := NewPubSub()

	for _, topic := range []string{"topic1", "topic2", "topic3"} {
		_, err := b.Subscribe(topic, dummy)
		require.NoError(t, err)
	}

	err := b.Close(context.Background())
	require.NoError(t, err)
}

func TestBus_CloseWithCancellation(t *testing.T) {
	b := NewPubSub()

	for _, topic := range []string{"topic1", "topic2", "topic3"} {
		_, err := b.Subscribe(topic, dummy)
		require.NoError(t, err)
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
