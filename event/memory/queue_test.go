package memory

import (
	"testing"
	"time"

	"github.com/naoina/kocha/event"
)

func TestMemoryQueue(t *testing.T) {
	handlerName := "testMemoryQueueHandler"
	called := make(chan struct{})
	if err := event.AddHandler(handlerName, "memory", func(args ...interface{}) error {
		called <- struct{}{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := event.Trigger(handlerName); err != nil {
		t.Errorf("event.Trigger(%q) => %#v, want nil", handlerName, err)
	}
	select {
	case <-called:
	case <-time.After(3 * time.Second):
		t.Errorf("event.Trigger(%q) has try to call handler but hasn't been called within 3 seconds", handlerName)
	}
}

func init() {
	event.Start()
}
