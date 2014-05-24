package queue

import "reflect"

// Tee allows piping of the same return value to different function calls.
//
// The return values from the given function are (apart from errors) discarded.
// If any of the arguments is the placeholder PIPE, it will be replaced by
// the return values of the previous regular call.
//
// If the tee call returns an error, it will be passed to the error handler of the queue
// and may stop the next regular call.
//
// If any check is performed on the queue, the tee calls are checked as well.
//
// To tee into other queues, use TeeAndRun or TeeAndFallback
func (q *Queue) Tee(function interface{}, arguments ...interface{}) *Queue {
	q.tees[len(q.calls)-1] = append(q.tees[len(q.calls)-1], &call{
		function:  reflect.ValueOf(function),
		arguments: arguments,
	})
	return q
}

func (q *Queue) Sub(feededQs ...Queuer) *Queue {
	return q.Add(feededQs)
}

// TeeNamed is like Tee, but allows a name to be assigned to the call for logging and error handling.
func (q *Queue) TeeNamed(name string, function interface{}, arguments ...interface{}) *Queue {
	q.tees[len(q.calls)-1] = append(q.tees[len(q.calls)-1], &call{
		function:  reflect.ValueOf(function),
		arguments: arguments,
		name:      name,
	})
	return q
}

// runTees runs the tees at position pos with the given vals
func (q *Queue) runTees(pos int, vals []reflect.Value) error {
	for i, tee := range q.tees[pos] {
		_, err := q.pipeFn(tee, pos*100+i, vals)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *Queue) defaultErrHandler() ErrHandler {
	if q.errHandler != nil {
		return q.errHandler
	}
	return STOP
}

// TeeAndRun allows piping of the same return value to different queues.
//
// The first call in each target queue should have the placeholder argument
// PIPE. When the main queue is run, each target queue is run via Run().
//
// If the Run() call of a target queue returns an error, it will be passed to the error handler
// of the main queue and immediately stop further processing (of other target queues and
// the next regular call).
//
// To be chainable, TeeAndRun returns the main queue.
func (q *Queue) TeeAndRun(feededQs ...Queuer) *Queue {
	fn := func(args ...interface{}) error {
		for _, feeded := range feededQs {
			err := feeded.Queue().run(toValues(args))
			if err != nil {
				return err
			}
		}
		return nil
	}
	q.Tee(fn, PIPE)
	return q
}

// TeeAndFallback works like TeeAndRun but runs the target queues via Fallback().
// The position returned by the particular Fallback() call on the target queue is discarded.
func (q *Queue) TeeAndFallback(feededQs ...Queuer) *Queue {

	fn := func(args ...interface{}) (err error) {
		errHandler := q.defaultErrHandler()
		for _, qe := range feededQs {
			err = qe.Queue().run(toValues(args))
			if err == nil {
				return
			}
		}

		if err != nil {
			err2 := errHandler.HandleError(err)
			q.logDebug("[E] %T(%#v) => %#v", errHandler, err, err2)
			err = err2
		}
		return
	}
	q.Tee(fn, PIPE)
	return q
}
