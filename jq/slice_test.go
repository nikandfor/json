package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceArray(t *testing.T) {
	data := `[1,null,"a",[4],{"b":true}]`

	w, i, err := Length{}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, "5\n", string(w))

	w, i, err = (&Slice{L: 0, R: 100}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, data+"\n", string(w))

	w, i, err = (&Slice{L: 1, R: 4}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[null,"a",[4]]`+"\n", string(w))

	w, i, err = (&Slice{L: 3, R: 3}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[]`+"\n", string(w))

	w, i, err = (&Slice{L: 3, R: 2}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[]`+"\n", string(w))

	w, i, err = (&Slice{L: 3, R: 2, Circle: true}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[[4],{"b":true},1,null]`+"\n", string(w))

	w, i, err = (&Slice{L: -4, R: -1}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[null,"a",[4]]`+"\n", string(w))

	w, i, err = (&Slice{L: -2, R: 2, Circle: true}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[[4],{"b":true},1,null]`+"\n", string(w))

	w, i, err = Array{}.Apply(nil, []byte(data[1:len(data)-1]), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data)-2, i)
	assert.Equal(t, data+"\n", string(w))
}

func TestSliceString(t *testing.T) {
	data := `"abcde"`

	w, i, err := Length{}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, "5\n", string(w))

	w, i, err = (&Slice{L: 0, R: 100}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"abcde"`+"\n", string(w))

	w, i, err = (&Slice{L: 1, R: 4}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"bcd"`+"\n", string(w))

	w, i, err = (&Slice{L: 3, R: 3}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `""`+"\n", string(w))

	w, i, err = (&Slice{L: 3, R: 2}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `""`+"\n", string(w))

	w, i, err = (&Slice{L: 3, R: 2, Circle: true}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"deab"`+"\n", string(w))

	w, i, err = (&Slice{L: -4, R: -1}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"bcd"`+"\n", string(w))

	w, i, err = (&Slice{L: -2, R: 2, Circle: true}).Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"deab"`+"\n", string(w))
}
