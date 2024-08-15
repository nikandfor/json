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
		if !assert.NoError(tb, err) {
			pref := r[:i]
			suff := r[i:]

			tb.Logf("%s|%s", pref, suff)
		}

		assert.Equal(tb, exp, string(w))

		if tb.Failed() {
			return
		}
	}

	w, i, state, err = f.Next(w[:0], r, i, state)
	if !assert.NoError(tb, err) {
		pref := r[:i]
		suff := r[i:]

		tb.Logf("%s|%s", pref, suff)
	}

	assert.Nil(tb, state)
	assert.Equal(tb, len(r), i)
}
