package event

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	queueName = "fakeQueue"
)

var stopped []struct{}

type fakeQueue struct {
	c    chan string
	done chan struct{}
}

func (q *fakeQueue) New(n int) Queue {
	return q
}

func (q *fakeQueue) Enqueue(data string) error {
	q.c <- data
	return nil
}

func (q *fakeQueue) Dequeue() (string, error) {
	select {
	case data := <-q.c:
		return data, nil
	case <-q.done:
		return "", nil
	}
}

func (q *fakeQueue) Stop() {
	stopped = append(stopped, struct{}{})
	q.done <- struct{}{}
}

func TestAddHandler(t *testing.T) {
	handlerName := "testAddHandler"
	if err := AddHandler(handlerName, "unknownQueue", func(args ...interface{}) error {
		return nil
	}); err == nil {
		t.Errorf("AddHandler(%q, %q, func) => nil, want error", handlerName, "unknownQueue")
	}
	if err := AddHandler(handlerName, queueName, func(args ...interface{}) error {
		return nil
	}); err != nil {
		t.Errorf("AddHandler(%q, %q, func) => %#v, want nil", handlerName, queueName, err)
	}
	if err := AddHandler(handlerName, queueName, func(args ...interface{}) error {
		return nil
	}); err != nil {
		t.Errorf("AddHandler(%q, %q, func) => %#v, want nil", handlerName, queueName, err)
	}
}

func TestTrigger(t *testing.T) {
	handlerName := "unknownHandler"
	if err := Trigger(handlerName); err == nil {
		t.Errorf("Trigger(%q) => nil, want error", handlerName)
	}

	handlerName = "testTrigger"
	var actual string
	timer := make(chan struct{})
	if err := AddHandler(handlerName, queueName, func(args ...interface{}) error {
		defer func() {
			timer <- struct{}{}
		}()
		actual += fmt.Sprintf("|call %s(%v)", handlerName, args)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2; i++ {
		if err := Trigger(handlerName); err != nil {
			t.Errorf("Trigger(%q) => %#v, want nil", handlerName, err)
		}
		select {
		case <-timer:
		case <-time.After(3 * time.Second):
			t.Fatalf("Trigger(%q) has try to call handler but hasn't been called within 3 seconds", handlerName)
		}
		expected := strings.Repeat("|call testTrigger([])", i)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Trigger(%q) has try to call handler, actual => %#v, want %#v", handlerName, actual, expected)
		}
	}

	handlerName = "testTriggerWithArgs"
	actual = ""
	if err := AddHandler(handlerName, queueName, func(args ...interface{}) error {
		defer func() {
			timer <- struct{}{}
		}()
		actual += fmt.Sprintf("|call %s(%v)", handlerName, args)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2; i++ {
		if err := Trigger(handlerName, 1, true, "arg"); err != nil {
			t.Errorf("Trigger(%q) => %#v, want nil", handlerName, err)
		}
		select {
		case <-timer:
		case <-time.After(3 * time.Second):
			t.Fatalf("Trigger(%q) has try to call handler but hasn't been called within 3 seconds", handlerName)
		}
		expected := strings.Repeat("|call testTriggerWithArgs([1 true arg])", i)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Trigger(%q) has try to call handler, actual => %#v, want %#v", handlerName, actual, expected)
		}
	}

	handlerName = "testTriggerWithMultipleHandlers"
	actual = ""
	actual2 := ""
	timer2 := make(chan struct{})
	if err := AddHandler(handlerName, queueName, func(args ...interface{}) error {
		defer func() {
			timer <- struct{}{}
		}()
		actual += fmt.Sprintf("|call1 %s(%v)", handlerName, args)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := AddHandler(handlerName, queueName, func(args ...interface{}) error {
		defer func() {
			timer2 <- struct{}{}
		}()
		actual2 += fmt.Sprintf("|call2 %s(%v)", handlerName, args)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2; i++ {
		if err := Trigger(handlerName); err != nil {
			t.Errorf("Trigger(%q) => %#v, want nil", handlerName, err)
		}
		select {
		case <-timer:
		case <-time.After(3 * time.Second):
			t.Fatalf("Trigger(%q) has try to call handler but hasn't been called within 3 seconds", handlerName)
		}
		expected := strings.Repeat("|call1 testTriggerWithMultipleHandlers([])", i)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Trigger(%q) has try to call handler, actual => %#v, want %#v", handlerName, actual, expected)
		}
		select {
		case <-timer2:
		case <-time.After(3 * time.Second):
			t.Fatalf("Trigger(%q) has try to call handler but hasn't been called within 3 seconds", handlerName)
		}
		expected = strings.Repeat("|call2 testTriggerWithMultipleHandlers([])", i)
		if !reflect.DeepEqual(actual2, expected) {
			t.Errorf("Trigger(%q) has try to call handler, actual => %#v, want %#v", handlerName, actual2, expected)
		}
	}
}

func TestStop(t *testing.T) {
	stopped = nil
	actual := len(stopped)
	expected := 0
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("len(stopped) before Stop => %#v, want %#v", actual, expected)
	}
	Stop()
	actual = len(stopped)
	expected = 1
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("len(stopped) after Stop => %#v, want %#v", actual, expected)
	}
}

func TestErrorHandler(t *testing.T) {
	handlerName := "testErrorHandler"
	expected := fmt.Errorf("testErrorHandlerError")
	if err := AddHandler(handlerName, queueName, func(args ...interface{}) error {
		return expected
	}); err != nil {
		t.Fatal(err)
	}
	origErrorHandler := ErrorHandler
	defer func() {
		ErrorHandler = origErrorHandler
	}()
	called := make(chan struct{})
	ErrorHandler = func(err interface{}) {
		if !reflect.DeepEqual(err, expected) {
			t.Errorf("ErrorHandler called with %#v, want %#v", err, expected)
		}
		called <- struct{}{}
	}
	if err := Trigger(handlerName); err != nil {
		t.Fatal(err)
	}
	select {
	case <-called:
	case <-time.After(3 * time.Second):
		t.Errorf("ErrorHandler hasn't been called within 3 seconds")
	}
}

func init() {
	RegisterQueue(queueName, &fakeQueue{c: make(chan string), done: make(chan struct{})})
	Start()
}
