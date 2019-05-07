package json

import (
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

// Type is a token type
type Type byte

var (
	errInvalidChar = errors.New("invalid character")
)

const (
	None   Type = 0
	Null   Type = 'n'
	Bool   Type = 'b'
	Array  Type = '['
	Object Type = '{'
	String Type = 's'
	Number Type = 'N'
)

// Reader parses byte stream and allows you to manipulate structured data
type Reader struct {
	b           []byte
	ref, i, end int
	locki       int
	locked      bool

	nozero bool

	//	d []Type

	err error

	decoded []byte
	r       io.Reader
}

// Wrap wraps single byte buffer info json reader object
func Wrap(b []byte) *Reader {
	//	if false && len(b) < 300 {
	//		log.Printf("Wrap      : '%s'", b)
	//		pad := []byte("_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_")
	//		if len(b) <= len(pad) {
	//			log.Printf("____      : '%s' = %d total", pad[:len(b)], len(b))
	//		}
	//	}
	return &Reader{b: b, end: len(b)}
}

// WrapString wraps string
func WrapString(s string) *Reader {
	return Wrap([]byte(s))
}

// NewReaderBufferSize creates reader from stream with specified buffer size
func NewReaderBufferSize(r io.Reader, s int) *Reader {
	rv := &Reader{
		b: make([]byte, s),
		r: r,
	}
	return rv
}

// NewReader reads stream
func NewReader(r io.Reader) *Reader {
	return NewReaderBufferSize(r, 1000)
}

// Reset resets the reader
func (r *Reader) Reset(b []byte) *Reader {
	r.b = b
	r.ref = 0
	r.i = 0
	r.end = len(b)
	r.locki = 0
	r.err = nil
	r.r = nil
	return r
}

// ResetString resets the reader
func (r *Reader) ResetString(s string) *Reader {
	return r.Reset([]byte(s))
}

// ResetReader resets the reader
// it reuses buffer (if you created reader by Wrap for example)
// if you used Lock and read a big object into the memory it doesn't shrink it buffer back
func (r *Reader) ResetReader(rd io.Reader) *Reader {
	if cap(r.b) > 0 {
		r.Reset(r.b[:cap(r.b)])
	}
	r.r = rd
	return r
}

func (r *Reader) more() bool {
	if r.r == nil {
		//	r.err = io.EOF
		return false
	}
	keep := r.end - r.i
	if r.locked {
		keep = r.end - r.locki
	}
	//	log.Printf("more  : %d+%d/%d %d %d", r.ref, r.i, r.end, keep, r.locki)
	if keep > 0 {
		st := r.end - keep
		r.ref += st
		copy(r.b, r.b[st:r.end])
		r.i -= st
		r.locki -= st
	} else {
		r.ref += r.end
		r.i -= r.end
		keep = 0
	}

	if keep == len(r.b) {
		if l := len(r.b); l < cap(r.b) {
			r.b = r.b[:cap(r.b)]
		} else {
			if l <= 1024 {
				l *= 2
			} else {
				l += l / 3
			}
			c := make([]byte, l)
			copy(c, r.b)
			r.b = c
		}
	}

	n, err := r.r.Read(r.b[keep:])
	r.end = keep + n
	if n != 0 && err == io.EOF {
		err = nil
	}
	if err != nil {
		r.err = err
		return false
	}
	return 0 < n
}

// Lock locks all data since the current position and prevents them from freeing.
// So if buffer has ended it would be grown to accommodate new data
// You should use it for inspecting some small part of data and decide will you
// parse them from the beginning or just unlock reader and will go further.
//
// See (*Reader).Inspect() source code for better understanding.
func (r *Reader) Lock() {
	r.locki = r.i
	r.locked = true
}

// Unlock unlocks previously locked reader by (*Reader).Lock().
// It doesn't change reader position
func (r *Reader) Unlock() {
	r.locked = false
}

// Return returns the reading position to the previously locked by (*Reader).Lock
func (r *Reader) Return() {
	r.i = r.locki
	r.locked = false
}

// Skip reads and skips one next value.
// It doesn't metter if you call it before ':' or just at start of value
// (it actually doesn't matter if ':' or ',' were or not at all)
func (r *Reader) Skip() {
	r.GoOut(0)
}

// GoOut goes out of object or array (reads until end).
// d means number of objects you want to out of.
// It's the ideal pair for (*Reader).Search method
// (See (*Reader).Inspect source code as an example)
func (r *Reader) GoOut(d int) {
	//	log.Printf("Skip _: %2v + %2v '%s'", r.ref, r.i, r.b)
start:
	i := r.i
	for i < r.end {
		c := r.b[i]
		switch c {
		case ' ', '\t', '\n':
			i++
			continue
			//	}
		//	log.Printf("skip _: %2v + %2v '%c' %d", r.ref, r.i, c, d)
		//	switch c {
		case '"':
			r.i = i
			r.skipString(true)
			i = r.i
		case ',':
			i++
		case ':':
			i++
			continue
		case 't':
			r.i = i
			r.skip3('r', 'u', 'e')
			i = r.i
		case 'f':
			r.i = i
			r.skip4('a', 'l', 's', 'e')
			i = r.i
		case 'n':
			r.i = i
			r.skip3('u', 'l', 'l')
			//	r.skip([]byte("null"))
			i = r.i
		case '{', '[':
			d++
			i++
		case '}', ']':
			d--
			i++
		default:
			if c >= '0' && c <= '9' || c == '.' || c == '-' || c == '+' || c == 'e' || c == 'E' {
				//	i++
				//	continue
				r.i = i
				r.NextNumber()
				i = r.i
			} else {
				r.i = i
				r.err = errInvalidChar
				return
			}
		}
		if r.err != nil {
			r.i = i
			return
		}
		if d == 0 {
			r.i = i
			return
		}
	}
	r.i = i
	if r.more() {
		goto start
	}
}

// NextAsBytes reads next object of any type and returns it as a raw byte slice
// without decoding (including string quotes)
func (r *Reader) NextAsBytes() []byte {
	r.Type()
	r.Lock()
	r.Skip()
	r.Unlock()
	return r.b[r.locki:r.i]
}

// Search searches for value at the specified path.
// It supports keys of type int for arrays and string or []byte for objects
func (r *Reader) Search(keys ...interface{}) *Reader {
loop:
	for _, k := range keys {
		//	if true {
		//		next := r.b
		//		if len(next) > 10 {
		//			next = next[:10]
		//		}
		//		log.Printf("get   : '%v' from %d+%d '%s'", k, r.ref, r.i, next)
		//	}
		var key []byte
		switch k := k.(type) {
		case int:
			j := 0
			for r.HasNext() {
				if j == k {
					continue loop
				}
				r.Skip()
				j++
			}
			r.err = fmt.Errorf("index out of range: %d", j)
		case string:
			key = UnsafeStringToBytes(k)
		case []byte:
			key = k
		default:
			r.err = fmt.Errorf("invalid argument type: %T", k)
			return r
		}
		for r.HasNext() {
			ok := r.CompareKey(key)
			//	log.Printf("compr: %v", ok)
			r.i++
			if ok {
				continue loop
			}
			r.Skip()
		}
		r.err = fmt.Errorf("no such key: %s", key)
	}
	return r
}

// Type returns type of the next value.
// It skips commas ',' and colons ':' silently
// It doesn't skip or read over other tokens even if called multiple times
func (r *Reader) Type() Type {
	//	log.Printf("Type  : %d+%d/%d '%s'", r.ref, r.i, r.end, r.b)
	//	defer func() {
	//		log.Printf("Type1 : %d+%d/%d '%s'", r.ref, r.i, r.end, r.b)
	//	}()
	if r.err != nil {
		return None
	}
start:
	for r.i < r.end {
		c := r.b[r.i]
		switch c {
		case ' ', '\t', '\n':
			r.i++
		case '"':
			return String
		case ',':
			r.i++
		case ':':
			r.i++
		case 't', 'f':
			return Bool
		case 'n':
			return Null
		case '{', '[':
			return Type(c)
		case '}', ']':
			return Type(c)
		case '+', '-', '.':
			return Number
		default:
			if c >= '0' && c <= '9' {
				return Number
			}

			r.err = errInvalidChar
			return None
		}
	}
	if r.more() {
		goto start
	}
	return None
}

// NextString read next object key or string checking utf8 encoding correctness
// and decoding '\t', '\n', '\r' escape sequences
func (r *Reader) NextString() []byte {
	if r.Type() != String { // read until value start
		r.setErr(ErrIncompatibleTypes)
		return nil
	}

	r.decoded = r.decoded[:0]
	//	log.Printf("Skip stri %d+%d/%d", r.ref, r.i, r.end)
	r.i++
start:
	i := r.i
	s := i
loop:
	for i < r.end {
		c := r.b[i]
		//	log.Printf("skip str0 %d+%d/%d '%c'", r.ref, i, r.end, c)
		switch {
		case c == '"':
			i++
			r.i = i
			if len(r.decoded) == 0 {
				return r.b[s : i-1]
			}
			r.decoded = append(r.decoded, r.b[s:i-1]...)
			return r.decoded
		case c == '\\':
			r.decoded = append(r.decoded, r.b[s:i]...)
			i++
			c = r.b[i]
			switch c {
			case 'n':
				c = '\n'
			case 't':
				c = '\t'
			case 'r':
				c = '\r'
			default:
				r.err = fmt.Errorf("unsupported escape sequence: %c", c)
				return nil
			}
			r.decoded = append(r.decoded, c)
			i++
			s = i
		case c < 0x80: // utf8.RuneStart
			//	log.Printf("skip stri %d+%d/%d '%c' (%d)", r.ref, i, r.end, c, c)
			i++
			continue
		default:
			if i+utf8.UTFMax > r.end {
				if !utf8.FullRune(r.b[i:r.end]) {
					break loop
				}
			}
			n, s := utf8.DecodeRune(r.b[i:])
			if n == utf8.RuneError {
				r.err = fmt.Errorf("undecodable unicode symbol")
				return nil
			}
			//	log.Printf("skip rune %d+%d/%d '%c'", r.ref, i, r.end, n)
			i += s
		}
	}
	r.i = i
	r.decoded = append(r.decoded, r.b[s:i]...)
	if r.more() {
		goto start
	}

	return nil
}

// CompareKey compares given key weth the next string value.
// It doesn't decodes escape sequences and doesn't check for utf8 correctness
func (r *Reader) CompareKey(k []byte) bool {
	//	log.Printf("compKy: '%s' to %d+%d/%d '%s'", k, r.ref, r.i, r.end, r.b)
	//	defer func() {
	//		log.Printf("compK1: '%s' to %d+%d/%d '%s'", k, r.ref, r.i, r.end, r.b)
	//	}()
	i := r.i
	i++
	j := 0
	r_ := true
start:
	for i < r.end {
		c := r.b[i]
		//	log.Printf("compK_: '%s' to %d+%d/%d '%c'", k, r.ref, r.i, r.end, c)
		i++
		if c == '"' {
			if j < len(k) {
				r_ = false
			}
			r.i = i
			return r_
		}
		if r_ {
			if j == len(k) || c != k[j] {
				r_ = false
			}
			j++
		}
	}
	r.i = i
	if r.more() {
		i = r.i
		goto start
	}
	return false
}

func (r *Reader) skip3(a, b, c byte) {
	//	log.Printf("skip3 : %d+%d/%d '%s' %v", r.ref, r.i, r.end, r.b, r.err)
	var i int
	if r.i+3 >= r.end {
		if !r.more() {
			if r.err == nil {
				goto fail
			}
			return
		}
	}
	i = r.i
	i++
	if a != r.b[i] {
		goto fail
	}
	i++
	if b != r.b[i] {
		goto fail
	}
	i++
	if c != r.b[i] {
		goto fail
	}
	i++
	r.i = i
	return
fail:
	r.err = fmt.Errorf("broken literal")
}

func (r *Reader) skip4(a, b, c, d byte) {
	var i int
	if r.i+4 >= r.end {
		if !r.more() {
			if r.err == nil {
				goto fail
			}
			return
		}
	}
	i = r.i
	i++
	if a != r.b[i] {
		goto fail
	}
	i++
	if b != r.b[i] {
		goto fail
	}
	i++
	if c != r.b[i] {
		goto fail
	}
	i++
	if d != r.b[i] {
		goto fail
	}
	i++
	r.i = i
	return
fail:
	r.err = fmt.Errorf("broken literal")
}

// NextNumber reads next number (possible incorrect) of any size and returns it as bytes
// It doesn't check if nubmer is correct. It could have two points or '+' and '-'
// signs in the middle of string
func (r *Reader) NextNumber() []byte {
	r.Type()

	r.decoded = r.decoded[:0]
start:
	i := r.i
	s := i
	for i < r.end {
		c := r.b[i]
		if c >= '0' && c <= '9' {
			i++
			continue
		}
		switch c {
		case '+', '-', '.', 'e', 'E':
			i++
			continue
		}
		// ret
		r.i = i
		if len(r.decoded) == 0 {
			return r.b[s:i]
		}
		r.decoded = append(r.decoded, r.b[s:i]...)
		return r.decoded
	}
	r.i = i
	r.decoded = append(r.decoded, r.b[s:i]...)
	if r.more() {
		goto start
	}

	return r.decoded
}

// Err returns first happend error.
// It returns Error type that could nicely format error message for you
// if you will see it at console or with monospace font
func (r *Reader) Err() error {
	if r.err == nil {
		return nil
	}
	return NewError(r.b, r.i, r.err)
}

func (r *Reader) setErr(err error) error {
	if err == nil {
		return nil
	}

	r.err = err

	return r.Err()
}

func (r *Reader) ResetErr() {
	r.err = nil
}
