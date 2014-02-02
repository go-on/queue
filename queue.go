// Copyright (c) 2014 Marc René Arns. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
	Package queue allows streamlined error handling and piping of returned values.

	This package is considered stable and ready for production.
	It requires Go >= 1.1.

	Motivation:

	In go, sometimes you need to run a bunch of functions that return errors and/or results. You might
	end up writing stuff like this

		err = fn1(...)

		if err != nil {
		   // handle error somehow
		}

		err = fn2(...)

		if err != nil {
		   // handle error somehow
		}

		...


	a lot of times.

	This is especially annoying if you want to handle all errors the same way
	(e.g. return the first error).

	This package provides a way to call functions in a queue while collecting the errors via a
	predefined or custom error handler. The predefined handler returns on the first error and
	custom error handlers might be used to catch/handle some/all kinds of errors while keeping the
	queue running.

	Usage:

		...
		// create a new queue
		err := New().
			// add function get to the queue that should be called with "Age" and m
			Add(get, "Age", m).

			// add function strconv.Atoi and pass the value returned from get via PIPE
			Add(strconv.Atoi, PIPE).

			// add method SetAge of p and pass the value returned from strconv.Atoi
			// note that the second return value error is not part of the pipe
			// it will however be sent to the error handler if it is not nil
			Add(p.SetAge, PIPE).
			...
			.OnError(STOP)  // optional custom error handler, STOP is default
			.Run()          // run it, returning unhandled errors.

			- OR -

			.CheckAndRun() // if you want to check for type errors of the functions/arguments before the run



		...

	The functions in the queue are checked for the type of the last return
	value. If it is an error, the value will be checked when running the queue
	and the error handler is invoked if the error is not nil.

	The error handler decides, if it can handle the error and the run continues
	(by returning nil) or if it can't and the run stops (by returning an/the error).

	Custom error handlers must fullfill the ErrHandler interface.

	When running the queue, the return values of the previous function with be injected into
	the argument list of the next function at the position of the pseudo argument PIPE.
	However, if the last return value is an error, it will be omitted.

	There is also a different running mode invoked by the method Fallback() that runs the queue
	until the first function returns no error.

	A package with shortcuts that has a more compact syntax and is better includable with dot (.)
	is provided at github.com/go-on/queue/q
*/
package queue

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
)

type (
	Queue struct {
		// queue of functions
		functions []reflect.Value

		// maps the position of a function in the queue to its arguments
		arguments map[int][]interface{}

		errHandler ErrHandler

		//logger Logger
		logTarget io.Writer

		logverbose bool
	}
)

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
		arguments: map[int][]interface{}{},
	}
}

// OnError returns a new empty *Queue, where the
// errHandler is set to the given handler
//
// More about adding functions to the Queue: see Add().
// More about error handling and running a Queue: see Run() and Fallback().
func OnError(handler ErrHandler) *Queue {
	return &Queue{
		arguments:  map[int][]interface{}{},
		errHandler: handler,
	}
}

// OnError sets the errHandler and may be chained.
//
// If OnError() is called multiple times, only the last
// call has any effect.
func (q *Queue) OnError(handler ErrHandler) *Queue {
	q.errHandler = handler
	return q
}

// LogErrorsTo logs errors and panics to the given io.Writer
//
// LogErrorsTo() is an alternative to LogDebugTo() and they should no be called both, because they are both
// changing the logging target and verbosity.
//
// If more than one logging setter is called, only the last one
// has any effect.
func (q *Queue) LogErrorsTo(logTarget io.Writer) *Queue {
	q.logTarget = logTarget
	q.logverbose = false
	return q
}

// LogDebugTo logs debugging information to the given io.Writer
//
// LogDebugTo() is an alternative to LogErrorsTo() and they should no be called both, because they are both
// changing the logging target and verbosity.
//
// If more than one logging setter is called, only the last one
// has any effect.
func (q *Queue) LogDebugTo(logTarget io.Writer) *Queue {
	q.logTarget = logTarget
	q.logverbose = true
	return q
}

func (q *Queue) logPanic(format string, a ...interface{}) {
	if q.logTarget != nil {
		fmt.Fprintf(q.logTarget, "\nPANIC: "+format, a...)
	}
}

func (q *Queue) logError(format string, a ...interface{}) {
	if q.logTarget != nil {
		fmt.Fprintf(q.logTarget, "\nERROR: "+format, a...)
	}
}

func (q *Queue) logDebug(format string, a ...interface{}) {
	if q.logTarget != nil && q.logverbose {
		fmt.Fprintf(q.logTarget, "\nDEBUG: "+format, a...)
	}
}

