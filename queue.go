package queue

import (
	"io"
	"reflect"
)

type Queue struct {
	startValues []reflect.Value

	// queue of functions
	functions []*call

	errHandler ErrHandler

	logTarget io.Writer

	logverbose bool

	// queues into which is feeded at position int
	feed map[int][]*Queue

	// queue of functions that are piped into at a certain point
	// however their return values will be discarded (apart from errors),
	// so they should take pointers to write something to them
	tees map[int][]*call

	// optional name of the queue (for logging and debugging)
	Name string
}

// New creates a new function queue
//
// Use Add() for adding functions to the Queue.
//
// Use OnError() to set a custom error handler.
//
// The default error handler is set by the runner function, Run() or Fallback().
//
// Use one of these runner functions to run the queue.
func New() *Queue {
	return &Queue{
		feed:        map[int][]*Queue{},
		startValues: []reflect.Value{},
		tees:        map[int][]*call{},
	}
}
