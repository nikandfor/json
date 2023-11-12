package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelect(t *testing.T) {
	var state State
	var err error
	var w []byte
	var i int

	data := `1 null "a" [4] {"b":true}`

	for _, exp := range []string{
		"1", ``, `"a"`, `[4]`, `{"b":true}`,
	} {
		w, i, state, err = (&Select{}).Next(w[:0], []byte(data), i, state)
		assert.NoError(t, err)
		//	assert.Len(t, data, i)
		assert.Equal(t, exp, string(w))
	}

	assert.Nil(t, state)
}

func TestMap(t *testing.T) {
	data := `[1,null,"a",[4],{"b":true}]`

	f := Map{
		Filter: NewComma(
			Literal(`5`),
		),
	}

	w, i, state, err := f.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Len(t, data, i)
	assert.Equal(t, `[5,5,5,5,5]`, string(w))

	f = Map{
		Filter: NewComma(
			Literal(`5`),
			Literal(`6`),
		),
	}

	w, i, state, err = f.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Len(t, data, i)
	assert.Equal(t, `[5,6,5,6,5,6,5,6,5,6]`, string(w))
}
