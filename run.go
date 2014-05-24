package queue

import "reflect"

// Run runs the function queue.
//
// In the run, every function in the queue is called with
// its arguments. If one of the arguments is PIPE, PIPE is replaced
// by the returned values of previous calls.
//
// If the last return value of a function is of type error, it value is
// skipped when piping to the next function and the error is checked.
//
// If the error is not nil, the ErrHandler of the Queue is called.
// If the ErrHandler returns nil, the next function is called.
// If it returns an error, the queue is stopped and the error is returned.
//
// The default ErrHandler is STOP, which will stop the run on the first error.
//
// If there are any errors with the given function types and arguments, the errors
// will no be very descriptive. In this cases use CheckAndRun() to see if there are any
// errors in the function or argument types or use LogDebugTo to get detailed debugging
// informations.
//
// Since no arguments are saved inside the queue, a queue might be run multiple times.
func (q *Queue) Run() (err error) {
	return q.run(nil)
}

// run with given start values and return the last return values
func (q *Queue) runAndReturn(vals []reflect.Value) (returns []reflect.Value, err error) {
	errHandler := q.errHandler
	// default error handler is STOP
	if errHandler == nil {
		errHandler = STOP
	}

	for i, fn := range q.calls {
		if fn.function.Type() == queuersType {
			for _, sub := range fn.function.Interface().([]Queuer) {
				vals, err = sub.Queue().runAndReturn(vals)
				if err != nil {
					err2 := errHandler.HandleError(err)
					q.logDebug("[E] %T(%#v) => %#v", errHandler, err, err2)
					if err2 != nil {
						err = err2
						return
					}
				}
			}
			returns = vals
			continue
		}

		vals, err = q.pipeFn(fn, i, vals)
		if err != nil {
			err2 := errHandler.HandleError(err)
			q.logDebug("[E] %T(%#v) => %#v", errHandler, err, err2)
			err = err2
		}
		if err != nil {
			return
		}

		err = q.runTees(i, vals)
		if err != nil {
			err2 := errHandler.HandleError(err)
			q.logDebug("[ET] %T(%#v) => %#v", errHandler, err, err2)
			err = err2
		}

		if err != nil {
			return
		}
	}
	returns = vals
	return
}

// run with given start values
func (q *Queue) run(vals []reflect.Value) (err error) {
	_, err = q.runAndReturn(vals)
	return
}
