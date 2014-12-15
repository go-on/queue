// Copyright (c) 2014 Marc René Arns. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
	Package q provides shortcuts for the package at http://github.com/go-on/queue

	It requires Go >= 1.1.

	It has a more compact syntax and is better includable with dot (.).

	Example 1

		err := New().Add(get, "Age", m).Add(strconv.Atoi, PIPE).Add(p.SetAge, PIPE).Run()

	would be rewritten to

		err := Q(get, "Age", m)(strconv.Atoi, V)(p.SetAge, V).Run()


	Example 2

		OnError(IGNORE).Add(get, "Age", m).Add(strconv.Atoi, PIPE).Add(p.SetAge, PIPE).CheckAndRun()

	would be rewritten to

		Err(IGNORE)(get, "Age", m)(strconv.Atoi, V)(p.SetAge, V).CheckAndRun()

*/
package q

import (
	"io"

	"gopkg.in/go-on/queue.v2"
)

var (
	V         = queue.PIPE
	STOP      = queue.STOP
	IGNORE    = queue.IGNORE
	PANIC     = queue.PANIC
	Call      = queue.Call
	CallNamed = queue.CallNamed
	Get       = queue.Get
	Set       = queue.Set
	Collect   = queue.Collect
	Value     = queue.Value
	Ok        = queue.Ok
	Fallback  = queue.Fallback
	Run       = queue.Run
)

type (
	add struct {
		fn   interface{}
		args []interface{}
		name string
	}

	setname struct {
		name string
	}

	tee struct {
		fn   interface{}
		args []interface{}
		name string
	}

	teeRun struct {
		qs       []queue.Queuer // []QFunc
		validate bool
	}

	sub struct {
		qs []queue.Queuer
	}

	teeFallback struct {
		qs       []queue.Queuer // []QFunc
		validate bool
	}

	run struct {
		validate bool
		err      error
	}

	/*
		fallback struct {
			validate bool
			err      error
			pos      int
		}
	*/

	log struct {
		writer  io.Writer
		verbose bool
	}

	onError struct {
		handler queue.ErrHandler
	}

	getQ struct {
		Queue *queue.Queue
	}

	// QFunc is a function that manages a queue and returns itself for chaining
	QFunc func(fn interface{}, params ...interface{}) QFunc
)

func (q QFunc) Queue() *queue.Queue {
	var qq = &getQ{}
	q(qq)
	return qq.Queue
}

func (q QFunc) Tee(fn interface{}, args ...interface{}) QFunc {
	var r = &tee{fn: fn, args: args}
	q(r)
	return q
}

func (q QFunc) SetName(name string) QFunc {
	q.Queue().SetName(name)
	return q
}

func (q QFunc) Name() string {
	return q.Queue().Name()
}

func (q QFunc) TeeNamed(name string, fn interface{}, args ...interface{}) QFunc {
	var r = &tee{fn: fn, args: args, name: name}
	q(r)
	return q
}

func (q QFunc) TeeAndRun(qs ...queue.Queuer) QFunc {
	var r = &teeRun{qs: qs}
	q(r)
	return q
}

func (q QFunc) Sub(qs ...queue.Queuer) QFunc {
	var r = &sub{qs: qs}
	q(r)
	return q
}

func (q QFunc) TeeAndCheckAndRun(qs ...queue.Queuer) QFunc {
	var r = &teeRun{qs: qs, validate: true}
	q(r)
	return q
}

func (q QFunc) TeeAndFallback(qs ...queue.Queuer) QFunc {
	var r = &teeFallback{qs: qs}
	q(r)
	return q
}

func (q QFunc) TeeAndCheckAndFallback(qs ...queue.Queuer) QFunc {
	var r = &teeFallback{qs: qs, validate: true}
	q(r)
	return q
}

func (q QFunc) Add(fn interface{}, args ...interface{}) QFunc {
	var r = &add{fn: fn, args: args}
	q(r)
	return q
}

func (q QFunc) AddNamed(name string, fn interface{}, args ...interface{}) QFunc {
	var r = &add{fn: fn, args: args, name: name}
	q(r)
	return q
}

// Run runs the queue
func (q QFunc) Run() error {
	var r = &run{validate: false}
	q(r)
	return r.err
}

// CheckAndRun first checks if there are any type errors in the
// function signatures or arguments and returns them. Without such errors,
// it is running the queue, like Run()
func (q QFunc) CheckAndRun() error {
	var r = &run{validate: true}
	q(r)
	return r.err
}

func (q QFunc) LogDebugTo(w io.Writer) QFunc {
	var r = &log{writer: w, verbose: true}
	q(r)
	return q

}
func (q QFunc) LogErrorsTo(w io.Writer) QFunc {
	var r = &log{writer: w}
	q(r)
	return q
}

// Err sets the ErrHandler of the queue
func (q QFunc) Err(handler queue.ErrHandler) QFunc {
	h := &onError{handler: handler}
	q(h)
	return q
}

func mkQFunc(q *queue.Queue) QFunc {
	var p QFunc
	p = func(fn interface{}, i ...interface{}) QFunc {
		switch v := fn.(type) {
		case *run:
			if v.validate {
				v.err = q.CheckAndRun()
			} else {
				v.err = q.Run()
			}
		case *getQ:
			v.Queue = q
		case *onError:
			q.OnError(v.handler)
		case *log:
			if v.verbose {
				q.LogDebugTo(v.writer)
			} else {
				q.LogErrorsTo(v.writer)
			}
		case *teeRun:
			if v.validate {
				q.TeeAndCheckAndRun(v.qs...)
			} else {
				q.TeeAndRun(v.qs...)
			}
		case *sub:
			q.Sub(v.qs...)
		case *teeFallback:
			if v.validate {
				q.TeeAndCheckAndFallback(v.qs...)
			} else {
				q.TeeAndFallback(v.qs...)
			}
		case *tee:
			q.TeeNamed(v.name, v.fn, v.args...)
		case *add:
			q.AddNamed(v.name, v.fn, v.args...)
		default:
			q.Add(fn, i...)
		}
		return p
	}
	return p
}

// Q returns a fresh queue as QFunc prefilled with the given function
// and arguments. The error handler is set to the default STOP (like in
// the queue package).
//
// The returned QFunc can be called to add new function/arguments combinations
// to the queue. Since it returns itself it could be chained.
func Q(function interface{}, arguments ...interface{}) QFunc {
	return mkQFunc(queue.New().Add(function, arguments...))
}

// Err returns a fresh queue as QFunc.
// It sets the given ErrHandler for the queue.
func Err(handler queue.ErrHandler) QFunc {
	return mkQFunc(queue.OnError(handler))
}
