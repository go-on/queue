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

/*
func TestFallbackErrSkip(t *testing.T) {
	var bf bytes.Buffer
	i, err := Q(strconv.Atoi, "3.5").Add(strconv.ParseFloat, "3.5", 64).LogErrorsTo(&bf).Fallback()
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
*/

/*
func TestTee(t *testing.T) {
	var bf1 bytes.Buffer
	var bf2 bytes.Buffer
	var bf3 bytes.Buffer
	err := Q(
		strconv.Atoi, "4",
	).Tee(
		fmt.Fprintf, &bf1, "1-%d", V,
	).TeeNamed(
		"print bf2",
		fmt.Fprintf, &bf2, "2-%d", V,
	).TeeAndFallback(
		Q(
			fmt.Fprintf, &bf1, "-1-fallback",
		)(
			fmt.Fprintf, &bf1, "should not print",
		),
	).TeeAndRun(
		Q(
			fmt.Fprintf, &bf2, "-2a-%d", 1,
		)(
			fmt.Fprintf, &bf2, "-2b-%d", 2,
		),
	).TeeAndCheckAndRun(
		Q(
			fmt.Fprintf, &bf2, "-2x",
		)(
			fmt.Fprintf, &bf2, "-2y",
		),
	).TeeAndCheckAndFallback(
		Q(
			fmt.Fprintf, &bf3, "3-fallback",
		)(
			fmt.Fprintf, &bf3, "should not print",
		),
	).AddNamed(
		"print bf3",
		fmt.Fprintf, &bf3, "-3-%d", V,
	).Run()

	if err != nil {
		t.Errorf("expected no error, but got: %#v", err.Error())
	}

	if bf1.String() != "1-4-1-fallback" {
		t.Errorf("expected 1-4-1-fallback, but got: %#v", bf1.String())
	}

	if bf2.String() != "2-4-2a-1-2b-2-2x-2y" {
		t.Errorf("expected 2-4-2a-1-2b-2-2x-2y, but got: %#v", bf2.String())
	}

	if bf3.String() != "3-fallback-3-4" {
		t.Errorf("expected 3-fallback-3-4, but got: %#v", bf3.String())
	}
}
*/
