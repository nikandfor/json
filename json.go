package json

import (
	"errors"
	"fmt"
	"io"
	"log"
	"unicode/utf8"
)

type Type byte

var (
	errInvalidChar = errors.New("invalid character")
)

const (
	None       Type = 0
	Null       Type = 'n'
	Bool       Type = 'b'
	ArrayStart Type = '['
	ArrayEnd   Type = ']'
	ObjStart   Type = '{'
	ObjEnd     Type = '}'
	ObjKey     Type = 'k'
	String     Type = 's'
	Number     Type = 'N'
)

func (t Type) String() string {
	return string(t)
}

type Reader struct {
	b           []byte
	ref, i, end int
	locked      bool
	locki       int

	d       []Type
	waitkey bool

	err error

	decoded []byte
	r       io.Reader
}

func Wrap(b []byte) *Reader {
	if false && len(b) < 300 {
		log.Printf("Wrap      : '%s'", b)
		pad := []byte("_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_")
		if len(b) <= len(pad) {
			log.Printf("____      : '%s' = %d total", pad[:len(b)], len(b))
		}
	}
	return &Reader{b: b, end: len(b)}
}

func WrapString(s string) *Reader {
	return Wrap([]byte(s))
}

func ReadBufferSize(r io.Reader, s int) *Reader {
	rv := &Reader{
		b: make([]byte, s),
		r: r,
	}
	return rv
}

func Read(r io.Reader) *Reader {
	return ReadBufferSize(r, 1000)
}

func (r *Reader) Reset(b []byte) {
	r.b = b
	r.ref = 0
	r.i = 0
	r.end = len(b)
	r.locki = 0
	r.d = r.d[:0]
	r.waitkey = false
	r.err = nil
	r.r = nil
}

func (r *Reader) ResetString(s string) {
	r.Reset([]byte(s))
}

func (r *Reader) ResetReader(rd io.Reader) {
	if cap(r.b) > 0 {
		r.Reset(r.b[:cap(r.b)])
	}
	r.r = rd
}

func (r *Reader) more() bool {
	if r.r == nil {
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

func (r *Reader) lock() {
	r.locki = r.i
	r.locked = true
}

func (r *Reader) unlock() {
	r.locked = false
}

func (r *Reader) Skip() {
	//	log.Printf("Skip _: %2v + %2v '%s'", r.ref, r.i, r.b)
	var d int
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
			r.skipString()
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
				i++
				continue
				//	r.i = i
				//	r.NextNumber()
				//	i = r.i
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

func (r *Reader) NextBytes() []byte {
	r.Type()
	r.lock()
	r.Skip()
	r.unlock()
	return r.b[r.locki:r.i]
}

func (r *Reader) Get(ks ...interface{}) {
loop:
	for _, k := range ks {
		if true {
			next := r.b
			if len(next) > 10 {
				next = next[:10]
			}
			//	log.Printf("get   : '%v' from %d+%d '%s'", k, r.ref, r.i, next)
		}
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
		case string:
			key = []byte(k)
		case []byte:
			key = k
		default:
			r.err = fmt.Errorf("invalid argument type: %T", k)
			return
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
	}
}

func (r *Reader) Type() Type {
	//	log.Printf("Type  : %d+%d/%d '%s'", r.ref, r.i, r.end, r.b)
	//	defer func() {
	//		log.Printf("Type1 : %d+%d/%d '%s'", r.ref, r.i, r.end, r.b)
	//	}()
start:
	for r.i < r.end {
		c := r.b[r.i]
		switch c {
		case ' ', '\t', '\n':
			r.i++
		case '"':
			if r.waitkey {
				return ObjKey
			} else {
				return String
			}
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

func (r *Reader) NextString() []byte {
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
		case c == '"':
			i++
			r.i = i
			if len(r.decoded) == 0 {
				return r.b[s : i-1]
			}
			r.decoded = append(r.decoded, r.b[s:i-1]...)
			return r.decoded
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

func (r *Reader) CompareKey(k []byte) (r_ bool) {
	//	log.Printf("compKy: '%s' to %d+%d/%d '%s'", k, r.ref, r.i, r.end, r.b)
	//	defer func() {
	//		log.Printf("compK1: '%s' to %d+%d/%d '%s'", k, r.ref, r.i, r.end, r.b)
	//	}()
	i := r.i
	i++
	j := 0
	r_ = true
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
			return
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

func (r *Reader) NextNumber() []byte {
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

func (r *Reader) Err() error {
	if r.err == nil {
		return nil
	}
	return NewError(r.b, r.i, r.err)
}
