package queue

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func TestCall(t *testing.T) {
	s := &S{}
	err := Add(set, "9").
		Add(read).
		Add(s.Set, Call(strconv.Atoi, PIPE)).
		Run()

	if err != nil {
		t.Errorf("expecting no error but got: %s", err)
	}

	if s.number != 9 {
		t.Errorf("number should be 9, but is: %d", s.number)
	}
}

func TestCallNamed(t *testing.T) {
	s := &S{}
	var bf bytes.Buffer
	err := Add(set, "9b").
		Add(read).
		Add(s.Set, CallNamed("Atoi", strconv.Atoi, PIPE)).
		LogErrorsTo(&bf).
		CheckAndRun()

	if err == nil {
		t.Errorf("expecting  error but got none %s", err)
	}

	_, ok := err.(*strconv.NumError)

	if !ok {
		t.Errorf("error is no *strconv.NumError, but %T", err)
	}

	errString := `ERROR: [200] "Atoi" func(string) (int, error)`
	if !strings.Contains(bf.String(), errString) {
		t.Errorf("error log should contain %#v, but is %#v", errString, bf.String())
	}

}

func TestCallNil(t *testing.T) {
	var s *S
	err := Add(set, "9").
		Add(read).
		Add(s.hi).
		Add(set, PIPE).
		Add((*S).hi, nil).
		Add(appendString, PIPE).
		Run()

	if err != nil {
		t.Errorf("expecting no error but got: %s", err)
	}

	if result != "hihohiho" {
		t.Errorf("result should be 'hihohiho', but is: %#v", result)
	}
}

func TestRun(t *testing.T) {
	var s = &S{}
	err := Add(
		set, "9",
	).Add(
		read,
	).Add(
		strconv.Atoi, Run(Add(appendString, PIPE).Add(read)),
	).Add(
		s.Set, PIPE,
	).
		Run()

	if err != nil {
		t.Errorf("expecting no error but got: %s", err)
	}

	if s.number != 99 {
		t.Errorf("s.number should be 99, but is: %#v", s.number)
	}
}

func fbtest(input string) error {
	return Add(
		set, input,
	).Add(
		read,
	).Add(
		appendString, Fallback(
			Add(strconv.Atoi, PIPE).Add(Value, " is int"),
			Add(strconv.ParseFloat, PIPE, 64).Add(set, "is float ").Add(read),
		), " - ", PIPE,
	).Run()
}

func TestSetGetCollect(t *testing.T) {
	s := "hu"
	result = ""

	err := Add(
		Value, "hi",
	).Add(
		strings.Join, Call(Collect, PIPE, Call(Get, &s)), "-",
	).Add(
		Set, &s, PIPE,
	).Run()

	if err != nil {
		t.Errorf("expecting no error but got: %s", err)
	}

	if s != "hi-hu" {
		t.Errorf("s should be 'hi-hu', but is: %#v", s)
	}
}

func TestSubs(t *testing.T) {
	s := "hu"
	result = ""

	q := Add(appendString, PIPE, "heho").Add(read)

	err := Add(
		Value, "hi",
	).Sub(
		q,
		q,
	).Add(
		Set, &s, PIPE,
	).Run()

	if err != nil {
		t.Errorf("expecting no error but got: %s", err)
	}

	expected := "hihehohihehoheho"

	if s != expected {
		t.Errorf("s should be %#v, but is: %#v", expected, s)
	}
}

func TestFallback(t *testing.T) {
	err := fbtest("9.7")

	if err != nil {
		t.Errorf("expecting no error but got: %s", err)
	}

	expected := `is float is float  - 9.7`

	if result != expected {
		t.Errorf("result should be %#v, but is: %#v", expected, result)
	}

	err = fbtest("7")

	if err != nil {
		t.Errorf("expecting no error but got: %s", err)
	}

	expected = `7 is int - 7`

	if result != expected {
		t.Errorf("result should be %#v, but is: %#v", expected, result)
	}

}
