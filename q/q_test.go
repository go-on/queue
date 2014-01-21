package q

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
)

func TestQ(t *testing.T) {
	var bf bytes.Buffer
	err := Q(strconv.Atoi, "4")(fmt.Fprintf, &bf, "%d", V).Err(STOP).Run(true)

	if err != nil {
		t.Errorf("expected no error, but got: %#v", err.Error())
	}

	if bf.String() != "4" {
		t.Errorf("expected 4, but got: %#v", bf.String())
	}
}

func TestErr(t *testing.T) {
	var bf bytes.Buffer
	err := Err(IGNORE)(strconv.Atoi, "b")(fmt.Fprintf, &bf, "%d", 5).Run(true)
	if err != nil {
		t.Errorf("expected no error, but got: %#v", err.Error())
	}

	if bf.String() != "5" {
		t.Errorf("expected 5, but got: %#v", bf.String())
	}
}
