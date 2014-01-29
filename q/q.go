// Copyright (c) 2014 Marc René Arns. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
	Package q provides shortcuts for the package at http://github.com/go-on/queue

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
	"github.com/go-on/queue"
)

var (
	V      = queue.PIPE
	STOP   = queue.STOP
	IGNORE = queue.IGNORE
)

type (
	run struct {
		validate bool
		err      error
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
