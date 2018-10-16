package json

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriter(t *testing.T) {
	w := NewWriter(make([]byte, 1000))

	w.ObjStart()
	w.ObjKeyString("key_a")
	w.StringString("value")
	w.ObjKeyString("key_b")
	w.ArrayStart()
	w.StringString("a")
	w.StringString("b")
	w.ObjStart()
	w.ObjEnd()
	w.StringString("c")
	w.ArrayEnd()
	w.ObjEnd()

	t.Logf("res: '%s'", w.Bytes())
}

func TestWriterIndent(t *testing.T) {
	w := NewIndentWriter(make([]byte, 1000), []byte(">"), []byte("--"))

	w.ObjStart()
	w.ObjKeyString("key_a")
	w.StringString("value")
	w.ObjKeyString("key_b")
	w.ArrayStart()
	w.StringString("a")
	w.StringString("b")
	w.ObjStart()
	w.ObjEnd()
	w.StringString("c")
	w.ArrayEnd()
	w.ObjEnd()
	if false {
	}

	t.Logf("res: \n%s", w.Bytes())
}

func TestWriterSetIndent(t *testing.T) {
	w := NewWriter(make([]byte, 1000))
	w.SetIndent([]byte(">"), []byte("--"))

	w.ObjStart()
	w.ObjKeyString("key_a")
	w.StringString("value")
	w.ObjKeyString("key_b")
	w.SetIndent(nil, nil)

	w.ArrayStart()
	w.StringString("a")
	w.StringString("b")
	w.ObjStart()
	w.ObjEnd()
	w.StringString("c")
	w.ArrayEnd()

	w.SetIndent([]byte(">"), []byte("--"))

	w.ObjEnd()

	t.Logf("res: \n%s", w.Bytes())
}

func TestCopy(t *testing.T) {
	w := NewWriter(nil)

	w.ObjStart()
	w.ObjKeyString("key_a")
	w.StringString("value")
	w.ObjKeyString("key_b")
	w.ArrayStart()
	w.StringString("a")
	w.StringString("b")
	w.ObjStart()
	w.ObjEnd()
	w.StringString("c")
	w.ArrayEnd()
	w.ObjEnd()

	t.Logf("res: '%s'", w.Bytes())

	ind := NewIndentWriter(nil, []byte("\t"), []byte("  "))
	r := Wrap(w.Bytes())

	Copy(ind, r)

	t.Logf("res:\n%s", ind.Bytes())

	w2 := NewWriter(nil)
	r2 := Wrap(ind.Bytes())

	Copy(w2, r2)

	t.Logf("res: '%s'", w2.Bytes())

	assert.Equal(t, w.Bytes(), w2.Bytes())
}

func TestType(t *testing.T) {
	data := []byte(`{"key_a":"value","key_b":["a","b",{},"c"]}`)
	t.Logf("data %d: '%s'", len(data), data)
	t.Logf("____   : '%s'", string("_123456789_123456789_123456789_123456789_123456789_123456789_")[:len(data)])
	r := Wrap(data)

	typeHelper(t, r)

	ind := NewIndentWriter(nil, []byte("\t"), []byte("  "))
	r = Wrap(ind.Bytes())

	Copy(ind, r)

	typeHelper(t, Wrap(ind.Bytes()))
}

func typeHelper(t *testing.T, r *Reader) {
	tp := r.Type()
	if tp == String {
		s := r.NextBytes()
		_ = s
		//	log.Printf("type: %v %s", tp, s)
	} else {
		//	log.Printf("type: %v", tp)
	}
	switch tp {
	case ObjStart:
		for r.HasNext() {
			tp := r.Type()
			k := r.NextString()
			log.Printf("kytp: %v %s", tp, k)
			typeHelper(t, r)
		}
	case ArrayStart:
		for r.HasNext() {
			typeHelper(t, r)
		}
	}
}
