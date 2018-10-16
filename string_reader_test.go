package json

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringReader(t *testing.T) {
	data := `"qwert asdfg"`
	w := WrapString(data)
	r := w.StringReader()

	buf := make([]byte, 6)

	n, err := r.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(buf), n)
	assert.Equal(t, []byte("qwert "), buf)

	n, err = r.Read(buf)
	assert.True(t, err == nil || err == io.EOF)
	assert.Equal(t, len(data)-2-len(buf), n)
	assert.Equal(t, []byte("asdfg"), buf[:n])

	n, err = r.Read(buf)
	assert.EqualError(t, err, io.EOF.Error())
	assert.Zero(t, n)

	err = r.Close()
	assert.NoError(t, err)

	tp := r.Type()
	assert.Equal(t, None, tp)
}

func TestStringReaderFromReader(t *testing.T) {
	data := `"qwert asdfg"`
	dr := strings.NewReader(data)

	w := ReadBufferSize(dr, 3)
	r := w.StringReader()

	buf := make([]byte, 5)

	n, err := r.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(buf), n)
	assert.Equal(t, []byte("qwert"), buf[:n])

	n, err = r.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(buf), n)
	assert.Equal(t, []byte(" asdf"), buf[:n])

	n, err = r.Read(buf)
	assert.True(t, err == nil || err == io.EOF)
	assert.Equal(t, len(data)-2-2*len(buf), n)
	assert.Equal(t, []byte("g"), buf[:n])

	n, err = r.Read(buf)
	assert.EqualError(t, err, io.EOF.Error())
	assert.Zero(t, n)

	err = r.Close()
	assert.NoError(t, err)

	tp := r.Type()
	assert.Equal(t, None, tp)
}

func TestStringReaderHalfRead(t *testing.T) {
	data := `"qwert asdfg" 1`
	w := WrapString(data)
	r := w.StringReader()

	buf := make([]byte, 6)

	n, err := r.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(buf), n)
	assert.Equal(t, []byte("qwert "), buf)

	err = r.Close()
	assert.NoError(t, err)

	tp := r.Type()
	assert.Equal(t, Number, tp)
}

func TestStringReaderEscape(t *testing.T) {
	data := `"qwert\tasdfg"`
	w := WrapString(data)
	r := w.StringReader()

	buf := make([]byte, 6)

	n, err := r.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(buf), n)
	assert.Equal(t, []byte("qwert\t"), buf)

	n, err = r.Read(buf)
	assert.True(t, err == nil || err == io.EOF)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte("asdfg"), buf[:n])

	n, err = r.Read(buf)
	assert.EqualError(t, err, io.EOF.Error())
	assert.Zero(t, n)

	err = r.Close()
	assert.NoError(t, err)

	tp := r.Type()
	assert.Equal(t, None, tp)
}

func TestStringReaderEscape2(t *testing.T) {
	data := `"qwert\tasdfg\n"`
	w := WrapString(data)
	r := w.StringReader()

	buf := make([]byte, 6)

	n, err := r.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(buf), n)
	assert.Equal(t, []byte("qwert\t"), buf)

	n, err = r.Read(buf)
	assert.True(t, err == nil || err == io.EOF)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte("asdfg\n"), buf[:n])

	n, err = r.Read(buf)
	assert.EqualError(t, err, io.EOF.Error())
	assert.Zero(t, n)

	err = r.Close()
	assert.NoError(t, err)

	tp := r.Type()
	assert.Equal(t, None, tp)
}