// Add adds the given function with optional arguments to the function queue
// and may be chained.
//
// The number and type signature of the arguments and piped return values must
// match with the receiving function.
//
// More about valid queues: see Check()
// More about function calling: see Run() and Fallback()
func (q *Queue) Add(function interface{}, arguments ...interface{}) *Queue {
	q.functions = append(q.functions, reflect.ValueOf(function))
	if len(arguments) > 0 {
		q.arguments[len(q.functions)-1] = arguments
	}
	return q
}

// Check checks if the function signatures and argument types match and returns any errors
func (q *Queue) Check() (err error) {
	var piped []reflect.Type
	for i, _ := range q.functions {
		piped, err = q.validateFn(i, piped)
		if err != nil {
			return
		}
	}
	return
}

func validateNums(fn reflect.Type, args []reflect.Type) (numIns int, numArgs int, diff int, err error) {
	numIns = fn.NumIn()
	numArgs = len(args)
	diff = numArgs - numIns
	// if number is equal, there is never an error in num
	if diff == 0 {
		return
	}
	// if number is not equal and function is not variadic,
	// it is an error for sure
	if !fn.IsVariadic() {
		err = fmt.Errorf("func wants %d arguments, but gets %d",
			numIns, numArgs)
		return
	}

	// we are here, if the number is not equal and
	// the function is variadic. There should not be to few
	if diff < -1 {
		err = fmt.Errorf("func wants at least %d arguments, but gets %d",
			numIns, numArgs)
		return
	}

	// in all other cases the number of arguments is ok
	return
}

func validateArgs(fn reflect.Type, args []reflect.Type) error {
	numIns, _, diff, err := validateNums(fn, args)

	// error in number of inputs, stop here
	if err != nil {
		return err
	}
	// no inputs: no check required
	if numIns == 0 {
		return nil
	}

	// check all ins of the function unless the
	// function is variadic, then skip the last in
	limit := numIns
	if fn.IsVariadic() {
		limit -= 1
	}
	for i := 0; i < limit; i++ {
		is := args[i]
		should := fn.In(i)
		if !is.AssignableTo(should) {
			return fmt.Errorf("%d. argument is a %#v but should be a %#v", i+1, is.String(), should.String())
		}
	}
	// if is not variadic, we're done
	if !fn.IsVariadic() {
		return nil
	}

	// now func must be variadic and we need to check all the args
	// that are defined by the variadic
	should := fn.In(numIns - 1).Elem()
	for i := 0; i < diff+1; i++ {
		j := i + numIns - 1
		is := args[j]
		if !is.AssignableTo(should) {
			return fmt.Errorf("%d. argument  is a %#v but should be a %#v", j+1, is.String(), should.String())
		}
	}

	return nil
}

