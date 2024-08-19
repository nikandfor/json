package jval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessage(tb *testing.T) {
	var b []byte
	var m Message

	for _, d := range []string{
		`null`,
		`true`,
		`false`,
		`0`, `1`, `10`, `55`,
		`""`, `"abc"`, `"abcdefj01234567890"`,
		`[]`, `[null,true,false,[0,5],"str"]`,
		`{}`, `{"a":"b"}`, `{"a":"b","c":["d",null],"f":[1,2,[3,4]],"g":{"i":"j"}}`,
	} {
		i, err := m.Decode([]byte(d), 0)
		assert.NoError(tb, err, `(%s)`, d)
		assert.Equal(tb, len(d), i, `(%s)`, d)

		raw, root := m.BytesRoot()
		tb.Logf("(%s) -> @%x/%x % 2x", d, root, len(raw), raw)

		b = m.Encode(b[:0])

		assert.Equal(tb, d, string(b), "(%s)", b)
	}
}
