package queue

// calls the func at position i, with its arguments,

import (
	"fmt"
	"reflect"
)

// an internal type used to identify the pseudo parameter PIPE
type pipe struct{}

// PIPE is a pseudo parameter that will be replaced by the returned
// non error values of the previous function
var PIPE = pipe{}

func isNilable(obj interface {
	Kind() reflect.Kind
}) bool {
	switch obj.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return true
	default:
		return false
	}
}

// prepended by the given prepended args (that come from
// a result if a previous function)
// it returns all values returned by the function, if the
// last returned value is an error, it is stripped out and returned
// separately
// it catches any call panic
func (q *Queue) pipeFn(c *call, i int, piped []reflect.Value) (returns []reflect.Value, err error) {
	all := []interface{}{}

	for j, p := range c.arguments {
		switch a := p.(type) {
		case pipe:
			all = append(all, toInterfaces(piped)...)
		case *call:
			returns, err = q.pipeFn(a, i*100+j*10, piped)
			if err != nil {
				return
			}
			all = append(all, toInterfaces(returns)...)
		case callrun:
			errHandler := q.defaultErrHandler()
			// default error handler is STOP
			vals := piped
			for _, qe := range a {
				vals, err = qe.Queue().runAndReturn(vals)
				if err != nil {
					err2 := errHandler.HandleError(err)
					q.logDebug("[E] %T(%#v) => %#v", errHandler, err, err2)
					err = err2
				}
				if err != nil {
					return
				}
			}

			all = append(all, toInterfaces(vals)...)
		case callfallback:
			errHandler := q.defaultErrHandler()
			for _, qe := range a {
				returns, err = qe.Queue().runAndReturn(piped)
				if err == nil {
					break
				}
			}

			if err != nil {
				err2 := errHandler.HandleError(err)
				q.logDebug("[E] %T(%#v) => %#v", errHandler, err, err2)
				err = err2
			}
			if err != nil {
				return
			}

			all = append(all, toInterfaces(returns)...)
		default:
			all = append(all, p)
		}
	}

	defer func() {
		e := recover()
		if e != nil {
			ce := CallPanic{}
			ce.ErrorMessage = fmt.Sprintf("%v", e)
			ce.Params = all
			ce.Type = c.function.Type().String()
			ce.Position = i
			ce.Name = c.name
			err = ce
			if c.name == "" {
				q.logPanic("[%d] Panic in %v: %v", i, c.function.Type().String(), e)
			} else {
				q.logPanic("[%d] %#v Panic in %v: %v", i, c.name, c.function.Type().String(), e)
			}
			//q.logPanic(ce.Error())
		}
	}()

	vals := toValues(all)
	for ia := range vals {
		if isNilable(vals[ia]) && vals[ia].IsNil() {
			ty := c.function.Type().In(ia)
			vals[ia] = reflect.New(ty).Elem()
		}
	}
	returns = c.function.Call(vals)
	num := c.function.Type().NumOut()
	if num == 0 {
		return
	}

	if c.name == "" {
		q.logDebug("[%d] %v{}(%s) => %s",
			i,
			c.function.Type().String(),
			argReturnStr(all...),
			argReturnStr(toInterfaces(returns)...),
		)
	} else {
		q.logDebug("[%d] %#v %v{}(%s) => %s",
			i,
			c.name,
			c.function.Type().String(),
			argReturnStr(all...),
			argReturnStr(toInterfaces(returns)...),
		)
	}

	last := num - 1
	// TODO: there should be a better way to do this
	if c.function.Type().Out(last).String() == "error" {
		res := returns[last]
		returns = returns[:last]
		if !res.IsNil() {
			err = res.Interface().(error)
			if !q.logverbose {
				if c.name == "" {
					q.logError("[%d] %v => error: %#v",
						i, c.function.Type().String(), err,
					)
				} else {
					q.logError("[%d] %#v %v => error: %#v",
						i, c.name, c.function.Type().String(), err,
					)
				}
			}
		}
	}
	return
}
