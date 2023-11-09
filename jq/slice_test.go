package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceArray(t *testing.T) {
	data := `[1,null,"a",[4],{"b":true}]`

	w, i, _, err := Length{}.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, "5", string(w))

	w, i, _, err = (&Slice{L: 0, R: 100}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, data, string(w))

	w, i, _, err = (&Slice{L: 1, R: 4}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[null,"a",[4]]`, string(w))

	w, i, _, err = (&Slice{L: 3, R: 3}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[]`, string(w))

	w, i, _, err = (&Slice{L: 3, R: 2}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[]`, string(w))

	w, i, _, err = (&Slice{L: 3, R: 2, Circle: true}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[[4],{"b":true},1,null]`, string(w))

	w, i, _, err = (&Slice{L: -4, R: -1}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[null,"a",[4]]`, string(w))

	w, i, _, err = (&Slice{L: -2, R: 2, Circle: true}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[[4],{"b":true},1,null]`, string(w))
}

func TestSliceString(t *testing.T) {
	data := `"abcde"`

	w, i, _, err := Length{}.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, "5", string(w))

	w, i, _, err = (&Slice{L: 0, R: 100}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"abcde"`, string(w))

	w, i, _, err = (&Slice{L: 1, R: 4}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"bcd"`, string(w))

	w, i, _, err = (&Slice{L: 3, R: 3}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `""`, string(w))

	w, i, _, err = (&Slice{L: 3, R: 2}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `""`, string(w))

	w, i, _, err = (&Slice{L: 3, R: 2, Circle: true}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"deab"`, string(w))

	w, i, _, err = (&Slice{L: -4, R: -1}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"bcd"`, string(w))

	w, i, _, err = (&Slice{L: -2, R: 2, Circle: true}).Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"deab"`, string(w))
}

func TestArray(t *testing.T) {
	tf := TestFilter{
		[]byte(`{"a":"b"}`),
		[]byte(`{"a":"c"}`),
		[]byte(`{"a":3}`),
	}

	f := Array{Filter: tf}

	w, i, state, err := f.Next(nil, []byte(`null`), 0, nil)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Equal(t, 4, i)
	assert.Equal(t, `[{"a":"b"},{"a":"c"},{"a":3}]`, string(w))

	//

	f = Array{Filter: NewComma(Dot{}, Literal(`1`))}

	w, i, state, err = f.Next(nil, []byte(`null`), 0, nil)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Equal(t, 4, i)
	assert.Equal(t, `[null,1]`, string(w))
}
