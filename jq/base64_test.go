package jq

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64(t *testing.T) {
	data := `"ab\ncd"`

	e := Base64{
		Encoding: base64.RawStdEncoding,
	}

	res1, i, _, err := e.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"YWIKY2Q"`, string(res1))

	d := Base64d{
		Encoding: base64.RawStdEncoding,
	}

	res2, i, _, err := d.Next(nil, res1, 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(res1), i)
	assert.Equal(t, data, string(res2))
}
