package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObject(tb *testing.T) {
	r := []byte(`{"a":1, "b": {"c": ["d", 4]}}`)

	for _, tc := range []struct {
		Key ObjectKey
		Exp string
	}{{
		Key: ObjectKey{Key: "create", Filter: Literal(`"new"`)},
		Exp: `{"create":"new"}`,
	}, {
		Key: ObjectKey{Key: "dot", Filter: Dot{}},
		Exp: `{"dot":` + string(r) + `}`,
	}, {
		Key: ObjectKey{Key: "index.a", Filter: NewIndex("a")},
		Exp: `{"index.a":1}`,
	}, {
		Key: ObjectKey{Key: "index.b.c.1", Filter: NewIndex("b", "c", 1)},
		Exp: `{"index.b.c.1":4}`,
	}} {
		f := Object{
			Keys: []ObjectKey{tc.Key},
		}

		w, i, state, err := f.Next(nil, r, 0, nil)
		assert.NoError(tb, err)
		assert.Nil(tb, state)
		assert.Equal(tb, len(r), i)

		assert.Equal(tb, []byte(tc.Exp), w)

		if tb.Failed() {
			break
		}
	}
}

func TestObjectMulti(tb *testing.T) {
	tb.Run("one", func(tb *testing.T) {
		r := []byte(`[1,2,"3"]`)
		f := Object{Keys: []ObjectKey{
			{Key: "a", Filter: Iter{}},
		}}

		var w []byte
		var err error
		var state State
		i := 0

		for _, exp := range []string{`{"a":1}`, `{"a":2}`, `{"a":"3"}`} {
			w, i, state, err = f.Next(w[:0], r, i, state)
			tb.Logf("obj: %s  next state %+v", w, state)
			assert.NoError(tb, err)
			assert.NotNil(tb, state)

			assert.Equal(tb, []byte(exp), w)
		}

		w, i, state, err = f.Next(w[:0], r, 0, state)
		assert.NoError(tb, err)
		assert.Nil(tb, state)
		assert.Equal(tb, len(r), i)
		assert.Equal(tb, []byte{}, w)
	})

	tb.Run("two", func(tb *testing.T) {
		r := []byte(`{"a":[1,2,"3"],"b":[4,null]}`)
		f := Object{Keys: []ObjectKey{
			{Key: "a", Filter: NewIndex("a", Iter{})},
			{Key: "b", Filter: NewIndex("b", Iter{})},
		}}

		var w []byte
		var err error
		var state State
		i := 0

		for _, exp := range []string{
			`{"a":1,"b":4}`,
			`{"a":1,"b":null}`,
			`{"a":2,"b":4}`,
			`{"a":2,"b":null}`,
			`{"a":"3","b":4}`,
			`{"a":"3","b":null}`,
		} {
			w, i, state, err = f.Next(w[:0], r, i, state)
			tb.Logf("obj: %s  next state %+v", w, state)
			assert.NoError(tb, err)
			assert.NotNil(tb, state)

			assert.Equal(tb, []byte(exp), w)
		}

		w, i, state, err = f.Next(w[:0], r, 0, state)
		assert.NoError(tb, err)
		assert.Nil(tb, state)
		assert.Equal(tb, len(r), i)
		assert.Equal(tb, []byte{}, w)
	})

	tb.Run("three", func(tb *testing.T) {
		r := []byte(`{"a":[1,2,"3"],"d":true,"b":[4,null]}`)
		f := Object{Keys: []ObjectKey{
			{Key: "a", Filter: NewIndex("a", Iter{})},
			{Key: "d", Filter: NewIndex("d")},
			{Key: "b", Filter: NewIndex("b", Iter{})},
		}}

		var w []byte
		var err error
		var state State
		i := 0

		for _, exp := range []string{
			`{"a":1,"d":true,"b":4}`,
			`{"a":1,"d":true,"b":null}`,
			`{"a":2,"d":true,"b":4}`,
			`{"a":2,"d":true,"b":null}`,
			`{"a":"3","d":true,"b":4}`,
			`{"a":"3","d":true,"b":null}`,
		} {
			w, i, state, err = f.Next(w[:0], r, i, state)
			tb.Logf("obj: %s  next state %+v", w, state)
			assert.NoError(tb, err)
			assert.NotNil(tb, state)

			assert.Equal(tb, []byte(exp), w)
		}

		w, i, state, err = f.Next(w[:0], r, 0, state)
		assert.NoError(tb, err)
		assert.Nil(tb, state)
		assert.Equal(tb, len(r), i)
		assert.Equal(tb, []byte{}, w)
	})
}
