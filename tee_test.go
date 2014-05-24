package queue

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
)

func TestTeeInvalid(t *testing.T) {
	err :=
		New().Add(
			strconv.Atoi, "9",
		).Tee(
			set, PIPE,
		).CheckAndRun()

	if err == nil {
		t.Errorf("expecting error, but got nil")
	}

	_, ok := err.(InvalidArgument)

	if !ok {
		t.Errorf("err is no InvalidArgument")
	}

}

func TestTeeAndRun(t *testing.T) {
	s := &S{}
	var bf bytes.Buffer
	err :=
		New().Add(
			set, "9",
		).Add(
			read,
		).TeeAndCheckAndRun(
			New().Add(
				strconv.Atoi, PIPE,
			).Add(
				fmt.Fprintf, &bf, "number is: %d", PIPE,
			),
		).TeeAndRun(
			New().Add(
				s.SetString, PIPE,
			),
		).CheckAndRun()

	if err != nil {
		t.Errorf("expecting no error, but got: %s", err)
	}

	expected := "number is: 9"
	if bf.String() != expected {
		t.Errorf("expecting buffer to be %#v, but is %#v", expected, bf.String())
	}

	if s.number != 9 {
		t.Errorf("expecting s.number to be 9, but is %d", s.number)
	}
}

func TestTeeAndFallback(t *testing.T) {
	result = ""
	s := &S{}
	err :=
		New().Add(
			set, "9.9",
		).Add(
			read,
		).TeeAndCheckAndFallback(
			Add(
				setErr, "a",
			),
			Add(
				setErr, "b",
			),
			Add(
				set, "c",
			),
			Add(Ok),
		).TeeAndFallback(
			New().Add(
				s.Set, 9,
			),
		).Run()

	if err != nil {
		t.Errorf("expecting no error, but got: %s", err)
	}

	expected := "c"
	if result != expected {
		t.Errorf("expecting buffer to be %#v, but is %#v", expected, result)
	}

	if s.number != 9 {
		t.Errorf("expecting s.number to be 9, but is %d", s.number)
	}
}

func TestTeeAndFallbackError(t *testing.T) {
	result = ""
	//_, err :=
	q := New().Add(
		set, "9.9",
	).Add(
		read,
	).TeeAndFallback(
		Add(
			setErr, "b",
		),
		Add(
			strconv.Atoi, "a",
		),
	) // .Fallback()

	err := q.Run()
	if err == nil {
		t.Errorf("expecting  error, but got nil")
	}

	expected := "b"
	if result != expected {
		t.Errorf("expecting buffer to be %#v, but is %#v", expected, result)
	}

}

func TestTeeAndFallbackError2(t *testing.T) {
	result = ""
	eh := ErrHandlerFunc(func(err error) error {
		// fmt.Printf("\nERR: %T\n", err)
		_, isNumError := err.(*strconv.NumError)
		if isNumError {
			return err
		}
		return nil
	})
	//_, err :=
	q1 := New().AddNamed(
		"setErr b",
		setErr, "b",
	)

	q2 := AddNamed(
		"strconv.Atoi a",
		strconv.Atoi, "a",
	)
	q := OnError(eh).AddNamed(
		"setErr x",
		setErr, "x",
	).TeeAndFallback(
		q1, q2,
	) //.LogDebugTo(os.Stdout) // .Fallback()

	q.SetName("main queue")

	//_, err := q.Fallback()
	err := q.Run()
	if err == nil {
		t.Errorf("expecting  error, but got nil")
	}

	expected := "b"
	if result != expected {
		t.Errorf("expecting buffer to be %#v, but is %#v", expected, result)
	}

}

func TestTeeAndRunError(t *testing.T) {
	s := &S{}
	var bf bytes.Buffer
	err :=
		New().Add(
			set, "9b",
		).Add(
			read,
		).TeeAndRun(
			New().Add(
				strconv.Atoi, PIPE,
			).Add(
				fmt.Fprintf, &bf, "number is: %d", PIPE,
			),
		).Tee(
			s.Set, 9,
		).Run()

	if err == nil {
		t.Errorf("expecting  error, but got nil")
	}

	expected := ""
	if bf.String() != expected {
		t.Errorf("expecting buffer to be %#v, but is %#v", expected, bf.String())
	}

	if s.number != 0 {
		t.Errorf("expecting s.number to be 0, but is %d", s.number)
	}
}

func TestTeeComplex(t *testing.T) {
	result = "9"
	s1 := &S{}
	s2 := &S{}
	var bf bytes.Buffer

	err := AddNamed(
		"read",
		read,
	).TeeNamed(
		"append 78",
		appendString, "78",
	).TeeAndRun(
		AddNamed(
			"atoi",
			strconv.Atoi, PIPE,
		).TeeAndFallback(
			AddNamed(
				"set s1",
				s1.Set, PIPE,
			).LogDebugTo(&bf).SetName("init s1"),
		).AddNamed("add s1",
			s1.Add, PIPE,
		).LogDebugTo(&bf).SetName("setting s1"),
	).TeeNamed(
		"set s2",
		s2.SetString, PIPE,
	).SetName("main").
		LogDebugTo(&bf). // log debugging to Stdout
		Run()            // checks types before running, faster is Run()

	if err != nil {
		t.Errorf("should not get an error, but got: %s", err)
	}

	if result != "978" {
		t.Errorf("result should be '978', but is: %#v", result)
	}

	if s1.number != 18 {
		t.Errorf("s1.number should be 18, but is: %d", s1.number)
	}

	if s2.number != 9 {
		t.Errorf("s2.number should be 9, but is: %d", s2.number)
	}

	expected := `
main - DEBUG: [0] "read" func() string{}() => "9"
main - DEBUG: [0] "append 78" func(...string) error{}("78") => <nil>
setting s1 - DEBUG: [0] "atoi" func(string) (int, error){}("9") => 9, <nil>
init s1 - DEBUG: [0] "set s1" func(int) error{}(9) => <nil>
setting s1 - DEBUG: [0] func(...interface {}) error{}(9) => <nil>
setting s1 - DEBUG: [1] "add s1" func(int) error{}(9) => <nil>
main - DEBUG: [1] func(...interface {}) error{}("9") => <nil>
main - DEBUG: [2] "set s2" func(string) error{}("9") => <nil>`

	if bf.String() != expected {
		t.Errorf("expected log: %#v, but got %#v", expected, bf.String())
	}
}
