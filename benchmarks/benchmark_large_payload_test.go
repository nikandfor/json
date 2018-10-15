package go_benchmark

import (
	"testing"

	"github.com/buger/jsonparser"
	"github.com/json-iterator/go"
	"github.com/nikandfor/json"
	"github.com/stretchr/testify/assert"
)

func BenchmarkJsonParserLarge(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		count := 0
		jsonparser.ArrayEach(LargeFixture, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			count++
		}, "topics", "topics")
	}
}

// jsoniter
func BenchmarkJsoniterLarge(b *testing.B) {
	iter := jsoniter.ParseBytes(jsoniter.ConfigDefault, LargeFixture)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.ResetBytes(LargeFixture)
		count := 0
		for field := iter.ReadObject(); field != ""; field = iter.ReadObject() {
			if "topics" != field {
				iter.Skip()
				continue
			}
			for field := iter.ReadObject(); field != ""; field = iter.ReadObject() {
				if "topics" != field {
					iter.Skip()
					continue
				}
				for iter.ReadArray() {
					iter.Skip()
					count++
				}
				break
			}
			break
		}
	}
}

func BenchmarkEncodingJsonLarge(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		payload := &LargePayload{}
		json.Unmarshal(LargeFixture, payload)
	}
}

// nikandfor
func BenchmarkNikandjsonLarge(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := json.Wrap(LargeFixture)
		count := 0
		for w.HasNext() {
			ok := w.CompareKey([]byte("topics"))
			if !ok {
				w.Skip()
				continue
			}
			for w.HasNext() {
				ok := w.CompareKey([]byte("topics"))
				if !ok {
					w.Skip()
					continue
				}
				for w.HasNext() {
					count++
					w.Skip()
				}
				break
			}
			break
		}
	}
}

func BenchmarkNikandjsonGetLarge(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := json.Wrap(LargeFixture)
		w.Get("topics", "topics")
		count := 0
		for w.HasNext() {
			count++
			w.Skip()
		}
		assert.Equal(b, 30, count)
	}
}
