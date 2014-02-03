package queue

// Error returned if a function is not valid

import "fmt"

type InvalidFunc struct {
	// position of the function in the queue
	Position int

	// type signature of the function
	Type string

	// error message
	ErrorMessage string

	// name of the function call, if it is named
	Name string
}

func (i InvalidFunc) Error() string {
	if i.Name == "" {
		return fmt.Sprintf("[%d] function %#v is invalid:\n\t%s", i.Position, i.Type, i.ErrorMessage)
	}

	return fmt.Sprintf("[%d] %#v function %#v is invalid:\n\t%s", i.Position, i.Name, i.Type, i.ErrorMessage)
}

// Error returned if a function is not valid
type InvalidArgument struct {
	// position of the function in the queue
	Position int

	// type signature of the function
	Type string

	// error message
	ErrorMessage string

	// name of the function call, if it is named
	Name string
}

func (i InvalidArgument) Error() string {
	if i.Name == "" {
		return fmt.Sprintf("[%d] function %#v gets invalid argument:\n\t%s", i.Position, i.Type, i.ErrorMessage)
	}
	return fmt.Sprintf("[%d] %#v function %#v gets invalid argument:\n\t%s", i.Position, i.Name, i.Type, i.ErrorMessage)
}

// Error returned if a function call triggered a panic
type CallPanic struct {
	// position of the function in the queue
	Position int

	// type signature of the function
	Type string

	// arguments passed to the function
	Params []interface{}

	// error message
	ErrorMessage string

	// name of the function call, if it is named
	Name string
}

func (c CallPanic) Error() string {
	if c.Name == "" {
		return fmt.Sprintf("[%d] function %#v panicked (was called with %#v):\n\t%s",
			c.Position, c.Type, c.Params, c.ErrorMessage)
	}
	return fmt.Sprintf("[%d] %#v function %#v panicked (was called with %#v):\n\t%s",
		c.Position, c.Name, c.Type, c.Params, c.ErrorMessage)
}
