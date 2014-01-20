package queue

import (
	"fmt"
	//	"testing"
)

var result string

type (
	testfunc struct {
		fn     interface{}
		params []interface{}
	}

	testcase struct {
		funcs  []testfunc
		result string
	}

	testcaseErr struct {
		funcs  []testfunc
		result string
		errMsg string
	}
)

func (tf testfunc) add(q *Queue) *Queue {
	if len(tf.params) > 0 {
		return q.Add(tf.fn, tf.params...)
	}
	return q.Add(tf.fn)
}

func (tc testcase) Q() *Queue {
	q := New()
	for _, tf := range tc.funcs {
		q = tf.add(q)
	}
	return q
}

func (tc testcaseErr) Q() *Queue {
	q := New()
	for _, tf := range tc.funcs {
		q = tf.add(q)
	}
	return q
}

func newT(result string, fns ...testfunc) testcase {
	return testcase{funcs: fns, result: result}
}

func newTErr(result string, errMsg string, fns ...testfunc) testcaseErr {
	return testcaseErr{funcs: fns, result: result, errMsg: errMsg}
}

func newF(fn interface{}, params ...interface{}) testfunc {
	return testfunc{fn, params}
}

func set(s string) error {
	result = s
	return nil
}

func setInt(i int) error {
	result = fmt.Sprintf("%d", i)
	return nil
}

func read() string {
	return result
}

func setToX() {
	result = "X"
}

func appendString(s string) error {
	result = result + s
	return nil
}

func appendInts(is ...int) error {
	for _, i := range is {
		result = fmt.Sprintf("%s%d", result, i)
	}
	return nil
}

func appendIntAndString(i int, s string) error {
	result = fmt.Sprintf("%s%d%s", result, i, s)
	return nil
}

func doPanic() {
	panic("something")
}

func setErr(s string) error {
	result = s
	return fmt.Errorf("setErr")
}

func setToXErr() error {
	result = "X"
	return fmt.Errorf("setToXErr")
}

func appendStringErr(s string) error {
	result = result + s
	return fmt.Errorf("appendStringErr")
}

func appendIntsErr(is ...int) error {
	for _, i := range is {
		result = fmt.Sprintf("%s%d", result, i)
	}
	return fmt.Errorf("appendIntsErr")
}

func appendIntAndStringErr(i int, s string) error {
	result = fmt.Sprintf("%s%d%s", result, i, s)
	return fmt.Errorf("appendIntAndStringErr")
}

type S struct {
	number int
}

func (s *S) Set(i int) error {
	if i == 5 {
		return fmt.Errorf("can't set to 5")
	}
	s.number = i
	return nil
}

func (s *S) Get() int {
	return s.number
}

func (s *S) Add(i int) error {
	if i == 6 {
		return fmt.Errorf("can't add 6")
	}
	s.number = s.number + i
	return nil
}
