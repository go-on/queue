package queue

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

var testCases = []testcase{
	newT("a", newF(set, "a")),
	newT("ab", newF(set, "a"), newF(appendString, "b")),
	newT("ab5p", newF(set, "a"), newF(appendString, "b"), newF(appendIntAndString, 5, "p")),
	newT("b5p", newF(appendString, "b"), newF(appendIntAndString, 5, "p")),
	newT("a", newF(appendString, "b"), newF(appendIntAndString, 5, "p"), newF(set, "a")),
	newT("X", newF(setToX)),
	newT("Xb", newF(setToX), newF(appendString, "b")),
	newT("X", newF(appendString, "b"), newF(appendIntAndString, 5, "p"), newF(setToX)),
}

func TestNoErrors(t *testing.T) {
	for i, tc := range testCases {
		result = ""
		q := tc.Q()
		err := q.Run()
		if err != nil {
			t.Errorf("in testCases[%d]: should get no error, but got: %s", i, err)
		}
		if result != tc.result {
			t.Errorf("in testCases[%d]: expected %#v, but got: %#v", i, tc.result, result)
		}
	}
}

var testCasesErr = []testcaseErr{
	newTErr("a", "setErr", newF(setErr, "a")),
	newTErr("a", "setErr", newF(setErr, "a"), newF(appendString, "b")),
	newTErr("ab", "appendStringErr", newF(set, "a"), newF(appendStringErr, "b")),
	newTErr("ab", "appendStringErr", newF(set, "a"), newF(appendStringErr, "b"), newF(appendIntAndString, 5, "p")),
}

func TestErrors(t *testing.T) {
	for i, tc := range testCasesErr {
		result = ""
		ti := tc.Q()
		err := ti.Run()
		if err == nil {
			t.Errorf("in testCasesErr[%d] should get an error, but got none", i)
		}
		if err.Error() != tc.errMsg {
			t.Errorf("in testCasesErr[%d] wrong error message, expected %#v, but got %#v", i, tc.errMsg, err.Error())
		}
		if result != tc.result {
			t.Errorf("in testCasesErr[%d] wrong result expected %#v, but got: %#v", i, tc.result, result)
		}
	}
}

func TestNoFunc(t *testing.T) {
	err := New().Add(setToX).Add(5).CheckAndRun()
	if err == nil {
		t.Errorf("expecting error, but got none")
	}
	details, ok := err.(InvalidFunc)

	if !ok {
		t.Errorf("error is no InvalidFunc, but: %T", err)
		return
	}

	if details.Position != 1 {
		t.Errorf("expecting error at position 1, but got %d", details.Position)
	}

	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expecting 'invalid' in error message, got: %#v", err.Error())
	}
}

func valsToTypes(vals []interface{}) []reflect.Type {
	types := make([]reflect.Type, len(vals))
	for i, v := range vals {
		types[i] = reflect.TypeOf(v)
	}
	return types
}

type validationtestCase struct {
	function  interface{}
	args      []interface{}
	shouldErr bool
}

