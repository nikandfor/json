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

	w, i, err = (&Array{}).Apply(nil, []byte("1\n2"), 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, i)
	assert.Equal(t, "[]\n", string(w))
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

func TestArray(t *testing.T) {
	var f Filter
	data := `[{"a":"b"}, {"a":"c"}, {"a":3}]`

	f = NewPipe(
		Iter{},
		Selector{"a"},
	)

	w, i, err := f.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, "\"b\"\n\"c\"\n3\n", string(w))

	f = &Array{
		Filter: NewPipe(
			Iter{},
			Selector{"a"},
		),
	}

	w, i, err = f.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `["b","c",3]`+"\n", string(w))

	f = NewPipe(
		&Array{
			Filter: NewPipe(
				Iter{},
				Selector{"a"},
			),
		},
	)

	w, i, err = f.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `["b","c",3]`+"\n", string(w))
}
