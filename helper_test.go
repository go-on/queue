package queue

import (
	"fmt"
	"reflect"
	"strconv"
	//	"testing"
)

func valsToTypes(vals []interface{}) []reflect.Type {
	types := make([]reflect.Type, len(vals))
	for i, v := range vals {
		types[i] = reflect.TypeOf(v)
	}
	return types
}

var result string

type (
	testfunc struct {
		fn     interface{}
		params []interface{}
		name   string
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

	testcaseFallback struct {
		funcs    []testfunc
		result   string
		errMsg   string
		position int
	}
)

func (tf testfunc) add(q *Queue) *Queue {
	if len(tf.params) > 0 {
		return q.AddNamed(tf.name, tf.fn, tf.params...)
	}
	return q.AddNamed(tf.name, tf.fn)
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

func (tc testcaseFallback) Q() *Queue {
	q := New()
	for _, tf := range tc.funcs {
		q = tf.add(q)
	}
	return q
}

func newTFallback(result string, position int, errMsg string, fns ...testfunc) testcaseFallback {
	return testcaseFallback{
		funcs:    fns,
		result:   result,
		errMsg:   errMsg,
		position: position,
	}
}

func newT(result string, fns ...testfunc) testcase {
	return testcase{funcs: fns, result: result}
}

func newTErr(result string, errMsg string, fns ...testfunc) testcaseErr {
	return testcaseErr{funcs: fns, result: result, errMsg: errMsg}
}

func newF(fn interface{}, params ...interface{}) testfunc {
	return testfunc{fn, params, ""}
}

func newFNamed(name string, fn interface{}, params ...interface{}) testfunc {
	return testfunc{fn, params, name}
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

func appendString(ss ...string) error {
	for _, s := range ss {
		result = result + s
	}
	return nil
}

func addIntsToString(s string, ints ...int) string {
	return fmt.Sprintf("%s %v", s, ints)
}

func addStringsandIntToString(s string, i int, strs ...string) string {
	return fmt.Sprintf("%s %d %v", s, i, strs)
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

func multiInts() (int, int, int) {
	return 1, 2, 3
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

type numError int

func (n numError) Error() string {
	return fmt.Sprintf("can't set to %d", int(n))
}

func (s *S) Set(i int) error {
	if i == 5 {
		//return fmt.Errorf("can't set to 5")
		return numError(5)
	}
	s.number = i
	return nil
}

func (s *S) hi() string {
	return "hiho"
}

func (s *S) SetString(str string) error {
	i, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	return s.Set(i)
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
