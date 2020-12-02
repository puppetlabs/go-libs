package concurrency_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/puppetlabs/go-libs/pkg/concurrency"
	"github.com/sirupsen/logrus"
)

// BenchmarkDispatcher creates a dispatcher, starts it and calls run
func BenchmarkDispatcher(b *testing.B) {
	logrus.SetOutput(ioutil.Discard)

	// Create Dispatcher and Start workers
	dispatcher := concurrency.NewDispatcher("test", 128, 1000)
	dispatcher.Start()

	// Create Fake Jobs
	log.Printf("Running With %d Jobs", b.N)
	for i := 0; i < b.N; i++ {
		dispatcher.Submit(&MockWork{id: strconv.Itoa(i), sleepTime: 20 * time.Millisecond})
	}
	log.Printf("Waiting to Stop")
	dispatcher.Stop()

	log.Printf("Completed %v jobs", dispatcher.ProcessedJobs())
}

// MockWork implements the Task interface, and can be submitted to the dispatcher to be processed by
// worker threads.
type MockWork struct {
	id        string
	sleepTime time.Duration
}

// Execute performs the actual work
func (w MockWork) Execute() error {
	mockMessage := MockMessage{Name: w.id}
	_, err := json.Marshal(mockMessage)
	time.Sleep(w.sleepTime)
	return err
}

// MockMessage ...
type MockMessage struct {
	Name string
}

// Create a new dispatcher with the name "foo".
// After creating a dispatcher, it must be started.
// You can submit jobs via any of the Submit methods.
// Once your done you should call Stop.  This will block and wait for all the worker go routines to complete.
func ExampleNewDispatcher() {
	dispatcher := concurrency.NewDispatcher("foo", 10, 5)
	dispatcher.Start()

	dispatcher.SubmitWork(func() error {
		// do some work
		return nil
	})

	dispatcher.Stop() // blocks and waits for all workers to complete
}

// Create a new dispatcher with the name "foo".
// Create an instance of MockWork and submit that to the dispatcher.
func ExampleNewDispatcher_second() {
	dispatcher := concurrency.NewDispatcher("foo", 10, 5)
	dispatcher.Start()

	w1 := MockWork{id: "bob", sleepTime: 5 * time.Second}
	dispatcher.Submit(w1)

	dispatcher.Stop() // blocks and waits for all workers to complete
}

// Start the dispatcher running.
// This starts a number of workers (go routines).  Start will return
// once all workers are started.
func ExampleDispatcher_Start() {
	dispatcher := concurrency.NewDispatcher("foo", 10, 5)
	dispatcher.Start()
}

// Stop the dispatcher.  Stop will wait for all workers to finish before returning.
func ExampleDispatcher_Stop() {
	dispatcher := concurrency.NewDispatcher("foo", 10, 5)
	dispatcher.Start()
	dispatcher.Stop()
}

// Submit work can be used to submit a func to the dispatcher instead of  creating a struct
// that implements the Task inteface.
func ExampleDispatcher_SubmitWork() {
	dispatcher := concurrency.NewDispatcher("foo", 10, 5)
	dispatcher.Start()

	var i int32 = 10

	dispatcher.SubmitWork(func() error {
		// Do some work here...
		atomic.AddInt32(&i, 1)
		return nil
	})
	dispatcher.SubmitWork(func() error {
		// Do some work here...
		atomic.AddInt32(&i, 1)
		return nil
	})
	dispatcher.Stop()

	fmt.Print(i)
	// Output: 12
}

// Submit three jobs to the dispatcher.  After calling stop verify the number of ProcessedJobs.
func ExampleDispatcher_ProcessedJobs() {
	dispatcher := concurrency.NewDispatcher("test", 1, 5)
	dispatcher.Start()

	work := func() error {
		// do some work
		return nil
	}

	dispatcher.SubmitWork(work)
	dispatcher.SubmitWork(work)
	dispatcher.SubmitWork(work)

	dispatcher.Stop()

	count := dispatcher.ProcessedJobs()

	fmt.Printf("Completed %d jobs", count)
	// Output: Completed 3 jobs
}
