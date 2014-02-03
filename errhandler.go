package queue

type (
	// Each Queue has an error handler that is called if
	// a function returns an error.
	//
	// The default error handler when calling Run() is STOP and when calling Fallback() is IGNORE.
	// The error handler PANIC might be chosen to panic on the first error (some kind of "Must" for every
	// function call).
	ErrHandler interface {
		// HandleError receives a non nil error and may handle it.
		// An error is considered not handled, if HandleError() returns the given error.
		// An error is considered handled, if HandleError() returns something other than
		// the given error.
		// An error is considered catched, if HandleError() returns nil.
		// If HandleError() catches an error, the queue run will continue.
		// Otherwise the queue will be stopped and the error is returned.
		// See Run() and Fallback() for more details about returning errors.
		HandleError(error) error
	}

	// shortcut to let a func be an error handler
	ErrHandlerFunc func(error) error
)

func (f ErrHandlerFunc) HandleError(err error) error { return f(err) }

var (
	// ErrHandler, stops on the first error
	STOP = ErrHandlerFunc(func(err error) error { return err })
	// ErrHandler, ignores all errors
	IGNORE = ErrHandlerFunc(func(err error) error { return nil })

	// ErrHandler, panics on the first error
	PANIC = ErrHandlerFunc(func(err error) error {
		panic(err.Error())
		return err
	})
)

// OnError returns a new empty *Queue, where the
// errHandler is set to the given handler
//
// More about adding functions to the Queue: see Add().
// More about error handling and running a Queue: see Run() and Fallback().
func OnError(handler ErrHandler) (q *Queue) {
	q = New().OnError(handler)
	return
}

// OnError sets the errHandler and may be chained.
//
// If OnError() is called multiple times, only the last
// call has any effect.
func (q *Queue) OnError(handler ErrHandler) *Queue {
	q.errHandler = handler
	return q
}
