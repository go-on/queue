package queue

import (
	"strconv"
	"strings"
	"testing"
)

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
	/*
		newTErr("456B456", `strconv.ParseInt: parsing "456B456": invalid syntax`,
			newF(set, "456B456"),
			newF(read),
			newF(strconv.Atoi, PIPE),
			newF(setInt, PIPE),
			newF(read),
			newF(strings.Replace, PIPE, "6", "B", -1),
			newF(set, PIPE),
		),
	*/
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

func TestNoFuncNamed(t *testing.T) {
	err := New().Add(setToX).AddNamed("five", 5).CheckAndRun()
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

	if details.Name != "five" {
		t.Errorf("expecting error details name to be 'five', but is %#v", details.Name)
	}

	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expecting 'invalid' in error message, got: %#v", err.Error())
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

func TestPanicNamed(t *testing.T) {
	err := New().AddNamed("doPanic", doPanic).Run()
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

	if details.Name != "doPanic" {
		t.Errorf("expecting call name in error to be 'doPanic', but is %#v", details.Name)
	}

	if !strings.Contains(details.Error(), "panicked") {
		t.Errorf("wrong error message: should contain 'panicked', but is: %#v", details.Error())
	}

}

func TestSubsError(t *testing.T) {
	s := "hu"
	result = ""

	q := Add(appendString, PIPE, "heho").Add(read)

	err := Add(
		Value, "hi",
	).Sub(
		Add(appendStringErr, "heho").Add(read),
		q,
	).Add(
		Set, &s, PIPE,
	).Run()

	if err == nil {
		t.Errorf("expecting error but got nil")
	}

}
