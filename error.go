package json

import (
	"errors"
	"fmt"
)

var (
	ErrUnexpectedChar = errors.New("unexpected char")
	ErrOverflow       = errors.New("type overflow")
	ErrOutOfRange     = errors.New("out of range")
	ErrExpectedValue  = errors.New("expected value")
	ErrUnexpectedEnd  = errors.New("unexpected end")
	ErrNoSuchKey      = errors.New("no such key")
	ErrConversion     = errors.New("type conversion")
)

var pad []byte

func init() {
	pad = make([]byte, 100)
	for i := range pad {
		pad[i] = '_'
	}
}

type Error struct {
	b   []byte
	p   int
	err error
}

func NewError(b []byte, s int, e error) Error {
	return Error{b: b, p: s, err: e}
}

func (e Error) Error() string {
	return e.err.Error()
}

func (e Error) Pos() int {
	return e.p
}

func (e Error) Format(s fmt.State, c rune) {
	if !s.Flag('+') && !s.Flag('#') {
		fmt.Fprintf(s, "%v", e.err.Error())
		return
	}
	fmt.Fprintf(s, "parse error at pos %d: %v", e.p, e.err.Error())
	if !s.Flag('+') {
		return
	}
	w := len(pad) / 2
	b := e.b
	p := e.p

	if p > w {
		d := p - w
		p = w
		b = b[d:]
		copy(b, []byte("..."))
	}
	if len(b)-p-1 > w {
		d := len(b) - p - 1 - w
		b = b[:len(b)-d]
		copy(b[len(b)-3:], []byte("..."))
	}

	//	nn := bytes.Count(b[:p], []byte{'\n'})
	//	nt := bytes.Count(b[:p], []byte{'\t'})
	//	b = bytes.Replace(b, []byte{'\n'}, []byte{'\\', 'n'}, -1)
	//	b = bytes.Replace(b, []byte{'\t'}, []byte{'\\', 't'}, -1)
	//	p += nn + nt
	b, ss := escapeString(b, p)
	p += ss

	one := 1
	if p == len(b) {
		one = 0
	}

	fmt.Fprintf(s, "\n%s\n", b)
	//	fmt.Fprintf(s, "%d ^ %d = %d [%d]\n", p, len(b)-p-one, len(b), len(pad))
	fmt.Fprintf(s, "%s%c%s\n", pad[:p], '^', pad[:len(b)-p-one])
}

func escapeString(b []byte, p int) ([]byte, int) {
	var r int
	for i, c := range b {
		if c >= 0x20 && c < 0x80 {
			continue
		}
		r = i
		break
	}
	if r == 0 {
		return b, 0
	}
	res := make([]byte, r+(len(b)-r)*2)
	w := copy(res, b[:r])
	ss := 0
	for _, c := range b[r:] {
		if c >= 0x20 && c < 0x80 {
			res[w] = c
			w++
			continue
		}
		switch c {
		case '\t':
			res[w] = '\\'
			res[w+1] = 't'
			w += 2
		case '\n':
			res[w] = '\\'
			res[w+1] = 'n'
			w += 2
		case '\r':
			res[w] = '\\'
			res[w+1] = 'r'
			w += 2
		case '\b':
			res[w] = '\\'
			res[w+1] = 'b'
			w += 2
		case '\f':
			res[w] = '\\'
			res[w+1] = 'f'
			w += 2
		default:
			res[w] = '\\'
			res[w+1] = '*'
			w += 2
		}
		if w-ss < p {
			ss++
		}
	}
	res = res[:w]

	return res, ss
}
