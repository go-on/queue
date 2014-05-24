package queue

import "reflect"

type call struct {
	function  reflect.Value
	arguments []interface{}
	name      string
}

type callrun []Queuer
type callfallback []Queuer

func Call(function interface{}, arguments ...interface{}) *call {
	return &call{
		function:  reflect.ValueOf(function),
		arguments: arguments,
	}
}

func CallNamed(name string, function interface{}, arguments ...interface{}) *call {
	return &call{
		function:  reflect.ValueOf(function),
		arguments: arguments,
		name:      name,
	}
}

// returns the given value and "Sets" the value of a pipe
func Value(i interface{}) interface{} { return i }

// ptr is a pointer to something that should be get
// the value to which the pointer points is put into the pipe
func Get(ptr interface{}) interface{} {
	return reflect.ValueOf(ptr).Elem().Interface()
}

func Set(ptrToSet interface{}, val interface{}) {
	reflect.ValueOf(ptrToSet).Elem().Set(reflect.ValueOf(val))
}

// a simple function at the end that does not return errors or anything
func Ok() {}

// collects different values and puts them in an interface
// interface is a slice of the type of the first given value
func Collect(val interface{}, vals ...interface{}) interface{} {

	sl := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(vals[0])), len(vals)+1, len(vals)+1)

	sl.Index(0).Set(reflect.ValueOf(val))

	for i := 0; i < len(vals); i++ {
		sl.Index(i + 1).Set(reflect.ValueOf(vals[i]))
	}

	return sl.Interface()
}

func Run(qs ...Queuer) callrun           { return callrun(qs) }
func Fallback(qs ...Queuer) callfallback { return callfallback(qs) }

// Add creates a call consisting of the given function and the given arguments and adds it
// to the call chain.
//
// The special argument PIPE is a placeholder for the return values of the previous call
// in the chain (minus returned errors).
//
// The number and type signature of the arguments and piped return values must
// match with the receiving function.
//
// More about valid queues: see Check()
// More about function calling: see Run() and Fallback()
func (q *Queue) Add(function interface{}, arguments ...interface{}) *Queue {
	q.calls = append(q.calls, &call{
		function:  reflect.ValueOf(function),
		arguments: arguments,
	})
	return q
}

// AddNamed behaves like Add, but names the call with the given name.
// This is useful for logging and debugging
func (q *Queue) AddNamed(name string, function interface{}, arguments ...interface{}) *Queue {
	q.calls = append(q.calls, &call{
		function:  reflect.ValueOf(function),
		arguments: arguments,
		name:      name,
	})
	return q
}

func Add(function interface{}, arguments ...interface{}) *Queue {
	return New().Add(function, arguments...)
}

func AddNamed(name string, function interface{}, arguments ...interface{}) *Queue {
	return New().AddNamed(name, function, arguments...)
}
