package queue

import (
	"bytes"
	"strconv"
	"testing"
)

func TestLog(t *testing.T) {
	s := &S{}
	tests := []testcaseErr{
		newTErr(
			`
logtest - DEBUG: [0] func(string) error{}("a") => &errors.errorString{s:"setErr"}
logtest - DEBUG: [E] queue.ErrHandlerFunc(&errors.errorString{s:"setErr"}) => &errors.errorString{s:"setErr"}`,
			`
ERROR: [0] func(string) error => error: &errors.errorString{s:"setErr"}`,
			newF(setErr, "a"),
		),

		newTErr(
			`
logtest - DEBUG: [0] "a setter" func(string) error{}("a") => &errors.errorString{s:"setErr"}
logtest - DEBUG: [E] queue.ErrHandlerFunc(&errors.errorString{s:"setErr"}) => &errors.errorString{s:"setErr"}`,
			`
ERROR: [0] "a setter" func(string) error => error: &errors.errorString{s:"setErr"}`,
			newFNamed("a setter", setErr, "a"),
			newF(appendString, "b"),
		),

		newTErr(
			`
logtest - DEBUG: [0] func(string) error{}("a") => <nil>
logtest - DEBUG: [1] func(string) error{}("b") => &errors.errorString{s:"appendStringErr"}
logtest - DEBUG: [E] queue.ErrHandlerFunc(&errors.errorString{s:"appendStringErr"}) => &errors.errorString{s:"appendStringErr"}`,
			`
ERROR: [1] func(string) error => error: &errors.errorString{s:"appendStringErr"}`,
			newF(set, "a"), newF(appendStringErr, "b")),

		newTErr(
			`
logtest - DEBUG: [0] func(string) error{}("7") => <nil>
logtest - DEBUG: [1] func() string{}() => "7"
logtest - DEBUG: [2] func(string) (int, error){}("7") => 7, <nil>
logtest - DEBUG: [3] func(int) error{}(7) => <nil>
logtest - PANIC: [4] Panic in func(string) error: reflect: Call with too few input arguments
logtest - DEBUG: [E] queue.ErrHandlerFunc(queue.CallPanic{Position:4, Type:"func(string) error", Params:[]interface {}{}, ErrorMessage:"reflect: Call with too few input arguments", Name:""}) => queue.CallPanic{Position:4, Type:"func(string) error", Params:[]interface {}{}, ErrorMessage:"reflect: Call with too few input arguments", Name:""}`,
			`
PANIC: [4] Panic in func(string) error: reflect: Call with too few input arguments`,
			newF(set, "7"),
			newF(read),
			newF(strconv.Atoi, PIPE),
			newF(s.Set, PIPE),
			newF(set, PIPE),
		),
	}

	for i, tc := range tests {
		result = ""
		ti := tc.Q().SetName("logtest")
		var bf bytes.Buffer
		ti.LogDebugTo(&bf)
		ti.Run()

		if bf.String() != tc.result {
			t.Errorf("in testlog[%d] wrong debug log, expected\n\t%s\n\nbut got\n\t%s", i, tc.result, bf.String())
		}

		bf.Reset()
		ti.SetName("")
		ti.LogErrorsTo(&bf)
		ti.Run()

		if bf.String() != tc.errMsg {
			t.Errorf("in testlog[%d] wrong fatal log, expected\n\t%s\n\nbut got\n\t%s", i, tc.errMsg, bf.String())
		}

	}
}
