package q

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
)

func TestQ(t *testing.T) {
	var bf bytes.Buffer
	err := Q(strconv.Atoi, "4")(fmt.Fprintf, &bf, "%d", V).Err(STOP).Run()

	if err != nil {
		t.Errorf("expected no error, but got: %#v", err.Error())
	}

	if bf.String() != "4" {
		t.Errorf("expected 4, but got: %#v", bf.String())
	}
}

func TestErr(t *testing.T) {
	var bf bytes.Buffer
	err := Err(IGNORE)(strconv.Atoi, "b")(fmt.Fprintf, &bf, "%d", 5).CheckAndRun()
	if err != nil {
		t.Errorf("expected no error, but got: %#v", err.Error())
	}

	if bf.String() != "5" {
		t.Errorf("expected 5, but got: %#v", bf.String())
	}
}

func TestFallbackErrSkip(t *testing.T) {
	var bf bytes.Buffer
	i, err := Q(strconv.Atoi, "3.5")(strconv.ParseFloat, "3.5", 64).LogErrorsTo(&bf).Fallback()
	if err != nil {
		t.Errorf("expected no error, but got: %#v", err.Error())
	}

	if bf.String() == "" {
		t.Errorf("error log should not be empty, but is")
	}

	if i != 1 {
		t.Errorf("should stop after last function (pos 1), but stops at %d", i)
	}

	// fmt.Println(bf.String())
}

func TestFallbackNoErr(t *testing.T) {
	var bf bytes.Buffer
	i, err := Q(strconv.Atoi, "3")(strconv.ParseFloat, "3", 64).LogDebugTo(&bf).CheckAndFallback()
	if err != nil {
		t.Errorf("expected no error, but got: %#v", err.Error())
	}

	if i != 0 {
		t.Errorf("should stop after first function (pos 0), but stops at %d", i)
	}

	expected := `
DEBUG: [0] func(string) (int, error){}("3") => 3, <nil>`

	if bf.String() != expected {
		t.Errorf("debug log should be %#v, but is %#v", expected, bf.String())
	}
}
