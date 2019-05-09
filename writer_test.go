package json

import (
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

	assert.Equal(t, []byte(`{"key_a":"value","key_b":["a","b",{},"c"]}`), w.Bytes())

	//	t.Logf("res: '%s'", w.Bytes())
}

func TestWriterStreamStrings(t *testing.T) {
	w := NewWriter(make([]byte, 1000))

	w.StringString("a")
	w.NewLine()
	w.StringString("b")
	w.NewLine()
	w.StringString("c")

	assert.Equal(t, []byte("\"a\"\n\"b\"\n\"c\""), w.Bytes())

	//	t.Logf("res: '%s'", w.Bytes())
}

func TestWriterStreamStructs(t *testing.T) {
	w := NewWriter(make([]byte, 1000))

	w.Marshal(struct{ A string }{"str"})
	w.NewLine()
	w.Marshal(struct{ A string }{"str"})
	w.NewLine()
	w.Marshal(struct{ A string }{"str"})
	w.NewLine()

	assert.Equal(t, []byte(`{"A":"str"}
{"A":"str"}
{"A":"str"}
`), w.Bytes())

	//	t.Logf("res: '%s'", w.Bytes())
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

	assert.Equal(t, `>{
>--"key_a": "value",
>--"key_b": [
>----"a",
>----"b",
>----{},
>----"c"
>--]
>}`, string(w.Bytes()))

	//	t.Logf("res: \n%s", w.Bytes())
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

	assert.Equal(t, `>{
>--"key_a": "value",
>--"key_b": ["a","b",{},"c"]
>}`, string(w.Bytes()))

	//	t.Logf("res: \n%s", w.Bytes())
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

func TestWriteSafe(t *testing.T) {
	w := NewWriter(make([]byte, 1000))

	w.SafeStringString("\xbd\xb2\x3d\xbc\x20\xe2\x8c\x98")

	assert.Equal(t, []byte(`"\xbd\xb2=\xbc \u2318"`), w.Bytes())

	//	t.Logf("res: '%s'", w.Bytes())
}

func typeHelper(t *testing.T, r *Reader) {
	tp := r.Type()
	if tp == String {
		s := r.NextAsBytes()
		_ = s
		//	log.Printf("type: %v %s", tp, s)
	} else {
		//	log.Printf("type: %v", tp)
	}
	switch tp {
	case Object:
		for r.HasNext() {
			//	tp := r.Type()
			//	k := r.NextString()
			//	log.Printf("kytp: %v %s", tp, k)
			typeHelper(t, r)
		}
	case Array:
		for r.HasNext() {
			typeHelper(t, r)
		}
	}
}
