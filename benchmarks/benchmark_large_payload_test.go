package go_benchmark

import (
	"testing"

	"github.com/buger/jsonparser"
	"github.com/json-iterator/go"
	"github.com/nikandfor/json"
)

// buger
func BenchmarkBugerJsonParserLarge(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(LargeFixture)))
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
	b.SetBytes(int64(len(LargeFixture)))
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

// nikandfor
func BenchmarkNikandjsonManualLarge(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(LargeFixture)))
	var w json.Reader
	for i := 0; i < b.N; i++ {
		w.Reset(LargeFixture)
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

func BenchmarkNikandjsonSearchLarge(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(LargeFixture)))
	var w json.Reader
	for i := 0; i < b.N; i++ {
		w.Reset(LargeFixture)
		w.Search("topics", "topics")
		count := 0
		for w.HasNext() {
			count++
			w.Skip()
		}
	}
}

// compare profiles
func BenchmarkCmpLarge(b *testing.B) {
	b.ReportAllocs()
	var w json.Reader
	for i := 0; i < b.N; i++ {
		count := 0
		jsonparser.ArrayEach(LargeFixture, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			count++
		}, "topics", "topics")

		w.Reset(LargeFixture)
		w.Search("topics", "topics")
		for w.HasNext() {
			count--
			w.Skip()
		}
	}
}
