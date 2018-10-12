package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt(t *testing.T) {
	assert.Equal(t, 4, WrapString(`4`).MustInt())
	assert.Equal(t, 100, WrapString(`100`).MustInt())
	assert.Equal(t, 123456789, WrapString(`123456789`).MustInt())
}

func TestUint(t *testing.T) {
	assert.Equal(t, uint(4), WrapString(`4`).MustUint())
	assert.Equal(t, uint(100), WrapString(`100`).MustUint())
	assert.Equal(t, uint(123456789), WrapString(`123456789`).MustUint())
}
