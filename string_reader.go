package json

import (
	"encoding/base64"
	"fmt"
	"io"
)

// StringReader allows to read string value as an io.ReadCloser interface
type StringReader struct {
	*Reader
}

// StringReader allows to read string value as an io.ReadCloser interface
func (r *Reader) StringReader() StringReader {
	r.Type()

	if r.b[r.i] != '"' {
		r.err = fmt.Errorf("expected string")
	}
	r.i++

	return StringReader{r}
}

// Read reads data from a big string value
// It returns io.EOF if entire string has been read
func (r StringReader) Read(p []byte) (int, error) {
	read := 0
start:
	i := r.i
	s := i
loop:
	for i < r.end {
		c := r.b[i]
		//	log.Printf("skip str0 %d+%d/%d '%c'  p %d+%d %d", r.ref, r.i, r.end, c, read, len(p), s)
		switch {
		case c == '\\':
			if i == r.end {
				break loop
			}

			n := copy(p, r.b[s:i])
			read += n

			if n == len(p) {
				return read, nil
			}

			i++
			switch r.b[i] {
			case 't':
				p[n] = '\t'
			case 'n':
				p[n] = '\n'
			case 'r':
				p[n] = '\r'
			}
			p = p[n+1:]
			read++
			s = i + 1

		//	log.Printf("read escape: %d/%d '%c' p %d+%d", i, r.end, r.b[i], read, len(p))
		case c == '"':
			r.i = i
			read += copy(p, r.b[s:i])
			return read, io.EOF
		case i-s == len(p):
			r.i = i
			read += copy(p, r.b[s:i])
			return read, nil
		}
		i++
	}
	n := copy(p, r.b[s:i])
	read += n
	p = p[n:]
	r.i = i
	if r.more() {
		goto start
	}
	if r.err == nil {
		r.err = io.ErrUnexpectedEOF
	}
	return read, r.Err()
}

// Close must be called to finish reading current string value
// It will skip unread part of string if any
func (r StringReader) Close() error {
	if r.err != nil {
		return r.Err()
	}
	if r.i < r.end && r.b[r.i] == '"' {
		r.i++
		return nil
	}
	r.skipString(false)
	return r.Err()
}

// Base64Reader allows to read raw bytes stream encoded in base64 encoding
type Base64Reader struct {
	StringReader
	d io.Reader
}

// Base64Reader allows to read raw bytes stream encoded in base64 encoding
func (r *Reader) Base64Reader(enc *base64.Encoding) Base64Reader {
	s := Base64Reader{StringReader: r.StringReader()}
	s.d = base64.NewDecoder(enc, s.StringReader)
	return s
}

// Read reads data from a big string value and decodes it from base64 encoding
// It returns io.EOF if entire string has been read
func (r Base64Reader) Read(p []byte) (int, error) {
	return r.d.Read(p)
}
