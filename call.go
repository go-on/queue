package queue

import "reflect"

type call struct {
	function  reflect.Value
	arguments []interface{}
	name      string
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
	q.functions = append(q.functions, &call{
		function:  reflect.ValueOf(function),
		arguments: arguments,
	})
	return q
}

// WithName sets the name for the last added function call
func (q *Queue) WithName(name string) *Queue {
	l := len(q.functions)
	if l == 0 {
		panic("add function before naming it")
	}
	q.functions[l-1].name = name
	return q
}
