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
	assert.Equal(t, data, string(b))

	b, i, err = Index{"a"}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"b"`, string(b))

	b, i, err = Index{"c"}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `4`, string(b))

	b, i, err = Index{"non"}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `null`, string(b))

	b, i, err = Index{"d", 2}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `true`, string(b))

	b, i, err = Index{"d", 4, "f"}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `3.4`, string(b))
}

func TestComma(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	f := Comma{
		Index{"a"},
		Index{"c"},
		Index{"d", 0},
	}

	b, i, err := f.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"b"
4
"e"`, string(b))
}

func TestPipe(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	f := NewPipe(
		Index{"d"},
		Index{4},
		Index{"f"},
	)

	b, i, err := f.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, "3.4", string(b))

	assert.True(t, cap(f.Bufs[0]) != 0)
	assert.True(t, cap(f.Bufs[1]) != 0)
}
