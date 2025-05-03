package subPub

import (
    "testing"
    "time"
)

func TestSubscriber_Unsubscribe(t *testing.T) {
    wait := make(chan struct{})
    handler := func(_ any) {
        wait <- struct{}{}
    }

    b := NewPubSub()

    sub, err := b.Subscribe("subject", handler)
    if err != nil {
        t.Errorf("failed to subscribe: %e", err)
    }

    sub.Unsubscribe()

    err = b.Publish("subject", 1)
    if err != nil {
        t.Errorf("failed to publish: %e", err)
    }

    select {
    case <-wait:
        t.Error("unsubscribe should have done")
    case <-time.After(time.Second):
    }
}
