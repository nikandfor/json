package json

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nikand.dev/go/json/benchmarks_data"
)

func BenchmarkSkip(b *testing.B) {
	var d Decoder

	for _, tc := range []struct {
		Name string
		Data []byte
	}{
		{Name: "Small", Data: benchmarks_data.SmallFixture},
		{Name: "Medium", Data: benchmarks_data.MediumFixture},
		{Name: "Large", Data: benchmarks_data.LargeFixture},
	} {
		tc := tc

		b.Run(tc.Name, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := d.Skip(tc.Data, 0)
				if err != nil {
					b.Errorf("skip: %v", err)
				}
			}
		})
	}
}

func BenchmarkUnmarshal(tb *testing.B) {
	bench := func(tb *testing.B, b []byte, x interface{}) {
		var err error
		var d Decoder

		for i := 0; i < tb.N; i++ {
			_, err = d.Unmarshal(b, 0, x)
		}

		assert.NoError(tb, err)
	}

	tb.Run("Small", func(tb *testing.B) {
		tb.ReportAllocs()

		bench(tb, benchmarks_data.SmallFixture, new(benchmarks_data.SmallPayload))
	})

	tb.Run("Medium", func(tb *testing.B) {
		tb.ReportAllocs()

		bench(tb, benchmarks_data.MediumFixture, new(benchmarks_data.MediumPayload))
	})

	tb.Run("Large", func(tb *testing.B) {
		tb.ReportAllocs()

		bench(tb, benchmarks_data.LargeFixture, new(benchmarks_data.LargePayload))
	})
}
