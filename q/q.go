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

	"github.com/go-on/queue"
)

var (
	V      = queue.PIPE
	STOP   = queue.STOP
	IGNORE = queue.IGNORE
	PANIC  = queue.PANIC
)

type (
	run struct {
		validate bool
		err      error
	}

	fallback struct {
		validate bool
		err      error
		pos      int
	}

	log struct {
		writer  io.Writer
		verbose bool
	}

	onError struct {
		handler queue.ErrHandler
	}

	// QFunc is a function that manages a queue and returns itself for chaining
	QFunc func(fn interface{}, params ...interface{}) QFunc
)

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

func (q QFunc) CheckAndFallback() (int, error) {
	var r = &fallback{validate: true}
	q(r)
	return r.pos, r.err
}

func (q QFunc) Fallback() (int, error) {
	var r = &fallback{}
	q(r)
	return r.pos, r.err
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
		case *onError:
			q.OnError(v.handler)
		case *log:
			if v.verbose {
				q.LogDebugTo(v.writer)
			} else {
				q.LogErrorsTo(v.writer)
			}
		case *fallback:
			if v.validate {
				v.pos, v.err = q.CheckAndFallback()
			} else {
				v.pos, v.err = q.Fallback()
			}
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
