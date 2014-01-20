package q

import (
	"github.com/go-on/queue"
)

// shortcut for PIPE
var V = queue.PIPE

type (
	stopOnError struct {
		err error
	}

	catchErrors struct {
		eh  ErrHandler
		err error
	}

	qfunc func(fn interface{}, params ...interface{}) qfunc
)

func (p qfunc) S() error {
	var r = &stopOnError{}
	p(r)
	return r.err
}

func (p qfunc) C(eh ErrHandler) (err error) {
	ce := &catchErrors{eh: eh}
	p(ce)
	return ce.err
}

func T(fn interface{}, params ...interface{}) qfunc {
	t := Tie(fn, params...)
	var p qfunc
	p = func(fn interface{}, i ...interface{}) qfunc {
		switch v := fn.(type) {
		case *stopOnError:
			v.err = t.StopOnError()
		case *catchErrors:
			v.err = t.CatchErrors(v.eh)
		default:
			t.And(fn, i...)
		}
		return p
	}
	return p
}
