/*
Package concurrency contains code that helps build multi-threaded applications
*/
package concurrency

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// Task is the interface a struct must implement, so that it can
// be submitted to the dispatcher.
type Task interface {
	Execute() error
}

// Dispatcher defines an interface.
// There is only one implementation of the dispatcher available, but by using an interface it forces
// the user to institate the dispatcher by using the NewDispatcher function
//
// The dispatcher can process work that is submitted either as a struct that implements the Task interface
// or as an anonymous function.
// If the work to be processed requires state to be maintained then creating a struct that implements the interface is
// the best approach.
// If the work is stateless then passing an anonymous function is more suitable.
type Dispatcher interface {
	Start()
	Stop()
	Submit(Task) error
	SubmitWork(fn func() error) error
	ProcessedJobs() uint64
}

var errTimedOutWaitingToSubmitTask = errors.New("timed out waiting to submit task")

// NewDispatcher create a new Dispatcher.  The ID is used to identify the dispatcher in log messages.
// workers is the number of go routines that this dispatcher will create to process work.
// queueSize is the size of the channel used to store tasks.
func NewDispatcher(id string, workers, queueSize int) Dispatcher {
	return &dispatcher{
		ID:      id,
		workers: workers,
		tasks:   make(chan Task, queueSize),
		wg:      sync.WaitGroup{},
	}
}

// Start the Dispatcher running.  No work will be processed until Start() is called.
// This will start a a number of go routines that will consume work from the tasks queue.   Start will
// return once all go routines have been started.
func (d *dispatcher) Start() {
	logrus.Infof("Creating %d workers for %s dispatcher", d.workers, d.ID)
	for w := 1; w <= d.workers; w++ {
		foo := w
		d.wg.Add(1)
		go d.worker(foo)
	}
}

// Stop the Dispatcher running.  This closes the task queue, which will prevent any further work from being
// submitted.  It waits until all go routines complete before returning.
func (d *dispatcher) Stop() {
	close(d.tasks)
	d.wg.Wait() // wait for all workers to return
}

// Submit a task to the work queue.  It will return an error if there is a timeout while
// waiting for the work to be submitted to the queue.  Note that just because the work is
// successfully submitted to the queue, does not mean that it will successfully be processed.
func (d *dispatcher) Submit(task Task) error {
	return submit(task, d.tasks, nil)
}

// SubmitWork allows the caller to submit a function to the task queue.   This is useful if
// the caller doesn't need to maintain state inside the Task.  This function will return an error if there is
// a timeout while waiting for the work to be submitted to the queue.  Note that just because the work is
// successfully submitted to the queue, does not mean that it will successfully be processed.
func (d *dispatcher) SubmitWork(fn func() error) error {
	return d.Submit(work{
		fn: fn,
	})
}

// ProcessedJobs returns the number of jobs that have been executed by this instance
// of the dispatcher.
func (d *dispatcher) ProcessedJobs() uint64 {
	return d.ops
}

// worker reads from the task channel and when it receives
// a Task, calls its Execute function.
func (d *dispatcher) worker(id int) {
	for task := range d.tasks {
		logrus.Debugf("worker %d started  job", id)
		err := task.Execute()
		if err != nil {
			logrus.Error(err)
		}
		atomic.AddUint64(&d.ops, 1)
		logrus.Debugf("worker %d finished  job", id)
	}
	d.wg.Done()
}

// submit a task to the task queue.
func submit(task Task, queue chan Task, timeout <-chan time.Time) error {
	select {
	case queue <- task:
	case <-timeout:
		return errTimedOutWaitingToSubmitTask
	}
	return nil
}

// dispatcher keeps state for the dispatcher, including a unique name, the number of works to create
// and a channel of tasks to be processed.
type dispatcher struct {
	ID      string
	workers int
	tasks   chan Task
	wg      sync.WaitGroup
	ops     uint64
}

// work is a  private struct that implements the Task interface.  It is used to wrap functions prior to adding them to
// the dispatchers work queue.
type work struct {
	fn func() error
}

// Execute implements the Task interface.
func (w work) Execute() error {
	return w.fn()
}
