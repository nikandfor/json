package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsmSkipSpaces(t *testing.T) {
	b := []byte("   \t\n\t\n3")
	i := skipSpaces(b, 0)
	assert.Equal(t, len(b)-1, i)

	i = skipSpaces(b, len(b))
	assert.Equal(t, len(b), i)

	b = []byte("  \t\t\n\n\t\t    \t")
	i = skipSpaces(b, 3)
	assert.Equal(t, len(b), i)

	b = []byte(" 1  \t\t\n\n\t\t    \t")
	i = skipSpaces(b, 3)
	assert.Equal(t, len(b), i)
}
