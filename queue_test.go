package queue

import "testing"

func TestName(t *testing.T) {
	expected := "hiho"
	n := New().SetName(expected).Name()

	if n != expected {
		t.Errorf("wrong name, expecting: %#v, got: %#v", expected, n)
	}
}