func TestValidateArgs(t *testing.T) {
	/*
		we want the following tests:

		non variadic functions:
		(0) matching number of args, matching types
		(1) matching number of args, not matching types
		(2) not matching number of args, matching types
		(3) not matching number of args, not matching types
		(4) no args

		variadic functions:
		(5) matching number of args, matching types
		(6) more args, matching types
		(7) missing optional arg, matching types

		(8) matching number of args, not matching types before variadic
		(9) matching number of args, not matching type on variadic

		(10) more args, not matching types before variadic
		(11) more args, not matching types in variadic
		(12) more args, not matching types after variadic
		(13) missing optional arg, not matching types

		(14) missing args, matching types
		(15) missing args, not matching types
	*/

	newT := func(shouldErr bool, fn interface{}, args ...interface{}) *validationtestCase {
		return &validationtestCase{fn, args, shouldErr}
	}

	var testCases = []*validationtestCase{

		newT(false, set, "hi"),      // 0
		newT(true, set, 4),          // 1
		newT(true, set, "hi", "ho"), // 2
		newT(true, set, 4, 5),       // 3
		newT(false, read),           // 4

		newT(false, addIntsToString, "a", 4),    // 5
		newT(false, addIntsToString, "a", 4, 5), // 6
		newT(false, addIntsToString, "a"),       // 7

		newT(true, addIntsToString, 4.5, 4),   // 8
		newT(true, addIntsToString, "a", "b"), // 9

		newT(true, addIntsToString, 4.5, 4, 5),   // 10
		newT(true, addIntsToString, "a", "b", 5), // 11
		newT(true, addIntsToString, "a", 5, "b"), // 12
		newT(true, addIntsToString, 5),           // 13

		newT(true, addStringsandIntToString, "a"), // 14
		newT(true, addStringsandIntToString, 2),   // 15

	}

	for i, tc := range testCases {
		err := validateArgs(
			reflect.TypeOf(tc.function),
			valsToTypes(tc.args))

		if err != nil && !tc.shouldErr {
			t.Errorf("error in testCase[%d]: should not err, but got: %s", i, err)
		}

		if err == nil && tc.shouldErr {
			t.Errorf("error in testCase[%d]: should err, but did not", i)
		}
	}

}

func TestValidateFn(t *testing.T) {
	type test struct {
		*Queue
		shouldErr bool
	}

	newT := func(q *Queue, shouldErr bool) *test {
		return &test{q, shouldErr}
	}

	s := &S{}

	// maps queue to if it should return an error
	tests := []*test{
		// wrong argument type
		newT(New().Add(read).Add(s.Set, PIPE), true),

		// too many arguments
		newT(New().Add(multiInts).Add(s.Set, PIPE), true),

		// too few arguments
		newT(New().Add(read).Add(addStringsandIntToString, PIPE), true),

		// variadic params ok
		newT(New().Add(multiInts).Add(addIntsToString, "s", PIPE), false),

		// variadic params some not ok
		newT(New().Add(multiInts).Add(addIntsToString, "s", PIPE, "hi"), true),
	}

	for i, tt := range tests {
		err := tt.Check()
		if err == nil && tt.shouldErr {
			t.Errorf("should raise error, but does not", i)
			continue
		}

		if err != nil && !tt.shouldErr {
			t.Errorf("should not raise error, but does: %s", i, err.Error())
			continue
		}

		if err != nil {
			_, ok := err.(InvalidArgument)
			if !ok {
				t.Errorf("should be InvalidArgument error, but is: %T", i, err)
			}
		}
	}
}

func TestWrongParams(t *testing.T) {
	err := New().Add(set, 4).Add(set, "hi").CheckAndRun()
	if err == nil {
		t.Errorf("expecting error, but got none")
	}

	details, ok := err.(InvalidArgument)

	if !ok {
		t.Errorf("error is no InvalidArgument, but: %T", err)
		return
	}

	if details.Position != 0 {
		t.Errorf("expecting error at position 0, but got %d", details.Position)
	}

	if !strings.Contains(details.Error(), "invalid") {
		t.Errorf("wrong error message: should contain 'invalid', but is: %#v", details.Error())
	}

}

func TestPanic(t *testing.T) {
	err := New().Add(doPanic).Run()
	if err == nil {
		t.Errorf("expecting error, but got none")
	}
	details, ok := err.(CallPanic)

	if !ok {
		t.Errorf("error is no CallPanic, but: %T", err)
		return
	}

	if details.Position != 0 {
		t.Errorf("expecting error at position 0, but got %d", details.Position)
	}

	if !strings.Contains(details.Error(), "panicked") {
		t.Errorf("wrong error message: should contain 'panicked', but is: %#v", details.Error())
	}

}

func TestMethod(t *testing.T) {
	s := &S{4}
	err := New().Add(s.Add, 4).Add(s.Add, 7).Run()

	if s.Get() != 15 {
		t.Errorf("wrong result: expected 15, got %d", s.Get())
	}

	if err != nil {
		t.Errorf("expecting no error, but got: %s", err.Error())
	}
}

