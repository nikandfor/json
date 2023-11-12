package jq

import (
	"bytes"
	"testing"

	"nikand.dev/go/json/benchmarks_data"
)

func BenchmarkIndex(b *testing.B) {
	for _, tc := range []struct {
		Name   string
		Data   []byte
		Result []byte
		Filter Filter
	}{
		{
			Name:   "Small",
			Data:   benchmarks_data.SmallFixture,
			Filter: &Index{"uuid"},
			Result: []byte(`"de305d54-75b4-431b-adb2-eb6b9e546014"`),
		},
		{
			Name:   "Medium",
			Data:   benchmarks_data.MediumFixture,
			Filter: &Index{"person", "gravatar", "avatars", 0, "url"},
			Result: []byte(`"http://1.gravatar.com/avatar/f7c8edd577d13b8930d5522f28123510"`),
		},
		{
			Name:   "Large",
			Data:   benchmarks_data.LargeFixture,
			Filter: &Index{"topics", "topics", 28, "posters", 0, "user_id"},
			Result: []byte(`52`),
		},
	} {
		tc := tc

		b.Run(tc.Name, func(b *testing.B) {
			b.ReportAllocs()

			var err error
			var res []byte

			for i := 0; i < b.N; i++ {
				res, _, _, err = tc.Filter.Next(res[:0], tc.Data, 0, nil)
				if err != nil {
					b.Errorf("index: %v", err)
				}
			}

			if !bytes.Equal(tc.Result, res) {
				b.Errorf("not equal: exp %q, got %q", tc.Result, res)
			}
		})
	}
}
