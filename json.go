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
	return ReadBufferSize(r, 1000)
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
			r.skip([]byte("true"))
			i = r.i
		case 'f':
			r.i = i
			r.skip([]byte("false"))
			i = r.i
		case 'n':
			r.i = i
			r.skip([]byte("null"))
			i = r.i
		case '{', '[':
			d++
			i++
		case '}', ']':
			d--
			i++
		default:
			if c >= '0' && c <= '9' || c == '.' || c == '-' || c == '+' {
				r.i = i
				r.NextNumber()
				i = r.i
			} else {
				r.i = i
				r.err = ErrError
				return
			}
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
	i := r.i
	r.Skip()
	return r.b[i:r.i]
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
			r.err = ErrError
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
	i := r.i
	i++
	s := i
start:
	for i < r.end {
		c := r.b[i]
		i++
		if c == '"' {
			r.i = i
			if len(r.decoded) == 0 {
				return r.b[s : i-1]
			}
			r.decoded = append(r.decoded, r.b[s:i-1]...)
			return r.decoded
		}
	}
	r.i = i
	r.decoded = append(r.decoded, r.b[s:i]...)
	if r.more() {
		s = r.i
		i = r.i
		goto start
	}
	return nil
}

func (r *Reader) skipString() {
	i := r.i
	i++
	esc := false
start:
	for i < r.end {
		c := r.b[i]
		i++
		switch c {
		case '\\':
			if esc {
				esc = false
				continue
			}
			esc = true
		case '"':
			if esc {
				esc = false
				continue
			}
			r.i = i
			return
		}
	}
	r.i = i
	if r.more() {
		i = r.i
		goto start
	}
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

func (r *Reader) skip(k []byte) {
	j := 0
start:
	i := r.i
	for i < r.end {
		if j == len(k) {
			r.i = i
			return
		}
		c := r.b[i]
		if c != k[j] {
			r.i = i
			r.err = ErrError
			return
		}
		j++
		i++
	}
	r.i = i
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

	return r.decoded
}

func (r *Reader) Err() error {
	if r.err == nil {
		return nil
	}
	return r.err
}
