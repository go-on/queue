package queue

import (
	"fmt"
	"strconv"
)

type person struct {
	Name string
	Age  int
}

func (p *person) SetAge(i int) { p.Age = i }
func (p *person) SetName(n string) error {
	if n == "Peter" {
		return fmt.Errorf("Peter is not allowed")
	}
	p.Name = n
	return nil
}

func getMap(k string, m map[string]string) string { return m[k] }

func setStruct(p *person, m map[string]string, handler ErrHandler) {
	// create a new queue with the default error handler
	q := New().
		// get the name from the map
		Add(getMap, "Name", m).
		// set the name in the struct
		Add(p.SetName, PIPE).
		// get the age from the map
		Add(getMap, "Age", m).
		// convert the age to int
		Add(strconv.Atoi, PIPE).
		// set the age in the struct
		Add(p.SetAge, PIPE).
		// inspect the struct
		Add(fmt.Printf, "SUCCESS %#v\n", p)

	// if a custom error handler is passed, use it,
	// otherwise the default error handler queue.STOP is used
	// which stops on the first error, returning it
	if handler != nil {
		q.OnError(handler)
	}
	// check the whole queue and run it
	err := q.CheckAndRun()

	// report, if there is an unhandled error
	if err != nil {
		fmt.Printf("ERROR %#v: %s\n", p, err)
	}
}

var ignoreAge = ErrHandlerFunc(func(err error) error {
	_, ok := err.(*strconv.NumError)
	if ok {
		return nil
	}
	return err
})

func Example() {
	var arthur = map[string]string{"Name": "Arthur", "Age": "42"}
	setStruct(&person{}, arthur, nil)

	var anne = map[string]string{"Name": "Anne", "Age": "4b"}
	// this will report the error of the invalid age that could not be parsed
	setStruct(&person{}, anne, nil)

	// this will ignore the invalid age, but no other errors
	setStruct(&person{}, anne, ignoreAge)

	var peter = map[string]string{"Name": "Peter", "Age": "4c"}

	// this will ignore the invalid age, but no other errors, so
	// it should err for the fact that peter is not allowed
	setStruct(&person{}, peter, ignoreAge)

	// this will ignore any errors and continue the queue run
	setStruct(&person{}, peter, IGNORE)

	// Output:
	// SUCCESS &queue.person{Name:"Arthur", Age:42}
	// ERROR &queue.person{Name:"Anne", Age:0}: strconv.ParseInt: parsing "4b": invalid syntax
	// SUCCESS &queue.person{Name:"Anne", Age:0}
	// ERROR &queue.person{Name:"", Age:0}: Peter is not allowed
	// SUCCESS &queue.person{Name:"", Age:0}
}