// validateFn validates the function at position i in the queue
func (q *Queue) validateFn(i int, piped []reflect.Type) (returns []reflect.Type, err error) {
	fn := q.functions[i]
	if fn.Kind() != reflect.Func {
		invErr := InvalidFunc{}
		invErr.ErrorMessage = fmt.Sprintf("%#v is no func", fn.Type().String())
		invErr.Position = i
		invErr.Type = fn.Type().String()
		err = invErr
		q.logPanic("[%d] %#v is no func", i, fn.Type().String())
		return
	}

	all := []reflect.Type{}
	args, hasArgs := q.arguments[i]
	if hasArgs {
		for _, p := range args {
			if _, isPipe := p.(pipe); isPipe {
				all = append(all, piped...)
			} else {
				all = append(all, reflect.TypeOf(p))
			}
		}
	}
	ftype := fn.Type()

	err = validateArgs(ftype, all)
	if err != nil {
		invErr := InvalidArgument{}
		invErr.ErrorMessage = err.Error()
		invErr.Position = i
		invErr.Type = fn.Type().String()
		err = invErr
		q.logPanic("[%d] %v Invalid arguments: %s", i, fn.Type().String(), err)
		return
	}

	num := ftype.NumOut()
	if num == 0 {
		return
	}

	if ftype.Out(num-1).String() == "error" {
		num = num - 1
	}
	returns = make([]reflect.Type, num)

	for i := 0; i < num; i++ {
		returns[i] = ftype.Out(i)
	}
	return
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
	var vals = []reflect.Value{}
	for i := range q.functions {
		vals, err = q.pipeFn(i, vals)
		if err != nil {
			err2 := errHandler.HandleError(err)
			q.logDebug("[E] %T(%#v) => %#v", errHandler, err, err2)
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
	var vals = []reflect.Value{}
	errHandler := q.errHandler
	// default error handler is IGNORE
	if errHandler == nil {
		errHandler = IGNORE
	}

	var errHandled error
	for i := range q.functions {
		pos = i
		vals, err = q.pipeFn(i, vals)
		// if the function did not err, it could handle the input
		// and therefor we will return because of success
		if err == nil {
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
	}
	// return the last function and the returned error from the last function call
	return
}

// CheckAndRun first runs Check() to see, if there are any type errors in the
// function signatures or arguments and returns them. Without such errors,
// it then calls Run()
func (q *Queue) CheckAndRun() (err error) {
	err = q.Check()
	if err != nil {
		return err
	}
	return q.Run()
}

// CheckAndFallback first runs Check() to see, if there are any type errors in the
// function signatures or arguments and returns them. Without such errors,
// it then calls Fallback()
func (q *Queue) CheckAndFallback() (i int, err error) {
	err = q.Check()
	if err != nil {
		return
	}
	return q.Fallback()
}

// calls the func at position i, with its arguments,
// prepended by the given prepended args (that come from
// a result if a previous function)
// it returns all values returned by the function, if the
// last returned value is an error, it is stripped out and returned
// separately
// it catches any call panic
func (q *Queue) pipeFn(i int, piped []reflect.Value) (returns []reflect.Value, err error) {
	fn := q.functions[i]
	all := []interface{}{}
	args, hasArgs := q.arguments[i]
	if hasArgs {
		for _, p := range args {
			if _, isPipe := p.(pipe); isPipe {
				all = append(all, toInterfaces(piped)...)
			} else {
				all = append(all, p)
			}
		}
	}
	defer func() {
		e := recover()
		if e != nil {
			ce := CallPanic{}
			ce.ErrorMessage = fmt.Sprintf("%v", e)
			ce.Params = all
			ce.Type = fn.Type().String()
			ce.Position = i
			err = ce
			q.logPanic("[%d] Panic in %v: %v", i, fn.Type().String(), e)
		}
	}()

	returns = fn.Call(toValues(all))
	num := fn.Type().NumOut()
	if num == 0 {
		return
	}

	q.logDebug("[%d] %v{}(%s) => %s",
		i,
		fn.Type().String(),
		argReturnStr(all...),
		argReturnStr(toInterfaces(returns)...),
	)

	last := num - 1
	// TODO: there should be a better way to do this
	if fn.Type().Out(last).String() == "error" {
		res := returns[last]
		returns = returns[:last]
		if !res.IsNil() {
			err = res.Interface().(error)
			if !q.logverbose {
				q.logError("[%d] %v => error: %#v",
					i, fn.Type().String(), err,
				)
			}
		}
	}
	return
}

func argReturnStr(args ...interface{}) string {
	var bf bytes.Buffer

	for i, arg := range args {
		if i > 0 {
			fmt.Fprintf(&bf, ", ")
		}
		fmt.Fprintf(&bf, "%#v", arg)
	}
	return bf.String()
}

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

// an internal type used to identify the pseudo parameter PIPE
type pipe struct{}

// PIPE is a pseudo parameter that will be replaced by the returned
// non error values of the previous function
var PIPE = pipe{}

// Error returned if a function is not valid
type InvalidFunc struct {
	// position of the function in the queue
	Position int

	// type signature of the function
	Type string

	// error message
	ErrorMessage string
}

func (i InvalidFunc) Error() string {
	return fmt.Sprintf("[%d] function %#v is invalid:\n\t%s", i.Position+1, i.Type, i.ErrorMessage)
}

// Error returned if a function is not valid
type InvalidArgument struct {
	// position of the function in the queue
	Position int

	// type signature of the function
	Type string

	// error message
	ErrorMessage string
}

func (i InvalidArgument) Error() string {
	return fmt.Sprintf("[%d] function %#v gets invalid argument:\n\t%s", i.Position+1, i.Type, i.ErrorMessage)
}

// Error returned if a function call triggered a panic
type CallPanic struct {
	// position of the function in the queue
	Position int

	// type signature of the function
	Type string

	// arguments passed to the function
	Params []interface{}

	// error message
	ErrorMessage string
}

func (c CallPanic) Error() string {
	return fmt.Sprintf("[%d] function %#v panicked (was called with %#v):\n\t%s",
		c.Position+1, c.Type, c.Params, c.ErrorMessage)
}

// toValues is a helper function that creates and returns a slice of
// reflect.Value values based on a given slice of interface{} values
func toValues(in []interface{}) []reflect.Value {
	out := make([]reflect.Value, len(in))
	for i := range in {
		if in[i] != nil {
			out[i] = reflect.ValueOf(in[i])
		}
	}
	return out
}

// toValues is a helper function that creates and returns a slice of
// interface{} values based on a given slice of reflect.Value values
func toInterfaces(in []reflect.Value) []interface{} {
	out := make([]interface{}, len(in))
	for i, vl := range in {
		out[i] = vl.Interface()
	}
	return out
}
