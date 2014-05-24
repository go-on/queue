package queue

import (
	"strconv"
	"strings"
	"testing"
)

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
