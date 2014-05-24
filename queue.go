package queue

import (
	"io"
	"reflect"
)

type Queue struct {
	// queue of to be run calls
	calls []*call

	errHandler ErrHandler

	logTarget io.Writer

	logverbose bool

	// queue of calls that are piped into at a certain point
	// however their return values will be discarded (apart from errors),
	// so they should take pointers to write something to them
	tees map[int][]*call

	subs map[int][]Queuer

	// optional name of the queue (for logging and debugging)
	name string
}

// New creates a new function queue
//
// Use Add() for adding calls to the Queue.
//
// Use OnError() to set a custom error handler.
//
// The default error handler is set by the runner function, Run() or Fallback().
//
// Use one of these runner calls to run the queue.
func New() *Queue {
	return &Queue{
		tees: map[int][]*call{},
		subs: map[int][]Queuer{},
	}
}

func (q *Queue) SetName(name string) *Queue {
	q.name = name
	return q
}

func (q *Queue) Name() string {
	return q.name
}

func (q *Queue) Queue() *Queue { return q }

type Queuer interface {
	Queue() *Queue
}

var queuersType = reflect.TypeOf([]Queuer{})
