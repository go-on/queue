package queue

/*

The package queue offers a queue of functions with optional parameters that
allows streamlined error handling and piping of returned values.

The functions in the queue are checked for the type of the last return
value. If it is an error, the value will be checked when running the queue
and the error handler is invoked if the error is not nil.

The error handler decides, if it can handle the error and the run continues
or if it can't and the run stops returning the unhandled error.

In a run, the return values of the previous function may be piped to the
next function. However, if the last return value of the previous function
is an error, the error will be ignored in the pipe.

Examples:

		package main

		import (
			"fmt"
			"github.com/go-on/queue"
			"strconv"
		)

		type Person struct {
			Name string
			Age  int
		}

		func (p *Person) SetAge(i int) { p.Age = i }
		func (p *Person) SetName(n string) error {
			if n == "Peter" {
				return fmt.Errorf("Peter is not allowed")
			}
			p.Name = n
			return nil
		}

		func get(k string, m map[string]string) string { return m[k] }

		func set(p *Person, m map[string]string, handler queue.ErrHandler) {
			// create a new queue with the default error handler
			q := queue.New().
				// get the name from the map
				Add(get, "Name", m).
				// set the name in the struct
				Add(p.SetName, queue.PIPE).
				// get the age from the map
				Add(get, "Age", m).
				// convert the age to int
				Add(strconv.Atoi, queue.PIPE).
				// set the age in the struct
				Add(p.SetAge, queue.PIPE).
				// inspect the struct
				Add(fmt.Printf, "SUCCESS %#v\n\n", p)

			// if a custom error handler is passed, use it,
			// otherwise the default error handler queue.STOP is used
			// which stops on the first error, returning it
			if handler != nil {
				q.OnError(handler)
			}
			// run the whole queue and validate it before running
			err := q.Run(true)

			// report, if there is an unhandled error
			if err != nil {
				fmt.Printf("ERROR %#v: %s\n\n", p, err)
			}
		}

		var ignoreAge = queue.ErrHandlerFunc(func(err error) error {
			_, ok := err.(*strconv.NumError)
			if ok {
				return nil
			}
			return err
		})

		func main() {
			var arthur = map[string]string{"Name": "Arthur", "Age": "42"}
			set(&Person{}, arthur, nil)

			var anne = map[string]string{"Name": "Anne", "Age": "4b"}
			// this will report the error of the invalid age that could not be parsed
			set(&Person{}, anne, nil)

			// this will ignore the invalid age, but no other errors
			set(&Person{}, anne, ignoreAge)

			var peter = map[string]string{"Name": "Peter", "Age": "4c"}

			// this will ignore the invalid age, but no other errors, so
			// it should err for the fact that peter is not allowed
			set(&Person{}, peter, ignoreAge)

			// this will ignore any errors and continue the queue run
			set(&Person{}, peter, queue.IGNORE)

		}

*/

import (
	"fmt"
	"reflect"
)

type (
	Queue struct {
		// queue of functions
		functions []reflect.Value

		// maps the position of a function in the queue to its parameters
		parameters map[int][]interface{}

		ErrHandler
	}
)

// New creates a new function queue, that has the default ErrHandler STOP
//
// More about adding functions to the Queue: see Add().
// More about error handling and running a Queue: see Run().
func New() *Queue {
	return &Queue{
		parameters: map[int][]interface{}{},
		ErrHandler: STOP,
	}
}

// OnError returns a new empty *Queue, where the
// ErrHandler is set to the given handler
//
// More about adding functions to the Queue: see Add().
// More about error handling and running a Queue: see Run().
func OnError(handler ErrHandler) *Queue {
	return &Queue{
		parameters: map[int][]interface{}{},
		ErrHandler: handler,
	}
}

// OnError sets the ErrHandler and may be chained
func (q *Queue) OnError(handler ErrHandler) *Queue {
	q.ErrHandler = handler
	return q
}

