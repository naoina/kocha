package kocha

import (
	"os"
	"reflect"
	"strconv"

	"github.com/naoina/kocha/event"
)

// EventHandlerMap represents a map of event handlers.
type EventHandlerMap map[event.Queue]map[string]func(app *Application, args ...interface{}) error

// Evevnt represents the event.
type Event struct {
	// HandlerMap is a map of queue/handlers.
	HandlerMap EventHandlerMap

	// WorkersPerQueue is a number of workers per queue.
	// The default value is taken from GOMAXPROCS.
	// If value of GOMAXPROCS is invalid, set to 1.
	WorkersPerQueue int

	// ErrorHandler is the handler for error.
	// If you want to use your own error handler, please set to ErrorHandler.
	ErrorHandler func(err interface{})

	e   *event.Event
	app *Application
}

// Trigger emits the event.
// The name is an event name that is defined in e.HandlerMap.
// If args given, they will be passed to event handler that is defined in e.HandlerMap.
func (e *Event) Trigger(name string, args ...interface{}) error {
	return e.e.Trigger(name, args...)
}

func (e *Event) addHandler(name string, queueName string, handler func(app *Application, args ...interface{}) error) error {
	return e.e.AddHandler(name, queueName, func(args ...interface{}) error {
		return handler(e.app, args...)
	})
}

func (e *Event) build(app *Application) (*Event, error) {
	if e == nil {
		e = &Event{}
	}
	e.e = event.New()
	for queue, handlerMap := range e.HandlerMap {
		queueName := reflect.TypeOf(queue).Name()
		if err := e.e.RegisterQueue(queueName, queue); err != nil {
			return nil, err
		}
		for name, handler := range handlerMap {
			if err := e.addHandler(name, queueName, handler); err != nil {
				return nil, err
			}
		}
	}
	n := e.WorkersPerQueue
	if n < 1 {
		if n, _ = strconv.Atoi(os.Getenv("GOMAXPROCS")); n < 1 {
			n = 1
		}
	}
	e.e.SetWorkersPerQueue(n)
	e.e.ErrorHandler = e.ErrorHandler
	return e, nil
}

func (e *Event) start() {
	e.e.Start()
}

func (e *Event) stop() {
	e.e.Stop()
}
