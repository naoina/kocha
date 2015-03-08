package event

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// DefaultEvent is the default event and is used by AddHandler, Trigger, AddQueue, Start and Stop.
	DefaultEvent = New()

	// ErrDone represents that a queue is finished.
	ErrDone = errors.New("queue is done")

	// ErrNotExist is passed to ErrorHandler if handler not exists.
	ErrNotExist = errors.New("handler not exist")
)

// AddHandler is shorthand of the DefaultEvent.AddHandler.
func AddHandler(name string, queueName string, handler func(args ...interface{}) error) error {
	return DefaultEvent.AddHandler(name, queueName, handler)
}

// Trigger is shorthand of the DefaultEvent.Trigger.
func Trigger(name string, args ...interface{}) error {
	return DefaultEvent.Trigger(name, args...)
}

// RegisterQueue is shorthand of the DefaultEvent.RegisterQueue.
func RegisterQueue(name string, queue Queue) error {
	return DefaultEvent.RegisterQueue(name, queue)
}

// Start is shorthand of the DefaultEvent.Start.
func Start() {
	DefaultEvent.Start()
}

// Stop is shorthand of the DefaultEvent.Stop.
func Stop() {
	DefaultEvent.Stop()
}

// Event represents an Event.
type Event struct {
	// ErrorHandler is the error handler.
	// If you want to use your own error handler, set ErrorHandler.
	ErrorHandler func(err interface{})

	workersPerQueue int
	queues          map[string]Queue
	handlerQueues   map[string]map[string][]handlerFunc
	workers         []*worker
	wg              struct{ enqueue, dequeue sync.WaitGroup }
}

// New returns a new Event.
func New() *Event {
	return &Event{
		workersPerQueue: 1,
		queues:          make(map[string]Queue),
		handlerQueues:   make(map[string]map[string][]handlerFunc),
		wg:              struct{ enqueue, dequeue sync.WaitGroup }{},
	}
}

// AddHandler adds handlers that related to name and queue.
// The name is an event name such as "log.error" that will be used for Trigger.
// The queueName is a name of queue registered by RegisterQueue in advance.
// If you add handler by name that has already been added, handler will associated
// to that name additionally.
// If queue of queueName still hasn't been registered, it returns error.
func (e *Event) AddHandler(name string, queueName string, handler func(args ...interface{}) error) error {
	queue := e.queues[queueName]
	if queue == nil {
		return fmt.Errorf("kocha: event: queue `%s' isn't registered", queueName)
	}
	if _, exist := e.handlerQueues[name]; !exist {
		e.handlerQueues[name] = make(map[string][]handlerFunc)
	}
	hq := e.handlerQueues[name]
	hq[queueName] = append(hq[queueName], handler)
	return nil
}

// Trigger emits the event.
// The name is an event name. It must be added in advance using AddHandler.
// If Trigger called by not added name, it returns error.
// If args are given, they will be passed to handlers added by AddHandler.
func (e *Event) Trigger(name string, args ...interface{}) error {
	hq, exist := e.handlerQueues[name]
	if !exist {
		return fmt.Errorf("kocha: event: handler `%s' isn't added", name)
	}
	e.triggerAll(hq, name, args...)
	return nil
}

// RegisterQueue makes a background queue available by the provided name.
// If queue is already registerd or if queue nil, it panics.
func (e *Event) RegisterQueue(name string, queue Queue) error {
	if queue == nil {
		return fmt.Errorf("kocha: event: Register queue is nil")
	}
	if _, exist := e.queues[name]; exist {
		return fmt.Errorf("kocha: event: Register queue `%s' is already registered", name)
	}
	e.queues[name] = queue
	return nil
}

