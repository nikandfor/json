package jq_test

import (
	"fmt"

	"github.com/nikandfor/json/jq"
)

func ExampleSelector() {
	data := []byte(`{"key0":"skip it", "key1": {"next_key": ["array", null, {"obj":"val"}, "trailing element"]}}  "next"`)

	f := jq.Selector{"key1", "next_key", 2} // string keys and int array indexes are supported

	var res []byte // reusable buffer
	var i int      // start index

	// Most filters only parse single value and return index where the value ended.
	// Use jq.ApplyToAll(f, res[:0], data, 0) to process all values in a buffer.
	res, i, err := f.Apply(res[:0], data, i)
	if err != nil {
		// i is an index in a source buffer where the error occured.
	}

	fmt.Printf("value: %s", res)                           // res ends on newline
	fmt.Printf("final position: %d of %d\n", i, len(data)) // object was parsed to the end to be able to read next
	_ = data                                               // but not the next value

	// Output:
	// value: {"obj":"val"}
	// final position: 92 of 100
}
