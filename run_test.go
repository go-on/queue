package queue

import (
	"bytes"
	"fmt"
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

func TestRuns(t *testing.T) {
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

var testCasesNamed = []testcase{
	newT("a", newFNamed("set1", set, "a")),
	newT("ab", newFNamed("set1", set, "a"), newFNamed("append1", appendString, "b")),
	newT("ab5p", newF(set, "a"), newF(appendString, "b"), newF(appendIntAndString, 5, "p")),
}

func TestRunsNamed(t *testing.T) {
	for i, tc := range testCasesNamed {
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
