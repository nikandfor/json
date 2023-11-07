package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	b, i, err := Dot{}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, data+"\n", string(b))

	b, i, err = Selector{"a"}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"b"`+"\n", string(b))

	b, i, err = Selector{"c"}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `4`+"\n", string(b))

	b, i, err = Selector{"non"}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `null`+"\n", string(b))

	b, i, err = Selector{"d", 2}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `true`+"\n", string(b))

	b, i, err = Selector{"d", 4, "f"}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `3.4`+"\n", string(b))
}

func TestComma(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	f := Comma{
		Selector{"a"},
		Selector{"c"},
		Selector{"d", 0},
	}

	b, i, err := f.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"b"
4
"e"
`, string(b))
}

func TestPipe(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	f := NewPipe(
		Selector{"d"},
		Selector{4},
		Selector{"f"},
	)

	b, i, err := f.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, "3.4\n", string(b))

	assert.True(t, cap(f.Bufs[0]) != 0)
	assert.True(t, cap(f.Bufs[1]) != 0)
}