func (e *Event) triggerAll(hq map[string][]handlerFunc, name string, args ...interface{}) {
	e.wg.enqueue.Add(len(hq))
	for queueName := range hq {
		queue := e.queues[queueName]
		go func() {
			defer e.wg.enqueue.Done()
			defer func() {
				if err := recover(); err != nil {
					if e.ErrorHandler != nil {
						e.ErrorHandler(err)
					}
				}
			}()
			if err := e.enqueue(queue, payload{name, args}); err != nil {
				panic(err)
			}
		}()
	}
}

// alias.
type handlerFunc func(args ...interface{}) error

func (e *Event) enqueue(queue Queue, pld payload) error {
	var data string
	if err := pld.encode(&data); err != nil {
		return err
	}
	return queue.Enqueue(data)
}

// Start starts background event workers.
// By default, workers per queue is 1. To set the workers per queue, use
// SetWorkersPerQueue before Start calls.
func (e *Event) Start() {
	for name, queue := range e.queues {
		for i := 0; i < e.workersPerQueue; i++ {
			worker := e.newWorker(name, queue.New(e.workersPerQueue))
			e.workers = append(e.workers, worker)
			go worker.start()
		}
	}
}

// SetWorkersPerQueue sets the number of workers per queue.
// It must be called before Start calls.
func (e *Event) SetWorkersPerQueue(n int) {
	if n < 1 {
		n = 1
	}
	e.workersPerQueue = n
}

// Stop wait for all workers to complete.
func (e *Event) Stop() {
	e.wg.enqueue.Wait()
	defer func() {
		e.workers = nil
	}()
	defer e.wg.dequeue.Wait()
	for _, worker := range e.workers {
		worker.stop()
	}
}

type worker struct {
	queueName string
	queue     Queue
	e         *Event
}

func (e *Event) newWorker(queueName string, queue Queue) *worker {
	return &worker{
		queueName: queueName,
		queue:     queue,
		e:         e,
	}
}

func (w *worker) start() {
	var done bool
	for !done {
		func() {
			defer func() {
				if err := recover(); err != nil {
					if w.e.ErrorHandler != nil {
						w.e.ErrorHandler(err)
					}
				}
			}()
			if err := w.run(); err != nil {
				if err == ErrDone {
					done = true
					return
				}
				panic(err)
			}
		}()
	}
}

func (w *worker) run() (err error) {
	w.e.wg.dequeue.Add(1)
	defer w.e.wg.dequeue.Done()
	pld, err := w.dequeue()
	if err != nil {
		return err
	}
	hq, exist := w.e.handlerQueues[pld.Name]
	if !exist {
		return ErrNotExist
	}
	w.runAll(hq, pld)
	return nil
}

func (w *worker) runAll(hq map[string][]handlerFunc, pld payload) {
	for queueName, handlers := range hq {
		if w.queueName != queueName {
			continue
		}
		w.e.wg.dequeue.Add(len(handlers))
		for _, h := range handlers {
			go func(handler handlerFunc) {
				defer w.e.wg.dequeue.Done()
				if err := handler(pld.Args...); err != nil {
					if w.e.ErrorHandler != nil {
						w.e.ErrorHandler(err)
					}
				}
			}(h)
		}
	}
}

func (w *worker) dequeue() (pld payload, err error) {
	data, err := w.queue.Dequeue()
	if err != nil {
		return pld, err
	}
	if err := pld.decode(data); err != nil {
		return pld, err
	}
	return pld, nil
}

func (w *worker) stop() {
	w.queue.Stop()
}

// Queue is the interface that must be implemeted by background event queue.
type Queue interface {
	// New returns a new Queue to launch the workers.
	// You can use an argument n as a hint when you create a new queue.
	// n is the number of workers per queue.
	New(n int) Queue

	// Enqueue add data to the queue.
	Enqueue(data string) error

	// Dequeue returns the data that fetch from the queue.
	// It will return ErrDone as err when Stop is called.
	Dequeue() (data string, err error)

	// Stop wait for Enqueue and/or Dequeue to complete then will stop a queue.
	Stop()
}
