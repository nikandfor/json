package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndex(tb *testing.T) {
	r := []byte(`{"a": [{"b":[1, "2"]}, {"b":[3]}, {"b": {"c": 4, "d": "5"}}]}`)

	f := NewIndex("a", Iter{}, "b", Iter{})

	var w []byte
	var err error
	var state any
	i := 0

	for j, exp := range []string{`1`, `"2"`, `3`, `4`, `"5"`} {
		assert.Equal(tb, j == 0, state == nil)

		w, i, state, err = f.Next(w[:0], r, i, state)
		assertBytesErr(tb, r, i, err)

		assert.Equal(tb, exp, string(w))

		if tb.Failed() {
			return
		}
	}

	w, i, state, err = f.Next(w[:0], r, i, state)
	assertBytesErr(tb, r, i, err)

	assert.Nil(tb, state)
	assert.Equal(tb, len(r), i)
}

func TestIndexEmpty(tb *testing.T) {
	r := []byte(`{"results":[],"key":"b"}`)
	f := NewIndex()

	w, i, err := NextAll(f, nil, r, 0, nil)
	assertBytesErr(tb, r, i, err)
	assert.Equal(tb, len(r), i)
	assert.Equal(tb, r, w)
}

func TestIndexIterEmpty(tb *testing.T) {
	var w []byte
	r := []byte(`{"a":[],"b":{}}`)

	f := NewIndex("a", Iter{})

	w, _, state, err := f.Next(w[:0], r, 0, nil)
	assert.NoError(tb, err)
	assert.Nil(tb, state)
	assert.Len(tb, w, 0)

	f = NewIndex("b", Iter{})

	w, _, state, err = f.Next(w[:0], r, 0, nil)
	assert.NoError(tb, err)
	assert.Nil(tb, state)
	assert.Len(tb, w, 0)
}

func assertBytesErr(tb *testing.T, r []byte, i int, err error) {
	if assert.NoError(tb, err) {
		return
	}

	pref := r[:i]
	suff := r[i:]

	tb.Logf("%s|%s", pref, suff)
}
