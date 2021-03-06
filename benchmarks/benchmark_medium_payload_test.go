package go_benchmark

import (
	"bytes"
	jsonstd "encoding/json"
	"testing"

	"github.com/buger/jsonparser"
	jsoniter "github.com/json-iterator/go"
	"github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	"github.com/nikandfor/json"
	"github.com/stretchr/testify/assert"
)

// raw loop
func BenchmarkRawLoopStructMedium(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(MediumFixture)))
	skipString := func(b []byte, i int) int {
		i++
		for b[i] != '"' {
			i++
		}
		i++
		return i
	}
	for i := 0; i < b.N; i++ {
		d := 0
		var c, p, q, n int
		for j := 0; j < len(MediumFixture); j++ {
			c := MediumFixture[j]
			switch c {
			case ' ', '\t', '\n':
				continue
			}
			switch c {
			case '{', '[':
				d++
			case '}', ']':
				d--
			case ':':
				c++
			case ',':
				p++
			case '"':
				q++
				j = skipString(MediumFixture, j)
			case '+', '-':
				n++
			default:
				if c >= '0' && c <= '9' {
					n++
				}
			}
		}
		_, _, _, _, _ = c, p, q, n, d
	}
}

// encode/json
func BenchmarkDecodeStdStructMedium(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(MediumFixture)))
	var err error
	var data MediumPayload
	for i := 0; i < b.N; i++ {
		err = jsonstd.Unmarshal(MediumFixture, &data)
	}
	assert.NoError(b, err)
}

func BenchmarkEncodeStdStructMedium(b *testing.B) {
	var data MediumPayload
	jsonstd.Unmarshal(MediumFixture, &data)
	b.ReportAllocs()
	b.SetBytes(int64(len(MediumFixture)))
	for i := 0; i < b.N; i++ {
		jsonstd.Marshal(data)
	}
}

// jsoniter
func BenchmarkDecodeJsoniterStructMedium(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(MediumFixture)))
	var err error
	var data MediumPayload
	for i := 0; i < b.N; i++ {
		err = jsoniter.Unmarshal(MediumFixture, &data)
	}
	assert.NoError(b, err)
}

func BenchmarkEncodeJsoniterStructMedium(b *testing.B) {
	var data MediumPayload
	jsoniter.Unmarshal(MediumFixture, &data)
	b.ReportAllocs()
	b.SetBytes(int64(len(MediumFixture)))
	for i := 0; i < b.N; i++ {
		jsoniter.Marshal(data)
	}
}

// easyjson
func BenchmarkDecodeEasyJsonMedium(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(MediumFixture)))
	var err error
	var data MediumPayload
	for i := 0; i < b.N; i++ {
		lexer := &jlexer.Lexer{Data: MediumFixture}
		data.UnmarshalEasyJSON(lexer)
		err = lexer.Error()
	}
	assert.NoError(b, err)
}

func BenchmarkEncodeEasyJsonMedium(b *testing.B) {
	var data MediumPayload
	lexer := &jlexer.Lexer{Data: MediumFixture}
	data.UnmarshalEasyJSON(lexer)
	b.ReportAllocs()
	b.SetBytes(int64(len(MediumFixture)))
	buf := &bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		writer := &jwriter.Writer{}
		data.MarshalEasyJSON(writer)
		buf.Reset()
		writer.DumpTo(buf)
	}
}

// buger
func BenchmarkDecodeBugerSearchStructMedium(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(MediumFixture)))
	for i := 0; i < b.N; i++ {
		jsonparser.Get(MediumFixture, "person", "gravatar", "avatars", "0", "url")
	}
}

// nikandfor
func TestDecodeNikandjsonSkipStructMedium(t *testing.T) {
	it := json.Wrap(MediumFixture)
	it.Skip()
	assert.NoError(t, it.Err())
	assert.Equal(t, json.None, it.Type())
}

func TestDecodeNikandjsonSearchStructMedium(t *testing.T) {
	v := json.Wrap(MediumFixture)
	v.Search("person", "gravatar", "avatars", 0, "url")
	assert.NoError(t, v.Err())
	assert.Equal(t, json.String, v.Type(), "%T %T %v %v", json.String, v.Type(), json.String, v.Type())
	assert.Equal(t, []byte("http://1.gravatar.com/avatar/f7c8edd577d13b8930d5522f28123510"), v.NextString())
}

func BenchmarkDecodeNikandjsonSkipStructMedium(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(MediumFixture)))
	var r json.Reader
	for i := 0; i < b.N; i++ {
		r.Reset(MediumFixture).Skip()
	}
	assert.NoError(b, r.Err())
}

func BenchmarkDecodeNikandjsonSearchStructMedium(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(MediumFixture)))
	//	keys := []interface{}{[]byte("person"), []byte("gravatar"), []byte("avatars"), 0, []byte("url")} // to not allocate them in the loop
	var r json.Reader
	for i := 0; i < b.N; i++ {
		r.Reset(MediumFixture)
		r.Search("person", "gravatar", "avatars", 0, "url")
		//	r.Search(keys...)
	}
}
