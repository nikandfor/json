package json

import (
	"bytes"
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

	nn := bytes.Count(b[:p], []byte{'\n'})
	nt := bytes.Count(b[:p], []byte{'\t'})
	b = bytes.Replace(b, []byte{'\n'}, []byte{'\\', 'n'}, -1)
	b = bytes.Replace(b, []byte{'\t'}, []byte{'\\', 't'}, -1)
	p += nn + nt

	fmt.Fprintf(s, "\n%s\n", b)
	//	fmt.Fprintf(s, "%d ^ %d = %d [%d]\n", p, len(b)-p-1, len(b), len(pad))
	fmt.Fprintf(s, "%s%c%s\n", pad[:p], '^', pad[:len(b)-p-1])
}
