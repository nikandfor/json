package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	data := `"\"abcd\"" "1" "{\"a\":\"b\"}"`

	var e JSONDecoder

	res, i, _, err := e.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, `"abcd"`, string(res))

	res, i, _, err = e.Next(nil, []byte(data), i, nil)
	assert.NoError(t, err)
	assert.Equal(t, `1`, string(res))

	res, i, _, err = e.Next(nil, []byte(data), i, nil)
	assert.NoError(t, err)
	assert.Equal(t, `{"a":"b"}`, string(res))

	assert.Equal(t, len(data), i)
}
