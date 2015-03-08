package memory

import "github.com/naoina/kocha/event"

// EventQueue implements the Queue interface.
// This doesn't require the external storages such as Redis.
// Note that EventQueue isn't persistent, this means that queued data may be
// lost by crash, shutdown or status of not running.
// If you want to do use a persistent queue, please use another Queue
// implementation that supports persistence.
// Also queue won't be shared between different servers but will be shared
// between other workers in same server.
type EventQueue struct {
	c    chan string
	done chan struct{}
	exit chan struct{}
}

// New returns a new EventQueue.
func (q *EventQueue) New(n int) event.Queue {
	if q.c == nil {
		q.c = make(chan string, n)
	}
	if q.done == nil {
		q.done = make(chan struct{})
	}
	if q.exit == nil {
		q.exit = make(chan struct{})
	}
	return &EventQueue{
		c:    q.c,
		done: q.done,
		exit: q.exit,
	}
}

// Enqueue adds data to queue.
func (q *EventQueue) Enqueue(data string) error {
	q.c <- data
	return nil
}

// Dequeue returns the data that fetch from queue.
func (q *EventQueue) Dequeue() (data string, err error) {
	select {
	case data = <-q.c:
		return data, nil
	case <-q.done:
		defer func() {
			q.exit <- struct{}{}
		}()
		return "", event.ErrDone
	}
}

// Stop wait for Dequeue to complete then will stop a queue.
func (q *EventQueue) Stop() {
	defer func() {
		q.c = nil
		q.done = nil
		q.exit = nil
	}()
	q.done <- struct{}{}
	<-q.exit
}
