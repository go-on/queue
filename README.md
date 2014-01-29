queue
=====

Streamlined error handling and piping through a queue of go functions

[![Build Status](https://secure.travis-ci.org/go-on/queue.png)](http://travis-ci.org/go-on/queue)

[![GoDoc](https://godoc.org/github.com/go-on/queue?status.png)](http://godoc.org/github.com/go-on/queue)

Status
------

This package is considered stable and ready for production.

Why
---

In go, sometimes you need to run a bunch of functions that return errors and/or results. You might end up writing stuff like this

```
err = fn1(...)

if err != nil {
   // handle error somehow
}

err = fn2(...)

if err != nil {
   // handle error somehow
}

...

```

a lot of times. This is especially annoying if you want to handle all errors the same way (e.g. return the first error).

`queue` provides a way to call functions in a queue while collecting the errors via a predefined or custom error handler. The predefined handler returns on the first error and custom error handlers might be used to catch/handle some/all kinds of errors while keeping the queue running.

Examples
--------

```go
package main

import (
    "fmt"
    "github.com/go-on/queue"
    "strconv"
)

type Person struct {
    Name string
    Age  int
}

func (p *Person) SetAge(i int) { p.Age = i }
func (p *Person) SetName(n string) error {
    if n == "Peter" {
        return fmt.Errorf("Peter is not allowed")
    }
    p.Name = n
    return nil
}

func get(k string, m map[string]string) string { return m[k] }

func set(p *Person, m map[string]string, handler queue.ErrHandler) {
    // create a new queue with the default error handler
    q := queue.New().
        // get the name from the map
        Add(get, "Name", m).
        // set the name in the struct
        Add(p.SetName, queue.PIPE).
        // get the age from the map
        Add(get, "Age", m).
        // convert the age to int
        Add(strconv.Atoi, queue.PIPE).
        // set the age in the struct
        Add(p.SetAge, queue.PIPE).
        // inspect the struct
        Add(fmt.Printf, "SUCCESS %#v\n\n", p)

    // if a custom error handler is passed, use it,
    // otherwise the default error handler queue.STOP is used
    // which stops on the first error, returning it
    if handler != nil {
        q.OnError(handler)
    }
    // run the whole queue
    err := q.Run()

    // if you want a check for validity of the given functions and
    // parameters before the run, use 
    // err := q.CheckAndRun()

    // report, if there is an unhandled error
    if err != nil {
        fmt.Printf("ERROR %#v: %s\n\n", p, err)
    }
}

var ignoreAge = queue.ErrHandlerFunc(func(err error) error {
    _, ok := err.(*strconv.NumError)
    if ok {
        return nil
    }
    return err
})

func main() {
    var arthur = map[string]string{"Name": "Arthur", "Age": "42"}
    set(&Person{}, arthur, nil)

    var anne = map[string]string{"Name": "Anne", "Age": "4b"}
    // this will report the error of the invalid age that could not be parsed
    set(&Person{}, anne, nil)

    // this will ignore the invalid age, but no other errors
    set(&Person{}, anne, ignoreAge)

    var peter = map[string]string{"Name": "Peter", "Age": "4c"}

    // this will ignore the invalid age, but no other errors, so
    // it should err for the fact that peter is not allowed
    set(&Person{}, peter, ignoreAge)

    // this will ignore any errors and continue the queue run
    set(&Person{}, peter, queue.IGNORE)

}

```

Shortcuts
---------

A package with shortcuts that has a more compact syntax and is better includable with dot (.) is provided at github.com/go-on/queue/q

[![Bitdeli Badge](https://d2weczhvl823v0.cloudfront.net/go-on/queue/trend.png)](https://bitdeli.com/free "Bitdeli Badge")

