package event_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/naoina/kocha/event"
)

const (
	queueName = "fakeQueue"
)

var stopped []struct{}

type fakeQueue struct {
	c    chan string
	done chan struct{}
}

func (q *fakeQueue) New(n int) event.Queue {
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
		return "", event.ErrDone
	}
}

func (q *fakeQueue) Stop() {
	stopped = append(stopped, struct{}{})
	close(q.done)
}

func TestDefaultEvent(t *testing.T) {
	actual := event.DefaultEvent
	expect := event.New()
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`DefaultEvent => %#v; want %#v`, actual, expect)
	}
}

func TestEvent_AddHandler(t *testing.T) {
	e := event.New()
	e.RegisterQueue(queueName, &fakeQueue{c: make(chan string), done: make(chan struct{})})

	handlerName := "testAddHandler"
	for _, v := range []struct {
		queueName string
		expect    error
	}{
		{"unknownQueue", fmt.Errorf("kocha: event: queue `unknownQueue' isn't registered")},
		{queueName, nil},
		{queueName, nil}, // testcase for override.
	} {
		actual := e.AddHandler(handlerName, v.queueName, func(args ...interface{}) error {
			return nil
		})
		expect := v.expect
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf("AddHandler(%q, %q, func) => %#v, want %#v", handlerName, v.queueName, actual, expect)
		}
	}
}

func TestEvent_Trigger(t *testing.T) {
	e := event.New()
	e.RegisterQueue(queueName, &fakeQueue{c: make(chan string), done: make(chan struct{})})
	e.Start()
	defer e.Stop()

	handlerName := "unknownHandler"
	var expect interface{} = fmt.Errorf("kocha: event: handler `unknownHandler' isn't added")
	if err := e.Trigger(handlerName); err == nil {
		t.Errorf("Trigger(%q) => %#v, want %#v", handlerName, err, expect)
	}

	handlerName = "testTrigger"
	var actual string
	timer := make(chan struct{})
	if err := e.AddHandler(handlerName, queueName, func(args ...interface{}) error {
		defer func() {
			timer <- struct{}{}
		}()
		actual += fmt.Sprintf("|call %s(%v)", handlerName, args)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2; i++ {
		if err := e.Trigger(handlerName); err != nil {
			t.Errorf("Trigger(%#v) => %#v, want nil", handlerName, err)
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
	if err := e.AddHandler(handlerName, queueName, func(args ...interface{}) error {
		defer func() {
			timer <- struct{}{}
		}()
		actual += fmt.Sprintf("|call %s(%v)", handlerName, args)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2; i++ {
		if err := e.Trigger(handlerName, 1, true, "arg"); err != nil {
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
	if err := e.AddHandler(handlerName, queueName, func(args ...interface{}) error {
		defer func() {
			timer <- struct{}{}
		}()
		actual += fmt.Sprintf("|call1 %s(%v)", handlerName, args)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := e.AddHandler(handlerName, queueName, func(args ...interface{}) error {
		defer func() {
			timer2 <- struct{}{}
		}()
		actual2 += fmt.Sprintf("|call2 %s(%v)", handlerName, args)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2; i++ {
		if err := e.Trigger(handlerName); err != nil {
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

func TestEvent_RegisterQueue(t *testing.T) {
	e := event.New()
	for _, v := range []struct {
		name   string
		queue  event.Queue
		expect error
	}{
		{"test_queue", nil, fmt.Errorf("kocha: event: Register queue is nil")},
		{"test_queue", &fakeQueue{}, nil},
		{"test_queue", &fakeQueue{}, fmt.Errorf("kocha: event: Register queue `test_queue' is already registered")},
	} {
		actual := e.RegisterQueue(v.name, v.queue)
		expect := v.expect
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`Event.RegisterQueue(%q, %#v) => %#v; want %#v`, v.name, v.queue, actual, expect)
		}
	}
}

func TestEvent_Stop(t *testing.T) {
	e := event.New()
	e.RegisterQueue(queueName, &fakeQueue{c: make(chan string), done: make(chan struct{})})
	e.Start()
	defer e.Stop()

	stopped = nil
	actual := len(stopped)
	expected := 0
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("len(stopped) before Stop => %#v, want %#v", actual, expected)
	}
	e.Stop()
	actual = len(stopped)
	expected = 1
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("len(stopped) after Stop => %#v, want %#v", actual, expected)
	}
}

func TestEvent_ErrorHandler(t *testing.T) {
	e := event.New()
	e.RegisterQueue(queueName, &fakeQueue{c: make(chan string), done: make(chan struct{})})
	e.Start()
	defer e.Stop()

	handlerName := "testErrorHandler"
	expected := fmt.Errorf("testErrorHandlerError")
	if err := e.AddHandler(handlerName, queueName, func(args ...interface{}) error {
		return expected
	}); err != nil {
		t.Fatal(err)
	}
	called := make(chan struct{})
	e.ErrorHandler = func(err interface{}) {
		if !reflect.DeepEqual(err, expected) {
			t.Errorf("ErrorHandler called with %#v, want %#v", err, expected)
		}
		called <- struct{}{}
	}
	if err := e.Trigger(handlerName); err != nil {
		t.Fatal(err)
	}
	select {
	case <-called:
	case <-time.After(3 * time.Second):
		t.Errorf("ErrorHandler hasn't been called within 3 seconds")
	}
}
