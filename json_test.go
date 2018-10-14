package json

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareKey(t *testing.T) {
	v := WrapString(`"key_a":"res"`)
	ok := v.compareKey([]byte("key_a"))
	assert.True(t, ok)
	assert.Equal(t, 7, v.i)
	assert.NoError(t, v.err)

	v = WrapString(`"key_a":"res"`)
	ok = v.compareKey([]byte("key_b"))
	assert.False(t, ok)
	assert.Equal(t, 7, v.i)
	assert.NoError(t, v.err)

	v = WrapString(`"key_a":"res"`)
	ok = v.compareKey([]byte("key"))
	assert.False(t, ok)
	assert.Equal(t, 7, v.i)
	assert.NoError(t, v.err)

	v = WrapString(`"key_a":"res"`)
	ok = v.compareKey([]byte("key_long"))
	assert.False(t, ok)
	assert.Equal(t, 7, v.i)
	assert.NoError(t, v.err)
}

func TestSkipStrings(t *testing.T) {
	data := `"str_1""str_2" "str_3" "str_4""str_5"`
	v := WrapString(data)

	j := 0
	for v.Type() != None {
		t.Logf("iter %d: %2v/%2v '%s' %v", j, v.i, v.end, v.b, v.err)
		v.Skip()
		j++
	}

	assert.NoError(t, v.Err())
	assert.Equal(t, 5, j)

	t.Logf("iter _: %2v/%2v '%s' %v", v.i, v.end, v.b, v.err)
}

func TestSkipArrays(t *testing.T) {
	data := `[] [] [][][]`
	v := WrapString(data)

	j := 0
	for v.Type() != None {
		t.Logf("iter %d: %2v/%2v '%s' %v", j, v.i, v.end, v.b, v.err)
		v.Skip()
		j++
	}

	assert.NoError(t, v.Err())
	assert.Equal(t, 5, j)

	t.Logf("iter _: %2v/%2v '%s' %v", v.i, v.end, v.b, v.err)
}

func TestSkipObjects(t *testing.T) {
	data := `{}{} {} {}{}`
	v := WrapString(data)

	j := 0
	for v.Type() != None {
		t.Logf("iter %d: %2v/%2v '%s' %v", j, v.i, v.end, v.b, v.err)
		v.Skip()
		j++
	}

	assert.NoError(t, v.Err())
	assert.Equal(t, 5, j)

	t.Logf("iter _: %2v/%2v '%s' %v", v.i, v.end, v.b, v.err)
}

func TestSkipObjectsNested(t *testing.T) {
	data := `{"a":{"b":{"c":{},"d":{}}},"e":{}}{}`
	v := WrapString(data)

	j := 0
	for v.Type() != None {
		t.Logf("iter %d: %2v/%2v '%s' %v", j, v.i, v.end, v.b, v.err)
		v.Skip()
		j++
	}

	assert.NoError(t, v.Err())
	assert.Equal(t, 2, j)

	t.Logf("iter _: %2v/%2v '%s' %v", v.i, v.end, v.b, v.err)
}

func TestReader(t *testing.T) {
	data := `{"a":{"b":[true,false,null],"c":false},"d":true,"e":null} {}`
	t.Logf("data %d: '%s'", len(data), data)
	t.Logf("____   : '%s'", string("_123456789_123456789_123456789_123456789_123456789_123456789_")[:len(data)])

	r := strings.NewReader(data)
	v := ReadBufferSize(r, 1)

	j := 0
	for v.Type() != None {
		t.Logf("iter %d: %2v + %2v/%2v '%s' %v", j, v.ref, v.i, v.end, v.b, v.err)
		v.Skip()
		j++
	}

	assert.Error(t, v.Err(), io.EOF.Error())
	assert.Equal(t, 2, j)

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", v.ref, v.i, v.end, v.b, v.err)
}

func TestGetObjects(t *testing.T) {
	data := `{"a":{"b":{"c":"d"},"e":{"f":"g"}}}`
	t.Logf("data %d: '%s'", len(data), data)
	t.Logf("____   : '%s'", string("_123456789_123456789_123456789_123456789_123456789_123456789_")[:len(data)])

	rd := strings.NewReader(data)
	r := ReadBufferSize(rd, 10)

	r.Get("a", "e", "f")
	assert.NoError(t, r.Err())
	assert.Equal(t, String, r.Type())

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", r.ref, r.i, r.end, r.b, r.err)

	assert.Equal(t, "g", string(r.NextString()))

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", r.ref, r.i, r.end, r.b, r.err)
}

func TestGetArrays(t *testing.T) {
	data := `[[[1,2,3],[4,5]],[6,7,[8,[9,10],11]]]`
	t.Logf("data %d: '%s'", len(data), data)
	t.Logf("____   : '%s'", string("_123456789_123456789_123456789_123456789_123456789_123456789_")[:len(data)])

	r := strings.NewReader(data)
	v := ReadBufferSize(r, 10)

	v.Get(1, 2, 1, 1)
	assert.NoError(t, v.Err())
	assert.Equal(t, Number, v.Type())

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", v.ref, v.i, v.end, v.b, v.err)

	assert.Equal(t, "10", string(v.NextNumber()))

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", v.ref, v.i, v.end, v.b, v.err)
}

func TestGet2(t *testing.T) {
	data := `{"a":{"b":[true,false,null],"c":false},"d":{"eee":[{"a":1,"c":{"val":"not_result"}},{"a":1,"b":[1,2,3],"c":{"val":"result"}}]},"e":null}`
	t.Logf("data %d: '%s'", len(data), data)
	t.Logf("____   : '%s'", string("_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_")[:len(data)])

	r := strings.NewReader(data)
	v := ReadBufferSize(r, 10)

	v.Get("d", "eee", 1, "c", "val")
	assert.NoError(t, v.Err())
	assert.Equal(t, String, v.Type())
	assert.Equal(t, "result", string(v.NextString()))

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", v.ref, v.i, v.end, v.b, v.err)
}
