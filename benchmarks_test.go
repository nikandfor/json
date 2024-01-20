package json

import (
	"testing"

	"nikand.dev/go/json/benchmarks_data"
)

func BenchmarkDecoderSkip(b *testing.B) {
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

func BenchmarkReaderSkip(b *testing.B) {
	var r Reader

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

			r.Reset(tc.Data, nil)

			defer r.Unlock()
			r.Lock()

			for i := 0; i < b.N; i++ {
				r.Rewind()

				err := r.Skip()
				if err != nil {
					b.Errorf("skip: %v", err)
				}
			}
		})
	}
}
