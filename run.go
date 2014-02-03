package queue

import (
	"fmt"
	"reflect"
)

var RUN func(*Queue) error
var runTy = reflect.TypeOf(func(*Queue) (e error) { return })
var FALLBACK func(*Queue) (int, error)
var fallbackTy = reflect.TypeOf(func(*Queue) (i int, e error) { return })

func init() {
	RUN = (*Queue).Run
	FALLBACK = (*Queue).Fallback
}

// Run runs the function queue.
//
// In the run, every function in the queue is called with
// its arguments. If one of the arguments is PIPE, PIPE is replaced
// by the returned values of previous functions.
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
// errors in the function or argument types.
//
// Since no arguments are saved inside the queue, a queue might be run multiple times.
func (q *Queue) Run() (err error) {
	errHandler := q.errHandler
	// default error handler is STOP
	if errHandler == nil {
		errHandler = STOP
	}

	var vals = q.startValues
	for i, fn := range q.functions {
		vals, err = q.pipeFn(fn, i, vals)
		if err != nil {
			err2 := errHandler.HandleError(err)
			q.logDebug("[E] %T(%#v) => %#v", errHandler, err, err2)
			err = err2
		}
		if err != nil {
			return
		}

		err = q.runTeesAndFeed(i, vals)
		if err != nil {
			err2 := errHandler.HandleError(err)
			q.logDebug("[ET] %T(%#v) => %#v", errHandler, err, err2)
			err = err2
		}

		if err != nil {
			return
		}
	}
	return
}

// Fallback runs the function queue similarly to Run() but has the following differences:
//
// - The first function call that returns no error finishes the run, returning the successful function and the error nil.
//
// - If an error happens the further processing of the queue depends on what the error handler returns. If the error handler returns nil, the queue continues to run. If the error handler returns an error the queue stops and the error is returned.
//
// - The default error handler is IGNORE, so that the run continues if an error happens.
//
// - If the last function in the queue ran and returned an error, this error will be returned, even if the ErrHandler catches it. The reason is that the last function is your final fallback, so you will want to get its error.
//
// Use cases:
//
// use
//
//     pos, err := New().
//       Add(fn1, ...).
//       Add(fn2, ...).
//       Add(fn3, ...).
//       Fallback()
//
// to have fn1, fn2, fn3 each called one after another, if the previous function call failed.
// If none of these function did run successful, err is the error of fn3 and pos is 2.
//
// If you have some kind of errors that should not be ignored, but stop the queue, you will need
// a custom error handler that returns a non nil error for these kinds of errors. Then you would use
//
//      pos, err := OnError(myErrHandler).
//       Add(fn1, ...).
//       Add(fn2, ...).
//       Add(fn3, ...).
//       Fallback()
//
// Now any error is passed to myErrHandler.HandleError(). If this method returns an error, the queue
// run is stopped and the error is returned.
//
// Since no arguments are saved inside the queue, a queue might be run in Fallback mode multiple times.
func (q *Queue) Fallback() (pos int, err error) {
	var vals = q.startValues
	errHandler := q.errHandler
	// default error handler is IGNORE
	if errHandler == nil {
		errHandler = IGNORE
	}

	var errHandled error
	for i, fn := range q.functions {
		pos = i
		vals, err = q.pipeFn(fn, i, vals)
		// if the function did not err, it could handle the input
		// and therefor we will return because of success
		if err == nil {
			for _, fe := range q.feed[i] {
				fe.startValues = vals
			}
			// return the last function that did not return an error
			return
		}
		// some error happened.
		// now the error handler gets a chance to "fix" some errors
		// it does that the same way the error handlers work in other runner functions:
		// if the error handler returns an error that is not nil, it interrupts the queue
		// otherwise the queue is continued and the error is "catched"
		// the default error handler is IGNORE
		// we need a different variable errHandled to be able to return the
		// original error if the last function call, even it was catched
		errHandled = errHandler.HandleError(err)
		q.logDebug("[E] %T(%#v) => %#v", errHandler, err, errHandled)
		if errHandled != nil {
			// let the error handler transform the error into some other and return that
			err = errHandled
			return
		}

		errTee := q.runTeesAndFeed(i, vals)
		if errTee != nil {
			errHandled = errHandler.HandleError(errTee)
			q.logDebug("[ET] %T(%#v) => %#v", errHandler, errTee, errHandled)
			if errHandled != nil {
				// let the error handler transform the error into some other and return that
				err = errHandled
				return
			}
		}
	}
	// return the last function and the returned error from the last function call
	return
}

// calls the func at position i, with its arguments,
// prepended by the given prepended args (that come from
// a result if a previous function)
// it returns all values returned by the function, if the
// last returned value is an error, it is stripped out and returned
// separately
// it catches any call panic
func (q *Queue) pipeFn(c *call, i int, piped []reflect.Value) (returns []reflect.Value, err error) {
	all := []interface{}{}

	for _, p := range c.arguments {
		if _, isPipe := p.(pipe); isPipe {
			all = append(all, toInterfaces(piped)...)
		} else {
			all = append(all, p)
		}
	}
	defer func() {
		e := recover()
		if e != nil {
			ce := CallPanic{}
			ce.ErrorMessage = fmt.Sprintf("%v", e)
			ce.Params = all
			ce.Type = c.function.Type().String()
			ce.Position = i
			ce.Name = c.name
			err = ce
			if c.name == "" {
				q.logPanic("[%d] Panic in %v: %v", i, c.function.Type().String(), e)
			} else {
				q.logPanic("[%d] %#v Panic in %v: %v", i, c.name, c.function.Type().String(), e)
			}
			//q.logPanic(ce.Error())
		}
	}()

	returns = c.function.Call(toValues(all))
	num := c.function.Type().NumOut()
	if num == 0 {
		return
	}

	if c.name == "" {
		q.logDebug("[%d] %v{}(%s) => %s",
			i,
			c.function.Type().String(),
			argReturnStr(all...),
			argReturnStr(toInterfaces(returns)...),
		)
	} else {
		q.logDebug("[%d] %#v %v{}(%s) => %s",
			i,
			c.name,
			c.function.Type().String(),
			argReturnStr(all...),
			argReturnStr(toInterfaces(returns)...),
		)
	}

	last := num - 1
	// TODO: there should be a better way to do this
	if c.function.Type().Out(last).String() == "error" {
		res := returns[last]
		returns = returns[:last]
		if !res.IsNil() {
			err = res.Interface().(error)
			if !q.logverbose {
				if c.name == "" {
					q.logError("[%d] %v => error: %#v",
						i, c.function.Type().String(), err,
					)
				} else {
					q.logError("[%d] %#v %v => error: %#v",
						i, c.name, c.function.Type().String(), err,
					)
				}
			}
		}
	}
	return
}

// an internal type used to identify the pseudo parameter PIPE
type pipe struct{}

// PIPE is a pseudo parameter that will be replaced by the returned
// non error values of the previous function
var PIPE = pipe{}
