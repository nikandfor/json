package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCat(t *testing.T) {
	f := Cat{
		Separator: []byte("-"),
	}

	data := `"ama", "ena", "uma", "viva"`

	w, i, state, err := f.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Len(t, data, i)
	assert.Equal(t, `"ama-ena-uma-viva"`, string(w))

	data = `"\nqwe 世"  "界\tend"`

	w, i, state, err = f.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Len(t, data, i)
	assert.Equal(t, `"\nqwe 世-界\tend"`, string(w))
}
