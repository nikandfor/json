package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayIterEmpty(t *testing.T) {
	s := `[]`
	v := WrapString(s)

	tp, err := v.Type()
	if assert.NoError(t, err) {
		assert.Equal(t, Array, tp, "%v", tp)
	}

	it, err := v.ArrayIter()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, false, it.HasNext(), "]")
}

func TestArrayIter(t *testing.T) {
	s := `[1,2,"3",{"a":"a"}]`
	v := WrapString(s)

	tp, err := v.Type()
	if assert.NoError(t, err) {
		assert.Equal(t, Array, tp, "%v", tp)
	}

	it, err := v.ArrayIter()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, true, it.HasNext(), "1")
	assert.Equal(t, 1, it.Next().MustInt())
	assert.Equal(t, true, it.HasNext(), "2")
	assert.Equal(t, 2, it.Next().MustInt())
	assert.Equal(t, true, it.HasNext(), "3")
	assert.Equal(t, "3", it.Next().String())
	assert.Equal(t, true, it.HasNext(), "obj")
	assert.Equal(t, `{"a":"a"}`, it.Next().String())
	assert.Equal(t, false, it.HasNext(), "]")
	assert.Equal(t, false, it.HasNext(), "]")
}

func TestArrayIterSpaces(t *testing.T) {
	s := ` [ 1 , 2 , "3" , { "a" : "a" } ] `
	v := WrapString(s)

	tp, err := v.Type()
	if assert.NoError(t, err) {
		assert.Equal(t, Array, tp, "%v", tp)
	}

	it, err := v.ArrayIter()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, true, it.HasNext(), "1")
	assert.Equal(t, 1, it.Next().MustInt())
	assert.Equal(t, true, it.HasNext(), "2")
	assert.Equal(t, 2, it.Next().MustInt())
	assert.Equal(t, true, it.HasNext(), "3")
	assert.Equal(t, "3", it.Next().String())
	assert.Equal(t, true, it.HasNext(), "obj")
	assert.Equal(t, `{ "a" : "a" }`, it.Next().String())
	assert.Equal(t, false, it.HasNext(), "]")
	assert.Equal(t, false, it.HasNext(), "]")
}

func TestObjectIterEmpty(t *testing.T) {
	s := `{}`
	v := WrapString(s)

	tp, err := v.Type()
	if assert.NoError(t, err) {
		assert.Equal(t, Object, tp, "%v", tp)
	}

	it, err := v.ObjectIter()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, false, it.HasNext(), "}")
}

func TestObjectIter(t *testing.T) {
	s := `{"a":1,"b":"b_val","c":true,"d":{"a":"a"}}`
	v := WrapString(s)

	tp, err := v.Type()
	if assert.NoError(t, err) {
		assert.Equal(t, Object, tp, "%v", tp)
	}

	it, err := v.ObjectIter()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, true, it.HasNext(), "a")
	k, val := it.Next()
	assert.Equal(t, "a", k.MustCheckString())
	assert.Equal(t, 1, val.MustInt())
	assert.Equal(t, true, it.HasNext(), "b")
	k, val = it.Next()
	assert.Equal(t, "b", k.MustCheckString())
	assert.Equal(t, "b_val", val.MustCheckString())
	assert.Equal(t, true, it.HasNext(), "c")
	k, val = it.Next()
	assert.Equal(t, "c", k.MustCheckString())
	assert.Equal(t, true, val.MustBool())
	assert.Equal(t, true, it.HasNext(), "d")
	k, val = it.Next()
	assert.Equal(t, "d", k.MustCheckString())
	assert.Equal(t, `{"a":"a"}`, val.String())
	assert.Equal(t, false, it.HasNext(), "}")
	assert.Equal(t, false, it.HasNext(), "}")
}

func TestObjectIterSpaces(t *testing.T) {
	s := ` { "a" : 1 , "b" : "b_val" , "c" : true , "d" : {"a":"a"} } `
	v := WrapString(s)

	tp, err := v.Type()
	if assert.NoError(t, err) {
		assert.Equal(t, Object, tp, "%v", tp)
	}

	it, err := v.ObjectIter()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, true, it.HasNext(), "a")
	k, val := it.Next()
	assert.Equal(t, "a", k.MustCheckString())
	assert.Equal(t, 1, val.MustInt())
	assert.Equal(t, true, it.HasNext(), "b")
	k, val = it.Next()
	assert.Equal(t, "b", k.MustCheckString())
	assert.Equal(t, "b_val", val.MustCheckString())
	assert.Equal(t, true, it.HasNext(), "c")
	k, val = it.Next()
	assert.Equal(t, "c", k.MustCheckString())
	assert.Equal(t, true, val.MustBool())
	assert.Equal(t, true, it.HasNext(), "d")
	k, val = it.Next()
	assert.Equal(t, "d", k.MustCheckString())
	assert.Equal(t, `{"a":"a"}`, val.String())
	assert.Equal(t, false, it.HasNext(), "}")
	assert.Equal(t, false, it.HasNext(), "}")
}
