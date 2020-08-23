// +build ignore

package go_benchmark

import (
	"testing"

	"github.com/nikandfor/json"
	"github.com/stretchr/testify/assert"
)

func TestDecodeNikandjsonStructMedium(t *testing.T) {
	var data MediumPayload
	err := json.Unmarshal(MediumFixture, &data)
	assert.NoError(t, err)
}

func BenchmarkDecodeNikandjsonStructMedium(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(MediumFixture)))
	var err error
	var data MediumPayload
	var r json.Reader
	for i := 0; i < b.N; i++ {
		err = r.Reset(MediumFixture).Unmarshal(&data)
	}
	assert.NoError(b, err)
}
