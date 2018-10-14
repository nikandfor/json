package json

import (
	"errors"
	"io"
	"log"
)

var ErrError = errors.New("error")

type Type byte

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

	d       []Type
	waitkey bool

	err error

	decoded []byte
	r       io.Reader
}

func Wrap(b []byte) *Reader {
	if false {
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
	return ReadBufferSize(r, 100)
}

func (r *Reader) more() bool {
	if r.r == nil {
		return false
	}
	r.ref += r.end
	r.i -= r.end
	n, err := r.r.Read(r.b)
	r.end = n
	if n != 0 && err == io.EOF {
		err = nil
	}
	if err != nil {
		r.err = err
		return false
	}
	return 0 < n
}

func (r *Reader) Skip() {
	log.Printf("Skip _: %2v + %2v '%s'", r.ref, r.i, r.b)
	var d int
start:
	for r.i < r.end {
		c := r.b[r.i]
		switch c {
		case ' ', '\t', '\n':
			r.i++
			continue
		}
		log.Printf("skip _: %2v + %2v '%c' %d", r.ref, r.i, c, d)
		switch c {
		case '"':
			r.skipString()
		case ',':
			r.i++
		case ':':
			r.i++
		case 't':
			r.skip([]byte("true"))
		case 'f':
			r.skip([]byte("false"))
		case 'n':
			r.skip([]byte("null"))
		case '{', '[':
			d++
			r.i++
		case '}', ']':
			d--
			r.i++
		case '+', '-', '.':
			r.NextNumber()
		default:
			if c >= '0' && c <= '9' {
				r.NextNumber()
			} else {
				r.err = ErrError
				return
			}
		}
		if d == 0 {
			return
		}
	}
	if r.more() {
		goto start
	}
}

func (r *Reader) Get(ks ...interface{}) {
loop:
	for _, k := range ks {
		if true {
			next := r.b
			if len(next) > 10 {
				next = next[:10]
			}
			log.Printf("get   : '%v' from %d+%d '%s'", k, r.ref, r.i, next)
		}
		var key []byte
		switch k := k.(type) {
		case int:
			j := 0
			for it := r.ArrayIter(); it.HasNext(); {
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
			r.err = ErrError
			return
		}
		for it := r.ObjectIter(); it.HasNext(); {
			ok := r.compareKey(key)
			log.Printf("compr: %v", ok)
			r.i++
			if ok {
				continue loop
			}
			r.Skip()
		}
	}
}

func (r *Reader) Type() Type {
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

			r.err = ErrError
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
	r.i++
	s := r.i
start:
	for r.i < r.end {
		c := r.b[r.i]
		r.i++
		if c == '"' {
			if len(r.decoded) == 0 {
				return r.b[s : r.i-1]
			}
			r.decoded = append(r.decoded, r.b[s:r.i-1]...)
			return r.decoded
		}
	}
	r.decoded = append(r.decoded, r.b[s:r.i]...)
	if r.more() {
		s = r.i
		goto start
	}
	return nil
}

func (r *Reader) skipString() {
	r.i++
start:
	for r.i < r.end {
		c := r.b[r.i]
		r.i++
		if c == '"' {
			return
		}
	}
	if r.more() {
		goto start
	}
}

func (r *Reader) compareKey(k []byte) (r_ bool) {
	log.Printf("compKy: '%s' to %d+%d/%d '%s'", k, r.ref, r.i, r.end, r.b)
	defer func() {
		log.Printf("compK1: '%s' to %d+%d/%d '%s'", k, r.ref, r.i, r.end, r.b)
	}()
	r.i++
	j := 0
	r_ = true
start:
	for r.i < r.end {
		c := r.b[r.i]
		log.Printf("compK_: '%s' to %d+%d/%d '%c'", k, r.ref, r.i, r.end, c)
		r.i++
		if c == '"' {
			if j < len(k) {
				r_ = false
			}
			return
		}
		if r_ {
			if j == len(k) || c != k[j] {
				r_ = false
			}
			j++
		}
	}
	if r.more() {
		goto start
	}
	return false
}

func (r *Reader) skip(k []byte) {
	j := 0
start:
	for r.i < r.end {
		if j == len(k) {
			return
		}
		c := r.b[r.i]
		if c != k[j] {
			r.err = ErrError
			return
		}
		j++
		r.i++
	}
	if r.more() {
		goto start
	}
}

func (r *Reader) NextNumber() []byte {
	r.decoded = r.decoded[:0]
	s := r.i
start:
	for r.i < r.end {
		c := r.b[r.i]
		if c >= '0' && c <= '9' {
			r.i++
			continue
		}
		switch c {
		case '+', '-', '.', 'e':
			r.i++
			continue
		}
		// ret
		if len(r.decoded) == 0 {
			return r.b[s:r.i]
		}
		r.decoded = append(r.decoded, r.b[s:r.i]...)
		return r.decoded
	}
	r.decoded = append(r.decoded, r.b[s:r.i]...)
	if r.more() {
		s = r.i
		goto start
	}

	return nil
}

func (r *Reader) Err() error {
	if r.err == nil {
		return nil
	}
	return r.err
}
