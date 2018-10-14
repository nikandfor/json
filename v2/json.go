package json

import (
	"errors"
	"io"
	"log"

	"github.com/nikandfor/json"
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

var (
	ErrUnexpectedSymbol = errors.New("unexpected symbol")
	ErrConversion       = errors.New("conversion")
	ErrExpectedValue    = errors.New("expected value")
	ErrOverflow         = errors.New("type overflow")
)

type (
	Iter struct {
		b   []byte
		ref int
		i   int
		end int

		lasttp Type
		err    error

		more func()

		pos     []Type
		poslast Type
		waitkey bool

		decoded []byte
	}

	Type byte
)

func Wrap(b []byte) *Iter {
	if false {
		log.Printf("Wrap      : '%s'", b)
		pad := []byte("_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_")
		if len(b) <= len(pad) {
			log.Printf("____      : '%s' = %d total", pad[:len(b)], len(b))
		}
	}
	return &Iter{b: b, end: len(b)}
}

func WrapString(s string) *Iter {
	return Wrap([]byte(s))
}

func Iterate(r io.Reader) *Iter {
	it := &Iter{}
	return it
}

func (it *Iter) Depth() int {
	return len(it.pos)
}

func (it *Iter) decodeString() {
	// "
	it.i++
	s := it.i
	for ; it.i < it.end; it.i++ {
		//	log.Printf("decodeStri: %d '%c'", it.i, it.b[it.i])
		if it.b[it.i] == '"' {
			it.decoded = it.b[s:it.i]
			it.i++
			return
		}
	}
	return
}

func (it *Iter) skipString() {
	it.i++
	for ; it.i < it.end; it.i++ {
		if it.b[it.i] == '"' {
			break
		}
	}
	it.i++
}

func (it *Iter) CompareKey(k []byte) (r_ bool) {
	it.i++
	j := 0
	r_ = true
	for ; it.i < it.end; it.i++ {
		if it.b[it.i] == '"' {
			break
		}
		if r_ {
			if j == len(k) || k[j] != it.b[it.i] {
				r_ = false
			}
			j++
		}
	}
	if j < len(k) {
		r_ = false
	}
	it.i++
	return
}

func (it *Iter) Skip() {
	//	defer func() {
	//		log.Printf("Skip      : %v %v %v", it.i, it.lasttp, it.err)
	//	}()

	d := 0
	done := false
loop:
	for it.i < it.end {
		if it.err != nil {
			return
		}
		c := it.b[it.i]
		//	log.Printf("skip      : %v '%c' %v", it.i, c, it.err)
		switch c {
		case ' ', '\t', '\n':
			it.i++
		case '"':
			it.skipString()
			done = true
		case ',':
			it.i++
		case ':':
			it.i++
		case 't':
			it.cmp([]byte("true"))
			done = true
		case 'f':
			it.cmp([]byte("false"))
			done = true
		case 'n':
			it.cmp([]byte("null"))
			done = true
		case '{', '[':
			done = true
			d++
			it.i++
		case '}', ']':
			d--
			it.i++
		case '+', '-', '.':
			it.NextNumber()
			done = true
		default:
			if c >= '0' && c <= '9' {
				it.NextNumber()
				done = true
			} else {
				it.err = ErrUnexpectedSymbol
				return
			}
		}
		//	if it.i < it.end {
		//		c = it.b[it.i]
		//	} else {
		//		c = 0
		//	}
		//	log.Printf("skip1     : %v '%c' %v", it.i, c, it.err)
		if done && d == 0 {
			break loop
		}
	}
}

func (it *Iter) Get(ks ...interface{}) {
	//	defer func() {
	//		log.Printf("Get       : %v %v %v", it.i, it.lasttp, it.err)
	//	}()
	if len(ks) == 0 {
		return
	}

	d := 0
	done := false
	it.waitkey = it.lasttp == '{'
	j := 0
	k := ks[0]
	var ok bool
loop:
	for it.i < it.end {
		if it.err != nil {
			return
		}
		c := it.b[it.i]
		switch c {
		case ' ', '\t', '\n':
			it.i++
			continue
		}
		//	log.Printf("get       : %v '%c' wk %v pos %v %v\t\tj %v  ok %v ks %v", it.i, c, it.waitkey, it.pos, it.poslast, j, ok, ks)

		if it.poslast == '[' && j == k.(int) {
			//	log.Printf("found %v elem", k)
			if len(ks) == 1 {
				break
			}
			ks = ks[1:]
			k = ks[0]
			j = 0
		}

		switch c {
		case '"':
			if it.waitkey {
				//	log.Printf("compare key %q with '%s'", k, it.b[it.i:])
				switch k := k.(type) {
				case string:
					ok = it.CompareKey([]byte(k))
				case []byte:
					ok = it.CompareKey(k)
				default:
					it.err = ErrConversion
					break
				}
			} else {
				it.skipString()
			}
			done = true
		case ',':
			j++
			it.waitkey = it.poslast == '{'
			it.i++
		case ':':
			it.waitkey = false
			it.i++
			if ok {
				//	log.Printf("shift key")
				if len(ks) == 1 {
					break loop
				}
				ok = false
				ks = ks[1:]
				k = ks[0]
				j = 0
			} else {
				it.Skip()
			}
		case 't':
			it.cmp([]byte("true"))
			done = true
		case 'f':
			it.cmp([]byte("false"))
			done = true
		case 'n':
			it.cmp([]byte("null"))
			done = true
		case '{', '[':
			it.pos = append(it.pos, it.poslast)
			it.poslast = Type(c)
			it.waitkey = it.poslast == '{'
			done = true
			d++
			it.i++
		case '}', ']':
			if d := len(it.pos) - 1; d < 0 || it.poslast != Type(c-2) {
				it.err = ErrUnexpectedSymbol
				break
			} else {
				it.poslast = it.pos[d]
				it.pos = it.pos[:d]
			}
			d--
			it.i++
		case '+', '-', '.':
			it.NextNumber()
			done = true
		default:
			if c >= '0' && c <= '9' {
				it.NextNumber()
				done = true
			} else {
				it.err = ErrUnexpectedSymbol
				return
			}
		}
		if false {
			if it.i < it.end {
				c = it.b[it.i]
			} else {
				c = 0
			}
			//	log.Printf("get1      : %v '%c' %v  %v %v", it.i, c, it.err, it.pos, it.poslast)
		}
		if done && d == 0 {
			break loop
		}
	}
}

func (it *Iter) BytesRead() int {
	return it.ref + it.i
}

func (it *Iter) TypeNext() Type {
	//	log.Printf("TypeNext  : %v %v %v", it.i, it.lasttp, it.err)
	//	defer func() {
	//		log.Printf("TypeNext1 : %v %v %v", it.i, it.lasttp, it.err)
	//	}()

start:
	for ; it.i < it.end; it.i++ {
		c := it.b[it.i]
		switch c {
		case ' ', '\t', '\n', '\r':
			continue
		default:
			break start
		}
	}

	if it.i == it.end {
		// read more
		it.lasttp = None
		return it.lasttp
	}

	c := it.b[it.i]

	switch c {
	case '"':
		if it.waitkey {
			it.lasttp = ObjKey
		} else {
			it.lasttp = String
		}
	case ',':
		it.waitkey = it.poslast == '{'
		it.i++
		goto start
	case ':':
		it.waitkey = false
		it.i++
		goto start
	case '{', '[':
		it.pos = append(it.pos, it.poslast)
		it.poslast = Type(c)
		it.lasttp = Type(it.b[it.i])
		it.waitkey = it.poslast == '{'
		it.i++
	case '}', ']':
		if d := len(it.pos) - 1; d < 0 || it.poslast != Type(c-2) {
			it.err = ErrUnexpectedSymbol
			break
		} else {
			it.poslast = it.pos[d]
			it.pos = it.pos[:d]
		}
		it.lasttp = Type(it.b[it.i])
		it.i++
	case 't', 'f':
		it.lasttp = Bool
	case 'n':
		it.lasttp = Null
	case '+', '-', '.':
		it.lasttp = Number
	default:
		if c >= '0' && c <= '9' {
			it.lasttp = Number
		} else {
			it.err = ErrUnexpectedSymbol
		}
	}

	return it.lasttp
}

func (it *Iter) Type() Type {
	return it.lasttp
}

func (it *Iter) NextString() (r_ []byte) {
	//	defer func() {
	//		log.Printf("NextString: %v %v %v -> '%s'", it.i, it.lasttp, it.err, r_)
	//	}()
	c := it.b[it.i]
	if c != '"' {
		it.err = ErrExpectedValue
		return nil
	}
	it.decodeString()
	return it.decoded
}

func (it *Iter) NextNumber() (r_ []byte) {
	//	defer func() {
	//		log.Printf("NextNumber: %v %v %v -> '%s'", it.i, it.lasttp, it.err, r_)
	//	}()
	s := it.i
	for ; it.i < it.end; it.i++ {
		c := it.b[it.i]
		switch c {
		case '+', '-', '.', 'e', 'E':
			continue
		default:
			if c >= '0' && c <= '9' {
				continue
			}
		}
		return it.b[s:it.i]
	}
	panic("need read more")
}

func (it *Iter) NextBool() (r_ bool) {
	//	defer func() {
	//		log.Printf("NextBool: %v %v %v -> %v", it.i, it.lasttp, it.err, r_)
	//	}()

	if it.i == it.end {
		it.err = ErrExpectedValue
		return
	}

	if it.b[it.i] == 't' {
		it.cmp([]byte("true"))
		return true
	} else {
		it.cmp([]byte("false"))
		return false
	}
}

func (it *Iter) NextNull() {
	//	defer func() {
	//		log.Printf("NextNull: %v %v %v", it.i, it.lasttp, it.err)
	//	}()
	it.cmp([]byte("null"))
}

func (it *Iter) cmp(v []byte) {
	j := 0
	for ; it.i < it.end && j < len(v); it.i++ {
		c := it.b[it.i]
		if v[j] == c {
			j++
		} else {
			it.err = ErrUnexpectedSymbol
			break
		}
	}
}

func (it *Iter) Err() error {
	if it.err == nil {
		return nil
	}
	return json.NewError(it.b, it.i, it.err)
}
