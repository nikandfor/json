package go_benchmark

import (
	"bytes"
	json "encoding/json"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	nikjson "github.com/nikandfor/json"
)

// encode/json
func BenchmarkDecodeStdStructMedium(b *testing.B) {
	b.ReportAllocs()
	var data MediumPayload
	for i := 0; i < b.N; i++ {
		json.Unmarshal(mediumFixture, &data)
	}
}

func BenchmarkEncodeStdStructMedium(b *testing.B) {
	var data MediumPayload
	json.Unmarshal(mediumFixture, &data)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		json.Marshal(data)
	}
}

// jsoniter
func BenchmarkDecodeJsoniterStructMedium(b *testing.B) {
	b.ReportAllocs()
	var data MediumPayload
	for i := 0; i < b.N; i++ {
		jsoniter.Unmarshal(mediumFixture, &data)
	}
}

func BenchmarkEncodeJsoniterStructMedium(b *testing.B) {
	var data MediumPayload
	jsoniter.Unmarshal(mediumFixture, &data)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		jsoniter.Marshal(data)
	}
}

// easyjson
func BenchmarkDecodeEasyJsonMedium(b *testing.B) {
	b.ReportAllocs()
	var data MediumPayload
	for i := 0; i < b.N; i++ {
		lexer := &jlexer.Lexer{Data: mediumFixture}
		data.UnmarshalEasyJSON(lexer)
	}
}

func BenchmarkEncodeEasyJsonMedium(b *testing.B) {
	var data MediumPayload
	lexer := &jlexer.Lexer{Data: mediumFixture}
	data.UnmarshalEasyJSON(lexer)
	b.ReportAllocs()
	buf := &bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		writer := &jwriter.Writer{}
		data.MarshalEasyJSON(writer)
		buf.Reset()
		writer.DumpTo(buf)
	}
}

// nikandfor
func BenchmarkDecodeNikandjsonV0StructMedium(b *testing.B) {
	b.ReportAllocs()
	var data MediumPayload
	for i := 0; i < b.N; i++ {
		nikjson.UnmarshalV0(mediumFixture, &data)
	}
}

func BenchmarkDecodeNikandjsonStructMedium(b *testing.B) {
	b.ReportAllocs()
	var data MediumPayload
	for i := 0; i < b.N; i++ {
		nikjson.Unmarshal(mediumFixture, &data)
	}
}

func BenchmarkDecodeNikandjsonIterStructMedium(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		it, _ := nikjson.Wrap(mediumFixture).ObjectIter()
		for it.HasNext() {
			_, _ = it.Next()
		}
	}
}

func BenchmarkDecodeNikandjsonGetStructMedium(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := nikjson.Wrap(mediumFixture)
		_, _ = v.Get("person", "gravatar", "avatars", 0, "url")
	}
}

func BenchmarkEncodeNikandjsonV0StructMedium(b *testing.B) {
	var data MediumPayload
	nikjson.UnmarshalV0(mediumFixture, &data)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		nikjson.MarshalV0(data)
	}
}
