package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyIndex(tb *testing.T) {
	data := []byte(`{"a":[{"b":"c"},{"b":"d"}]}`)

	var w []byte

	l0 := len(w)
	w, _, _, err := Key("a").Next(w, data, 0, nil)
	assert.NoError(tb, err)

	tb.Logf("key a: (%s) %v", w, err)

	l1 := len(w)
	w, _, _, err = Index(1).Next(w, w, l0, nil)
	assert.NoError(tb, err)

	tb.Logf("key a: (%s) %v", w, err)

	l2 := len(w)
	w, _, _, err = NewEqual(Key("b"), Literal(`"d"`)).Next(w, w, l1, nil)
	assert.NoError(tb, err)

	tb.Logf("key a: (%s) %v", w, err)

	assert.Equal(tb, `true`, string(w[l2:]))
}
