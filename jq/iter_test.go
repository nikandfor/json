package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIter(t *testing.T) {
	data := `[1,null,"a",[4],{"b":true}]`

	w, i, err := Iter{}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `1
null
"a"
[4]
{"b":true}
`, string(w))

	data = `{"a":"b","c": 4, "d": [null, true, false], "e": {"f": "g"}}`

	w, i, err = Iter{}.Apply(nil, []byte(data), 0)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"b"
4
[null, true, false]
{"f": "g"}
`, string(w))
}
