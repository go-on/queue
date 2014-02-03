package queue

import (
	"fmt"
	"reflect"
)

// Check checks if the function signatures and argument types match and returns any errors
func (q *Queue) Check() (err error) {
	var piped []reflect.Type
	for i, fn := range q.functions {
		piped, err = q.validateFn(fn, i, piped)
		if err != nil {
			return
		}
	}
	return
}

func validateNums(fn reflect.Type, args []reflect.Type) (numIns int, numArgs int, diff int, err error) {
	numIns = fn.NumIn()
	numArgs = len(args)
	diff = numArgs - numIns
	// if number is equal, there is never an error in num
	if diff == 0 {
		return
	}
	// if number is not equal and function is not variadic,
	// it is an error for sure
	if !fn.IsVariadic() {
		err = fmt.Errorf("func wants %d arguments, but gets %d",
			numIns, numArgs)
		return
	}

	// we are here, if the number is not equal and
	// the function is variadic. There should not be to few
	if diff < -1 {
		err = fmt.Errorf("func wants at least %d arguments, but gets %d",
			numIns, numArgs)
		return
	}

	// in all other cases the number of arguments is ok
	return
}

func validateArgs(fn reflect.Type, args []reflect.Type) error {
	numIns, _, diff, err := validateNums(fn, args)

	// error in number of inputs, stop here
	if err != nil {
		return err
	}
	// no inputs: no check required
	if numIns == 0 {
		return nil
	}

	// check all ins of the function unless the
	// function is variadic, then skip the last in
	limit := numIns
	if fn.IsVariadic() {
		limit -= 1
	}
	for i := 0; i < limit; i++ {
		is := args[i]
		should := fn.In(i)
		if !is.AssignableTo(should) {
			return fmt.Errorf("%d. argument is a %#v but should be a %#v", i+1, is.String(), should.String())
		}
	}
	// if is not variadic, we're done
	if !fn.IsVariadic() {
		return nil
	}

	// now func must be variadic and we need to check all the args
	// that are defined by the variadic
	should := fn.In(numIns - 1).Elem()
	for i := 0; i < diff+1; i++ {
		j := i + numIns - 1
		is := args[j]
		if !is.AssignableTo(should) {
			return fmt.Errorf("%d. argument  is a %#v but should be a %#v", j+1, is.String(), should.String())
		}
	}

	return nil
}

// validateFn validates the function at position i in the queue
func (q *Queue) validateFn(c *call, i int, piped []reflect.Type) (returns []reflect.Type, err error) {
	// fn := q.functions[i]
	if c.function.Kind() != reflect.Func {
		invErr := InvalidFunc{}
		invErr.ErrorMessage = fmt.Sprintf("%#v is no func", c.function.Type().String())
		invErr.Position = i
		invErr.Name = c.name
		invErr.Type = c.function.Type().String()
		err = invErr
		if c.name == "" {
			q.logPanic("[%d] %#v is no func", i, c.function.Type().String())

		} else {
			q.logPanic("[%d] %#v %#v is no func", i, c.name, c.function.Type().String())
		}
		return
	}

	all := []reflect.Type{}
	/*
		args, hasArgs := q.arguments[i]
		if hasArgs {
			for _, p := range args {
				if _, isPipe := p.(pipe); isPipe {
					all = append(all, piped...)
				} else {
					all = append(all, reflect.TypeOf(p))
				}
			}
		}
	*/
	for _, p := range c.arguments {
		if _, isPipe := p.(pipe); isPipe {
			all = append(all, piped...)
		} else {
			all = append(all, reflect.TypeOf(p))
		}
	}
	ftype := c.function.Type()

	err = validateArgs(ftype, all)
	if err != nil {
		invErr := InvalidArgument{}
		invErr.ErrorMessage = err.Error()
		invErr.Position = i
		invErr.Type = c.function.Type().String()
		invErr.Name = c.name
		err = invErr
		if c.name == "" {
			q.logPanic("[%d] %v Invalid arguments: %s", i, c.function.Type().String(), err)
		} else {
			q.logPanic("[%d] %#v %v Invalid arguments: %s", i, c.name, c.function.Type().String(), err)
		}
		return
	}

	num := ftype.NumOut()
	if num == 0 {
		return
	}

	if ftype.Out(num-1).String() == "error" {
		num = num - 1
	}
	returns = make([]reflect.Type, num)

	for i := 0; i < num; i++ {
		returns[i] = ftype.Out(i)
	}
	return
}

// CheckAndRun first runs Check() to see, if there are any type errors in the
// function signatures or arguments and returns them. Without such errors,
// it then calls Run()
func (q *Queue) CheckAndRun() (err error) {
	err = q.Check()
	if err != nil {
		return err
	}
	return q.Run()
}

// CheckAndFallback first runs Check() to see, if there are any type errors in the
// function signatures or arguments and returns them. Without such errors,
// it then calls Fallback()
func (q *Queue) CheckAndFallback() (i int, err error) {
	err = q.Check()
	if err != nil {
		return
	}
	return q.Fallback()
}
