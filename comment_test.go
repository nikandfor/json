package json2

import (
	"errors"
	"io"
	"testing"
)

func TestDecodeComment(tb *testing.T) {
	data := []byte(`
// example object
{
	"a": "b", // first key
	/* "c": 4,
	/********
	d: e, */
	"f": [],
}
/*final*/
`)

	var d Iterator

	i, err := d.Skip(data, 0)
	if err != nil {
		tb.Errorf("error: %v", err)
	}

	_, i, err = d.Type(data, i)
	if !errors.Is(err, ErrShortBuffer) {
		tb.Errorf("wanted end-of-buffer: %v", err)
	}
	if i != len(data) {
		tb.Errorf("wrong index: %d / %d", i, len(data))
	}
}

func TestReaderComment(tb *testing.T) {
	data := []byte(`
// example object
{
	"a": "b", // first key
	/* "c": 4,
	/********
	d: e, */
	"f": [],
}
/*final*/
`)

	r := NewReader(data, nil)

	err := r.Skip()
	if err != nil {
		tb.Errorf("error: %v", err)
	}

	tp, err := r.Type()
	if !errors.Is(err, io.EOF) {
		tb.Errorf("wanted EOF: %v", err)
	}
	if tp != None {
		tb.Errorf("wanted none: %v", tp)
	}
}
