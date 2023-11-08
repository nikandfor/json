package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	data := `"\"abcd\"" "1" "{\"a\":\"b\"}"`

	var e JSONDecoder

	res, i, err := e.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, `"abcd"`, string(res))

	res, i, err = e.Apply(nil, []byte(data), i)
	assert.NoError(t, err)
	assert.Equal(t, `1`, string(res))

	res, i, err = e.Apply(nil, []byte(data), i)
	assert.NoError(t, err)
	assert.Equal(t, `{"a":"b"}`, string(res))

	assert.Equal(t, len(data), i)
}
