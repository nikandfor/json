package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64(t *testing.T) {
	data := `"ab\ncd"`

	var e Base64

	res1, i, err := e.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"YWIKY2Q"`+"\n", string(res1))

	var d Base64d

	res2, i, err := d.Apply(nil, res1, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(res1)-1, i) // newline is not parsed
	assert.Equal(t, data+"\n", string(res2))
}
