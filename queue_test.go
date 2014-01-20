package queue

import (
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
		err := q.Run(true)
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
		err := ti.Run(true)
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
	err := New().Add(setToX).Add(5).Run(true)
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

	exp := "InvalidFunc int at position 1: int is no func"
	if err.Error() != exp {
		t.Errorf("expecting error message: %#v, got: %#v", exp, err.Error())
	}
}

func TestWrongParams(t *testing.T) {
	err := New().Add(set, 4).Add(set, "hi").Run(true)
	if err == nil {
		t.Errorf("expecting error, but got none")
	}
	details, ok := err.(CallError)

	if !ok {
		t.Errorf("error is no CallError, but: %T", err)
		return
	}

	if details.Position != 0 {
		t.Errorf("expecting error at position 0, but got %d", details.Position)
	}

	exp := `CallError func(string) error at position 0 called with []interface {}{4}: reflect: Call using int as type string`
	if err.Error() != exp {
		t.Errorf("expecting error message: %#v, got: %#v", exp, err.Error())
	}
}

func TestPanic(t *testing.T) {
	err := New().Add(doPanic).Run(false)
	if err == nil {
		t.Errorf("expecting error, but got none")
	}
	details, ok := err.(CallError)

	if !ok {
		t.Errorf("error is no CallError, but: %T", err)
		return
	}

	if details.Position != 0 {
		t.Errorf("expecting error at position 0, but got %d", details.Position)
	}
}

func TestMethod(t *testing.T) {
	s := &S{4}
	err := New().Add(s.Add, 4).Add(s.Add, 7).Run(true)

	if s.Get() != 15 {
		t.Errorf("wrong result: expected 15, got %d", s.Get())
	}

	if err != nil {
		t.Errorf("expecting no error, but got: %s", err.Error())
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
		err := ti.Run(true)
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
		err := ti.Run(true)
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
		Add(s.Set, PIPE).Run(true)

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
		OnError(IGNORE).Run(true)

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
		Run(true)

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
