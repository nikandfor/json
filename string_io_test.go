package json

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
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

func ExampleReader_Base64Reader() {
	var buf bytes.Buffer

	token := make([]byte, 10)
	_, _ = rand.New(rand.NewSource(1)).Read(token)

	// Writer

	fmt.Printf("Writer sends message with a raw bytes token:   % 2x\n", token)

	// create json writer on top of io.Writer
	w := NewStreamWriter(&buf)

	// generate object
	w.ObjStart()

	w.ObjKey([]byte("user"))
	w.String([]byte("nikandfor"))

	w.ObjKey([]byte("token"))
	// this way we can handle json strings of any size with constant memory usage
	sw := w.Base64Writer(base64.RawStdEncoding)
	_, _ = sw.Write(token) // we don't care errors here, we'll care them later at once
	_ = sw.Close()

	w.ObjEnd()

	// check all errors until now
	err := w.Flush()
	if err != nil {
		// process
	}

	// Reader

	// read io.Reader stream on the fly with constant memory usage
	r := NewReader(&buf)

	var user string
	var tokenDec []byte = make([]byte, len(token))

	// check if it's really object (but we can skip this)
	if tp := r.Type(); tp != Object {
		fmt.Printf("Expected object, got %v\n", tp)
	}

	// read all key-value pairs
	for r.HasNext() {
		// key
		k := r.NextString()

		// value
		switch {
		case bytes.Equal(k, []byte("user")):
			user = string(r.NextString())
		case bytes.Equal(k, []byte("token")):
			sr := r.Base64Reader(base64.RawStdEncoding)
			_, err := io.ReadFull(sr, tokenDec)
			if err != nil {
				// will be io.EOF if there is less that len(tokenDec) bytes at string
			}
			sr.Close()
		default:
			r.Skip()
		}
	}

	// check error just before application logic
	err = r.Err()
	if err != nil {
		// process
	}

	fmt.Printf("Reader got message from user: %v, token % 2x\n", user, tokenDec)

	// Output:
	// Writer sends message with a raw bytes token:   52 fd fc 07 21 82 65 4f 16 3f
	// Reader got message from user: nikandfor, token 52 fd fc 07 21 82 65 4f 16 3f
}