// Add adds the given function with optional parameters to the function queue
// and may be chained.
//
// The number and type signature of the parameters and piped return values must
// match with the receiving function.
//
// More about valid queues: see Validate()
// More about function calling: see Run()
func (q *Queue) Add(function interface{}, parameters ...interface{}) *Queue {
	q.functions = append(q.functions, reflect.ValueOf(function))
	if len(parameters) > 0 {
		q.parameters[len(q.functions)-1] = parameters
	}
	return q
}

// TODO: check if all functions are functions
// and if all input and piped parameters have the right type
// return specific error types
func (q *Queue) Validate() (err error) {
	for i, fn := range q.functions {
		if fn.Kind() != reflect.Func {
			invErr := InvalidFunc{}
			invErr.ErrorMessage = fmt.Sprintf("%s is no func", fn.Type().String())
			invErr.Position = i
			invErr.Type = fn.Type().String()
			err = invErr
			return
		}
	}
	return
}

// Run runs the function queue.
//
// If validate is true, Validate() is called and if it returns an error
// the queue will not be run and the error will be returned
//
// When the queue is run, every function in the queue is called with
// its parameters. If one of the parameters is PIPE, PIPE is replaced
// by the returned values of previous functions.
//
// If the last return value of a function is of type error, it value is
// skipped when piping to the next function and the error is checked.

// If the error is not nil, the ErrHandler of the Queue is called.
// If the ErrHandler returns nil, the next function is called.
// If it returns an error, the queue is stopped and the error is returned.
//
// The default ErrHandler is STOP, which will stop the run on the first error.
func (q *Queue) Run(validate bool) (err error) {
	if validate {
		err = q.Validate()
		if err != nil {
			return
		}
	}
	var vals = []reflect.Value{}
	for i := range q.functions {
		vals, err = q.pipeFn(i, vals)
		if err != nil {
			err = q.HandleError(err)
		}
		if err != nil {
			return
		}
	}
	return
}

// calls the func at position i, with its parameters,
// prepended by the given prepended params (that come from
// a result if a previous function)
// it returns all values returned by the function, if the
// last returned value is an error, it is stripped out and returned
// separately
// it catches any call panic
func (q *Queue) pipeFn(i int, piped []reflect.Value) (returns []reflect.Value, err error) {
	fn := q.functions[i]
	all := []interface{}{}
	params, hasParams := q.parameters[i]
	if hasParams {
		for _, p := range params {
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
			ce := CallError{}
			ce.ErrorMessage = fmt.Sprintf("%v", e)
			ce.Params = all
			ce.Type = fn.Type().String()
			ce.Position = i
			err = ce
		}
	}()
	returns = fn.Call(toValues(all))
	num := fn.Type().NumOut()
	if num == 0 {
		return
	}

	// TODO: there should be a better way to do this
	if fn.Type().Out(num-1).String() == "error" {
		res := returns[num-1]
		returns = returns[:num-1]
		if !res.IsNil() {
			err = res.Interface().(error)
		}
	}
	return
}

type (
	ErrHandler interface {
		// receives and error and either
		// handles the error and returns nil
		// or can't handle the error and returns the error
		// or returns another error
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
)

// an internal type used to identify the pseudo parameter PIPE
type pipe struct{}

// PIPE is a pseudo parameter that will be replaced by the returned
// non error values of the previous function
var PIPE = pipe{}

type InvalidFunc struct {
	Position     int
	Type         string
	ErrorMessage string
}

func (i InvalidFunc) Error() string {
	return fmt.Sprintf("InvalidFunc %s at position %d: %s", i.Type, i.Position, i.ErrorMessage)
}

type CallError struct {
	Position     int
	Type         string
	ErrorMessage string
	Params       []interface{}
}

func (c CallError) Error() string {
	return fmt.Sprintf("CallError %s at position %d called with %#v: %s",
		c.Type, c.Position, c.Params, c.ErrorMessage)
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
