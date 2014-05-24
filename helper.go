package queue

// toValues is a helper function that creates and returns a slice of

import (
	"bytes"
	"fmt"
	"reflect"
)

func toValues(in []interface{}) []reflect.Value {
	out := make([]reflect.Value, len(in))
	for i := range in {
		if in[i] != nil {
			out[i] = reflect.ValueOf(in[i])
		} else {
			out[i] = reflect.ValueOf(&in[i]).Elem()
		}
	}
	return out
}

func toTypes(in []interface{}) []reflect.Type {
	out := make([]reflect.Type, len(in))
	for i := range in {
		//if in[i] != nil {
		out[i] = reflect.TypeOf(in[i])
		/*
			} else {
				out[i] = reflect.TypeOf(&in[i]).Elem()
				// out[i] = reflect.TypeOf(nil)
			}
		*/

	}
	return out
}

// toValues is a helper function that creates and returns a slice of
// interface{} values based on a given slice of reflect.Value values
func toInterfaces(in []reflect.Value) []interface{} {
	out := make([]interface{}, len(in))
	for i, vl := range in {
		//		if vl.IsNil() {
		//		out[i] = nil
		//} else {
		out[i] = vl.Interface()
		//}
	}
	return out
}

func argReturnStr(args ...interface{}) string {
	var bf bytes.Buffer

	for i, arg := range args {
		if i > 0 {
			fmt.Fprintf(&bf, ", ")
		}
		fmt.Fprintf(&bf, "%#v", arg)
	}
	return bf.String()
}
