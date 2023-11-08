package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelect(t *testing.T) {
	data := `[1,null,"a",[4],{"b":true}]`

	f := NewPipe(
		Iter{},
		&Select{},
	)

	w, i, err := f.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `1
"a"
[4]
{"b":true}`, string(w))
}

func TestMap(t *testing.T) {
	data := `[1,null,"a",[4],{"b":true}]`

	f := Map{
		Filter: Comma{
			Literal("5"),
			//	Literal("6"),
		},
	}

	w, i, err := f.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `[5,5,5,5,5]`, string(w))
}
