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
