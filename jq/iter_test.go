package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIter(t *testing.T) {
	var state State
	var err error
	var w []byte
	var i int

	data := `[1,null,"a",[4],{"b":true}]`

	for _, exp := range []string{
		"1", "null", `"a"`, `[4]`, `{"b":true}`,
	} {
		w, i, state, err = Iter{}.Next(w[:0], []byte(data), i, state)
		assert.NoError(t, err)
		assert.NotNil(t, state)
		//	assert.Len(t, data, i)
		assert.Equal(t, exp, string(w))
	}

	w, i, state, err = Iter{}.Next(w[:0], []byte(data), i, state)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Len(t, data, i)
	assert.Equal(t, "", string(w))

	state = nil
	i = 0

	data = `{"a":"b","c": 4, "d": [null, true, false], "e": {"f": "g"}}`

	for _, exp := range []string{
		`"b"`, "4", `[null, true, false]`, `{"f": "g"}`,
	} {
		w, i, state, err = Iter{}.Next(w[:0], []byte(data), i, state)
		assert.NoError(t, err)
		assert.NotNil(t, state)
		//	assert.Len(t, data, i)
		assert.Equal(t, exp, string(w))
	}

	w, i, state, err = Iter{}.Next(w[:0], []byte(data), i, state)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Len(t, data, i)
	assert.Equal(t, "", string(w))

	state = nil
	i = 0

	data = `{}`

	w, i, state, err = Iter{}.Next(w[:0], []byte(data), i, state)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Len(t, data, i)
	assert.Equal(t, "", string(w))

	state = nil
	i = 0

	data = `[]`

	w, i, state, err = Iter{}.Next(w[:0], []byte(data), i, state)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Len(t, data, i)
	assert.Equal(t, "", string(w))
}
