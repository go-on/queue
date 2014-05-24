package queue

import (
	"strconv"
	"testing"
)

func TestPanicErrHandler(t *testing.T) {
	defer func() {
		e := recover()
		if e == nil {
			t.Errorf("should panic, but does not")
		}
	}()

	OnError(PANIC).Add(strconv.Atoi, "b").Run()
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