func TestInterface(t *testing.T) {
	v := ""
	a := func(s fmt.Stringer) {
		v = s.String()
	}
	err := New().Add(bytes.NewBufferString, "hi").Add(a, PIPE).Run()

	if err != nil {
		t.Errorf("expecting no error, but got: %s", err.Error())
	}

	if v != "hi" {
		t.Errorf("wrong result: expected \"hi\", got %#v", v)
	}

}

var testsPipe = []testcase{
	newT("45B745B",
		newF(strconv.Atoi, "4567456"),
		newF(setInt, PIPE),
		newF(read),
		newF(strings.Replace, PIPE, "6", "B", -1),
		newF(set, PIPE),
	),
	newT("45B745B",
		newF(set, "4567456"),
		newF(read),
		newF(strconv.Atoi, PIPE),
		newF(setInt, PIPE),
		newF(read),
		newF(strings.Replace, PIPE, "6", "B", -1),
		newF(set, PIPE),
	),
}

func TestPipeNoErrors(t *testing.T) {
	for i, tc := range testsPipe {
		result = ""
		ti := tc.Q()
		err := ti.Run()
		if err != nil {
			t.Errorf("in testsPipe[%d]: should get no error, but got: %s", i, err)
		}
		if result != tc.result {
			t.Errorf("in testsPipe[%d]: expected %#v, but got: %#v", i, tc.result, result)
		}
	}
}

var testsPipeErr = []testcaseErr{
	newTErr("456B456", `strconv.ParseInt: parsing "456B456": invalid syntax`,
		newF(set, "456B456"),
		newF(read),
		newF(strconv.Atoi, PIPE),
		newF(setInt, PIPE),
		newF(read),
		newF(strings.Replace, PIPE, "6", "B", -1),
		newF(set, PIPE),
	),
}

func TestPipeErrors(t *testing.T) {

	for i, tc := range testsPipeErr {
		result = ""
		ti := tc.Q()
		err := ti.Run()
		if err == nil {
			t.Errorf("in testsPipeErr[%d] should get an error, but got none", i)
		}
		if err.Error() != tc.errMsg {
			t.Errorf("in testsPipeErr[%d] wrong error message, expected %#v, but got %#v", i, tc.errMsg, err.Error())
		}
		if result != tc.result {
			t.Errorf("in testsPipeErr[%d] wrong result expected %#v, but got: %#v", i, tc.result, result)
		}
	}
}

func TestPipeMethod(t *testing.T) {
	s := &S{4}

	fn := func(i int) int {
		return i * 3
	}

	err := New().
		Add(s.Get).
		Add(fn, PIPE).
		Add(s.Set, PIPE).Run()

	if s.Get() != 12 {
		t.Errorf("wrong result: expected 12, got %d", s.Get())
	}

	if err != nil {
		t.Errorf("expecting no error, but got: %s", err.Error())
	}
}

func TestCatchHandle(t *testing.T) {
	s := &S{4}
	err := New().
		Add(s.Set, 30).
		Add(s.Add, 6).
		Add(s.Add, 10).
		OnError(IGNORE).Run()

	if err != nil {
		t.Errorf("expecting no returned error, but got %s", err.Error())
	}

	if s.Get() != 40 {
		t.Errorf("wrong value, expecting 40, but got %d", s.Get())
	}
}

func TestCatchHandleNot(t *testing.T) {
	s := &S{4}
	var catched error
	handleNot := ErrHandlerFunc(func(err error) error {
		catched = err
		return err
	})
	err := OnError(handleNot).
		Add(s.Set, 30).
		Add(s.Add, 6).
		Add(s.Add, 10).
		Run()

	if err == nil {
		t.Errorf("expecting returned error, but got none")
	}

	if catched == nil {
		t.Errorf("expecting catched error, but got none")
	}

	exp := "can't add 6"
	if err.Error() != exp {
		t.Errorf("wrong catched error messages, expected: %#v, got %#v", exp, err.Error())

	}
	if catched.Error() != exp {
		t.Errorf("wrong catched error messages, expected: %#v, got %#v", exp, catched.Error())

	}

	if s.Get() != 30 {
		t.Errorf("wrong value, expecting 30, but got %d", s.Get())
	}
}
