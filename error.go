package json

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

var (
	ErrIncompatibleTypes = errors.New("incompatible types")
	ErrEncoding          = errors.New("string encoding")
)

var pad = []byte("__________")
var notPrintableChar byte = '.'

// Error keeps position and context of an error and can pretty print it
// It supports 3 forms of formatting:
//		%v - produces output:
//		parse error at pos 15: expected object key
//
//		%#v
//		parse error at pos 15: expected object key `{"some":"text",123}`
//
//		%+v
//		parse error at pos 15: expected object key
//		{"some":"text",123}
//		_______________^___
// It also supports width setting which changes max size of context shown in extended forms
type Error struct {
	b      []byte
	ref, i int
	err    error
}

// NewError creates new error
func NewError(b []byte, ref, i int, e error) Error {
	return Error{b: b, ref: ref, i: i, err: e}
}

func (e Error) Error() string {
	return e.err.Error()
}

// Pos returns stream position at which error has happened
func (e Error) Pos() int {
	return e.ref
}

func (e Error) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "parse error at pos %d: %v", e.ref, e.err.Error())
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
	p := e.i

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

// Cause returns cause error
func (e Error) Cause() error {
	return e.err
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
			if n == utf8.RuneError && s == 1 {
				res[w] = notPrintableChar
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

// Cause returns cause of underlying error if any
//
// copy of github.com/pkg/errors.Cause
func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}
