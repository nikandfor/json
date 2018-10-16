package json

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringWriter(t *testing.T) {
	w := NewWriter(nil)
	sw := w.StringWriter()

	n, err := sw.Write([]byte("one"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	n, err = sw.Write([]byte(" "))
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
	n, err = sw.Write([]byte("two"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)

	err = sw.Close()
	assert.NoError(t, err)

	assert.Equal(t, []byte(`"one two"`), w.Bytes())
}

func TestBase64Key(t *testing.T) {
	var buf bytes.Buffer

	token := make([]byte, 1024)
	_, _ = rand.Read(token)
	//	for i := range token {
	//		token[i] = (byte)(i % 10)
	//	}

	w := NewStreamWriterBuffer(&buf, make([]byte, 10))

	w.ObjStart()

	w.ObjKey([]byte("user"))
	w.String([]byte("nikandfor"))

	w.ObjKey([]byte("token"))
	sw := w.Base64Writer(base64.RawStdEncoding)
	//	sw := w.StringWriter()
	_, _ = sw.Write(token)
	sw.Close()

	w.ObjEnd()

	err := w.Flush()
	assert.NoError(t, err)

	//	t.Logf("msg: %s", buf.Bytes())

	// reader
	r := NewReader(&buf)

	assert.NoError(t, r.Err())

	var user string
	var tokenDec []byte = make([]byte, len(token))
	for r.HasNext() {
		k := r.NextString()
		//	t.Logf("iter _: %2v + %2v/%2v '%s' %v  %s %v", r.ref, r.i, r.end, r.b, r.err, k, r.Type())
		switch {
		case bytes.Equal(k, []byte("user")):
			user = string(r.NextString())
		case bytes.Equal(k, []byte("token")):
			/*
				r.Lock()
				tk := r.NextString()
				t.Logf("tk: %s", tk)
				r.Return()
			*/

			sr := r.Base64Reader(base64.RawStdEncoding)
			//	sr := r.StringReader()
			n, err := io.ReadFull(sr, tokenDec)
			assert.True(t, err == nil || Cause(err) == io.EOF)
			assert.Equal(t, len(tokenDec), n)
			sr.Close()
		default:
			r.Skip()
		}
	}

	err = r.Err()
	assert.True(t, err == nil || Cause(err) == io.EOF)

	assert.Equal(t, "nikandfor", user)
	assert.Equal(t, token, tokenDec, "token")
}
