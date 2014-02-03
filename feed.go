package queue

import "reflect"

// to tee queues, run it like this
// Tee(RUN, New().Add(.....))
// and to fallback queues
// Tee(FALLBACK, New().Add(.....))
// and to tee normal functions:
// Tee(fn, args...)
// if the first argument is a queue, the last return values will be piped into the queue
// otherwise they will be piped via the PIPE placeholder as normal
func (q *Queue) Tee(function interface{}, arguments ...interface{}) *Queue {
	q.tees[len(q.functions)-1] = append(q.tees[len(q.functions)-1], &call{
		function:  reflect.ValueOf(function),
		arguments: arguments,
	})
	return q
}

func (q *Queue) runTeesAndFeed(pos int, vals []reflect.Value) error {
	for _, fe := range q.feed[pos] {
		fe.startValues = vals
	}
	for i, tee := range q.tees[pos] {
		var err error
		switch tee.function.Type() {
		case runTy:
			queue := tee.arguments[0].(*Queue)
			queue.startValues = vals
			// allow other functions with the type signature of RUN
			r := tee.function.Interface().(func(*Queue) error)
			err = r(queue)
			queue.startValues = []reflect.Value{}
		case fallbackTy:
			queue := tee.arguments[0].(*Queue)
			queue.startValues = vals
			// allow other functions with the type signature of FALLBACK
			fb := tee.function.Interface().(func(*Queue) (int, error))
			_, err = fb(queue)
			queue.startValues = []reflect.Value{}
		default:
			_, err = q.pipeFn(tee, pos*100+i, vals)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// Feed registers the given Queues to be feeded by the current function
// Feed maybe chained and therefore the main Queue is returned again
func (q *Queue) Feed(feededQs ...*Queue) *Queue {
	q.feed[len(q.functions)-1] = append(q.feed[len(q.functions)-1], feededQs...)
	return q
}
