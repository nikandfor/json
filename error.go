package json

import (
	"fmt"
	"unicode/utf8"
)

var ()

var pad = []byte("__________")

type Error struct {
	b   []byte
	p   int
	err error
}

func NewError(b []byte, p int, e error) Error {
	return Error{b: b, p: p, err: e}
}

func (e Error) Error() string {
	return e.err.Error()
}

func (e Error) Pos() int {
	return e.p
}

func (e Error) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "parse error at pos %d: %v", e.p, e.err.Error())
	if !s.Flag('+') && !s.Flag('#') {
		return
	}

	w := 50
	if s.Flag('#') {
		w = 15
	}

	if d, ok := s.Width(); ok {
		w = d / 2
	}

	b := e.b
	p := e.p

	if p > w {
		d := p - w
		p = w
		b = b[d:]
		if !s.Flag('#') {
			copy(b, []byte("..."))
		}
	}
	if len(b)-p-1 > w {
		d := len(b) - p - 1 - w
		b = b[:len(b)-d]
		if !s.Flag('#') {
			copy(b[len(b)-3:], []byte("..."))
		}
	}

	b, p = escapeString(b, p)

	if s.Flag('#') {
		//	r, _ := utf8.DecodeRune(b[p:])
		//	if r == utf8.RuneError {
		//		r = '*'
		//	}
		fmt.Fprintf(s, " `%s`", b)
		return
	}

	fmt.Fprintf(s, "\n%s\n", b)
	//	fmt.Fprintf(s, "%d ^ %d = %d [%d]\n", p, len(b)-p-one, len(b), len(pad))
	for i := 0; i < p; i += len(pad) {
		w := len(pad)
		if p-i < w {
			w = p - i
		}
		fmt.Fprintf(s, "%s", pad[:w])
	}
	fmt.Fprintf(s, "%c", '^')
	for i := p + 1; i < len(b); i += len(pad) {
		w := len(pad)
		if len(b)-i < w {
			w = len(b) - i
		}
		fmt.Fprintf(s, "%s", pad[:w])
	}
	fmt.Fprintf(s, "\n")
}

func escapeString(b []byte, p int) ([]byte, int) {
	var r int = -1
	for i, c := range b {
		if c >= 0x20 && c < 0x80 {
			continue
		}
		r = i
		break
	}
	if r == -1 {
		return b, p
	}
	p0 := p
	res := make([]byte, r+(len(b)-r)*2)
	w := copy(res, b[:r])
	sym := r
	for i := r; i < len(b); i++ {
		c := b[i]
		if c >= 0x20 && c < 0x80 {
			res[w] = c
			sym++
			w++
			continue
		}
		//	log.Printf("symb at pos %d (%d) '%c'  w %d p %d", i, c, c, w, p)
		switch c {
		case '\t', '\n', '\r', '\b', '\f':
			res[w] = '\\'
			w++
			switch c {
			case '\t':
				res[w] = 't'
			case '\n':
				res[w] = 'n'
			case '\r':
				res[w] = 'r'
			case '\b':
				res[w] = 'b'
			case '\f':
				res[w] = 'f'
			}
			w++
			if sym < p0 {
				p++
			}
			sym++
		default:
			n, s := utf8.DecodeRune(b[i:])
			//	log.Printf("decode from %d '%s' -> '%c' %d %d != %d  p %d %d", i, b[i:], n, s, n, utf8.RuneError, p, p0)
			if n == utf8.RuneError {
				res[w] = '*'
				w++
				break
			}
			for j := 0; j < s; j++ {
				res[w] = b[i+j]
				w++
			}
			if sym < p0 {
				p -= s - 2
			}
			i += s - 1
			sym++
		}
	}
	res = res[:w]
	return res, p
}
