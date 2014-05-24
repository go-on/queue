package queue

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestCallCheck(t *testing.T) {
	s := &S{}
	err := AddNamed("set", set, "9").
		AddNamed("read", read).
		AddNamed("Atoi1", strconv.Atoi, PIPE).
		AddNamed("s.Set", s.Set, CallNamed("Atoi2", strconv.Atoi, PIPE)).
		Check()

	if err == nil {
		t.Errorf("expecting error but got none")
	}

	ia, ok := err.(InvalidArgument)

	if !ok {
		t.Errorf("error is no InvalidArgument, but %T", err)
	}

	if ia.Name != "Atoi2" {
		t.Errorf("error should happen in Atoi2, but happened in %#v", ia.Name)
	}
}

func TestCheckNil(t *testing.T) {
	var s *S
	err := Add(set, "9").
		Add(read).
		Add(s.hi).
		Add(set, PIPE).
		Add((*S).hi, nil).
		Add(appendString, PIPE).
		Check()

	if err != nil {
		t.Errorf("expecting no error but got: %s", err)
	}

}

func TestFallbackArgumentCheck(t *testing.T) {
	s := "hu"
	result = ""

	err := Add(
		set, "hi",
	).Add(
		appendString,
		Fallback(
			Add(setErr, "he").Add(read),
			Add(set, "ho").Add(read),
		),
	).Add(
		read,
	).Add(
		Set, &s, PIPE,
	).CheckAndRun()

	if err != nil {
		t.Errorf("expecting no error but got: %s", err)
	}

	expected := "hoho"

	if s != expected {
		t.Errorf("s should be %#v, but is: %#v", expected, s)
	}
}

func TestFallbackArgumentCheckError(t *testing.T) {
	s := "hu"
	result = ""

	err := Add(
		set, "hi",
	).Add(
		appendString,
		Fallback(
			Add(setErr, "he").Add(read),
			Add(set, 4).Add(read),
		),
	).Add(
		read,
	).Add(
		Set, &s, PIPE,
	).CheckAndRun()

	expected := `[0] function "func(string) error" gets invalid argument:
	1. argument is a "int" but should be a "string"`

	if err == nil {
		t.Errorf("expecting error but got none")
	}

	if err.Error() != expected {
		t.Errorf("expecting error %#v but got: %#v", expected, err.Error())
	}

}

func TestRunArgumentCheckError(t *testing.T) {
	s := "hu"
	result = ""

	err := Add(
		set, "hi",
	).Add(
		appendString,
		Run(Add(set, 4).Add(read)),
	).Add(
		read,
	).Add(
		Set, &s, PIPE,
	).CheckAndRun()

	expected := `[0] function "func(string) error" gets invalid argument:
	1. argument is a "int" but should be a "string"`

	if err == nil {
		t.Errorf("expecting error but got none")
	}

	if err.Error() != expected {
		t.Errorf("expecting error %#v but got: %#v", expected, err.Error())
	}

}

func TestSubsCheck(t *testing.T) {
	s := "hu"
	result = ""

	err := Add(
		set, "hi",
	).Sub(
		Add(read).Add(appendString, PIPE, "he").Add(read),
		Add(appendString, PIPE, "ho").Add(read),
	).Add(
		Set, &s, PIPE,
	).CheckAndRun()

	if err != nil {
		t.Errorf("expecting no error but got: %s", err)
	}

	expected := "hihihehihiheho"

	if s != expected {
		t.Errorf("s should be %#v, but is: %#v", expected, s)
	}
}

func TestSubsCheckError(t *testing.T) {
	s := "hu"
	result = ""

	err := Add(
		set, "hi",
	).Sub(
		Add(set, 4).Add(read),
	).Add(
		Set, &s, PIPE,
	).CheckAndRun()

	if err == nil {
		t.Errorf(`expecting error but got none`)
	}

	expected := `[0] function "func(string) error" gets invalid argument:
	1. argument is a "int" but should be a "string"`
	if err.Error() != expected {
		t.Errorf("expecting error %#v but got: %#v", expected, err.Error())
	}
}

/*
func TestFallbackCheck(t *testing.T) {
	s := &S{}
	_, err := New().Add(s.Set, 5).Add(s.SetString, PIPE).CheckAndFallback()
	if err == nil {
		t.Errorf("should return error, but returns nil")
	}

	_, ok := err.(InvalidArgument)

	if !ok {
		t.Errorf("error should be of type InvalidArgument, but is %T", err)
	}
}
*/

func TestTeeAndFallbackCheckInvalid(t *testing.T) {
	result = ""
	err :=
		New().Add(
			strconv.Atoi, "9",
		).TeeAndCheckAndFallback(
			New().Add(
				set, PIPE,
			),
		).Run()

	if err == nil {
		t.Errorf("expecting error, but got none")
	}
}

func TestTeeAndFallbackCheckError(t *testing.T) {
	result = ""
	err :=
		New().Add(
			set, "4.5",
		).Add(
			read,
		).TeeAndCheckAndFallback(
			New().Add(
				strconv.Atoi, PIPE,
			),
		).Run()

	if err == nil {
		t.Errorf("expecting error, but got none")
	}
}

func TestTeeAndRunCheckInvalid(t *testing.T) {
	result = ""
	err :=
		New().Add(
			strconv.Atoi, "9",
		).TeeAndCheckAndRun(
			New().Add(
				set, PIPE,
			),
		).Run()

	if err == nil {
		t.Errorf("expecting error, but got none")
	}
}

func TestTeeAndRunCheckError(t *testing.T) {
	result = ""
	err :=
		New().Add(
			set, "4.5",
		).Add(
			read,
		).TeeAndCheckAndRun(
			New().Add(
				strconv.Atoi, PIPE,
			),
		).Run()

	if err == nil {
		t.Errorf("expecting error, but got none")
	}
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
		newT(New().Add(multiInts).AddNamed("s.Set", s.Set, PIPE), true),

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

func TestWrongParamsNamed(t *testing.T) {
	err := New().AddNamed("setwrong", set, 4).Add(set, "hi").CheckAndRun()
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

	if details.Name != "setwrong" {
		t.Errorf("expecting error details name to be 'setwrong', but is %#v", details.Name)
	}

	if !strings.Contains(details.Error(), "invalid") {
		t.Errorf("wrong error message: should contain 'invalid', but is: %#v", details.Error())
	}

}
