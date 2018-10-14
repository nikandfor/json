package go_benchmark

import (
	"bytes"
	json "encoding/json"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	nikv2 "github.com/nikandfor/json"
	nikv1 "github.com/nikandfor/json/v1"
	"github.com/stretchr/testify/assert"
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
func TestDecodeNikandjsonGetStructMedium(t *testing.T) {
	v := nikv1.Wrap(mediumFixture)
	res, err := v.Get("person", "gravatar", "avatars", 0, "url")
	assert.NoError(t, err)
	assert.Equal(t, []byte("http://1.gravatar.com/avatar/f7c8edd577d13b8930d5522f28123510"), res.Bytes())
}

func BenchmarkDecodeNikandjsonV0StructMedium(b *testing.B) {
	b.ReportAllocs()
	var data MediumPayload
	for i := 0; i < b.N; i++ {
		nikv1.UnmarshalV0(mediumFixture, &data)
	}
}

func BenchmarkDecodeNikandjsonStructMedium(b *testing.B) {
	b.ReportAllocs()
	var data MediumPayload
	for i := 0; i < b.N; i++ {
		nikv1.Unmarshal(mediumFixture, &data)
	}
}

func BenchmarkDecodeNikandjsonIterStructMedium(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		it, _ := nikv1.Wrap(mediumFixture).ObjectIter()
		for it.HasNext() {
			_, _ = it.Next()
		}
	}
}

func BenchmarkDecodeNikandjsonGetStructMedium(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := nikv1.Wrap(mediumFixture)
		_, _ = v.Get("person", "gravatar", "avatars", 0, "url")
	}
}

func BenchmarkEncodeNikandjsonV0StructMedium(b *testing.B) {
	var data MediumPayload
	nikv1.UnmarshalV0(mediumFixture, &data)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		nikv1.MarshalV0(data)
	}
}

// nikandfor v2
func TestDecodeNikandjsonV2StructMedium(t *testing.T) {
	var data MediumPayload
	err := nikv2.Unmarshal(mediumFixture, &data)
	if !assert.NoError(t, err) {
		return
	}
	var exp MediumPayload
	err = json.Unmarshal(mediumFixture, &exp)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, exp, data)
}

func TestDecodeNikandjsonV2SkipStructMedium(t *testing.T) {
	it := nikv2.Wrap(mediumFixture)
	it.Skip()
	assert.NoError(t, it.Err())
	assert.Equal(t, nikv2.None, it.Type())
}

func TestDecodeNikandjsonV2GetStructMedium(t *testing.T) {
	v := nikv2.Wrap(mediumFixture)
	v.Get("person", "gravatar", "avatars", 0, "url")
	assert.NoError(t, v.Err())
	assert.Equal(t, nikv2.String, v.Type(), "%T %T %v %v", nikv2.String, v.Type(), nikv2.String, v.Type())
	assert.Equal(t, []byte("http://1.gravatar.com/avatar/f7c8edd577d13b8930d5522f28123510"), v.NextString())
}

func BenchmarkDecodeNikandjsonV2StructMedium(b *testing.B) {
	b.ReportAllocs()
	var data MediumPayload
	for i := 0; i < b.N; i++ {
		nikv2.Unmarshal(mediumFixture, &data)
	}
}

func BenchmarkDecodeNikandjsonV2SkipStructMedium(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		it := nikv2.Wrap(mediumFixture)
		it.Skip()
	}
}

func BenchmarkDecodeNikandjsonV2GetStructMedium(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := nikv2.Wrap(mediumFixture)
		v.Get("person", "gravatar", "avatars", 0, "url")
	}
}
