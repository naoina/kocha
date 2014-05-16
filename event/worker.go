package event

import "sync"

type worker struct {
	queueName string
	queue     Queue
	m         map[string]handlerQueueName
	wg        *sync.WaitGroup
}

func newWorker(queueName string, queue Queue, m map[string]handlerQueueName, wg *sync.WaitGroup) *worker {
	return &worker{
		queueName: queueName,
		queue:     queue,
		m:         m,
		wg:        wg,
	}
}

func (w *worker) start() {
	var done bool
	for !done {
		func() {
			defer func() {
				if err := recover(); err != nil {
					ErrorHandler(err)
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
	w.wg.Add(1)
	defer w.wg.Done()
	pld, err := w.dequeue()
	if err != nil {
		return err
	}
	hq, exist := w.m[pld.Name]
	if !exist {
		return ErrNotExist
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		if err := hq.handler(pld.Args...); err != nil {
			ErrorHandler(err)
		}
	}()
	return nil
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
