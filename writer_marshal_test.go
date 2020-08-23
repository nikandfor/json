// +build ignore

package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriterStreamStructs(t *testing.T) {
	w := NewWriter(make([]byte, 1000))

	w.Marshal(struct{ A string }{"str"})
	w.NewLine()
	w.Marshal(struct{ A string }{"str"})
	w.NewLine()
	w.Marshal(struct{ A string }{"str"})
	w.NewLine()

	assert.Equal(t, []byte(`{"A":"str"}
{"A":"str"}
{"A":"str"}
`), w.Bytes())

	//	t.Logf("res: '%s'", w.Bytes())
}
