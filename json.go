package json

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
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

func NewError(s int, e error) Error {
	return Error{p: s, err: e}
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

type Value struct {
	buf    []byte
	parsed bool
	l      []int
}

func Wrap(b []byte) *Value {
	return &Value{buf: b}
}

func WrapString(s string) *Value {
	return Wrap([]byte(s))
}

func Unmarshal(data []byte, v interface{}) error {
	panic("not implemented")
}

func Marshal(v interface{}) ([]byte, error) {
	panic("not implemented")
}

func (v *Value) Get(ks ...interface{}) (*Value, error) {
	if len(ks) == 0 {
		return v, nil
	}
	var err error
	b := v.buf
	var i, end int = 0, len(b)
	for _, k := range ks {
		i = skipSpaces(b, i)
		switch k := k.(type) {
		case int:
			i, end, err = getFromArray(b[:end], k, i)
		case string:
			i, end, err = getFromObject(b[:end], []byte(k), i)
		case []byte:
			i, end, err = getFromObject(b[:end], k, i)
		default:
			panic(k)
		}
		if err != nil {
			return nil, err
		}
	}

	return &Value{buf: b[i:end], parsed: true}, nil
}

func (v *Value) MustGet(ks ...interface{}) *Value {
	v, err := v.Get(ks...)
	if err != nil {
		panic(err)
	}
	return v
}

func getFromArray(b []byte, k int, i int) (st_ int, i_ int, e_ error) {
	if b[i] != '[' {
		return i, len(b), NewError(i, ErrUnexpectedChar)
	}
	i++
	var err error
	var start int
	for j := 0; ; j++ {
		i = skipSpaces(b, i)
		if i == len(b) {
			return i, len(b), NewError(i, ErrUnexpectedEnd)
		}
		if b[i] == ']' {
			return i, len(b), NewError(i, ErrOutOfRange)
		}
		if j != 0 {
			if b[i] != ',' {
				return i, len(b), NewError(i, ErrUnexpectedChar)
			}
			i++
			i = skipSpaces(b, i)
		}
		start = i
		i, err = skipValue(b, i)
		if err != nil {
			return i, len(b), err
		}
		if j == k {
			break
		}
	}
	return start, i, nil
}

func getFromObject(b []byte, k []byte, i int) (st_ int, i_ int, e_ error) {
	if b[i] != '{' {
		return i, len(b), NewError(i, ErrUnexpectedChar)
	}
	i++
	var err error
	var eq bool
	var start int
	first := true
	for {
		i = skipSpaces(b, i)
		if i == len(b) {
			return i, len(b), NewError(i, ErrUnexpectedEnd)
		}
		if b[i] == '}' {
			return i, len(b), NewError(i, ErrNoSuchKey)
		}
		if !first {
			if b[i] != ',' {
				return i, len(b), NewError(i, ErrUnexpectedChar)
			}
			i++
			i = skipSpaces(b, i)
		} else {
			first = false
		}
		eq, i, err = checkKey(b, k, i)
		if err != nil {
			return i, len(b), err
		}
		i = skipSpaces(b, i)
		if i == len(b) {
			return i, len(b), NewError(i, ErrUnexpectedEnd)
		}
		if b[i] != ':' {
			return i, len(b), NewError(i, ErrUnexpectedChar)
		}
		i++
		i = skipSpaces(b, i)
		start = i
		i, err = skipValue(b, i)
		if err != nil {
			return i, len(b), err
		}
		if eq {
			break
		}
	}
	return start, i, nil
}

func checkKey(b, k []byte, i int) (bool, int, error) {
	if i == len(b) {
		return false, i, NewError(i, ErrExpectedValue)
	}
	if b[i] != '"' {
		return false, i, NewError(i, ErrUnexpectedChar)
	}
	i++
	esc := false
	var j int
	var eq bool = len(b) > i
	for p, c := range b[i:] {
		if c == '\\' && !esc {
			esc = true
			continue
		}
		if c == '"' && !esc {
			i += p + 1
			break
		}
		if eq {
			if j == len(k) || c != k[j] {
				eq = false
			}
			j++
		}
		esc = false
	}
	if j < len(k) {
		eq = false
	}
	return eq, i, nil
}

func skipSpaces(b []byte, s int) int {
	if s == len(b) {
		return s
	}
	for i, c := range b[s:] {
		switch c {
		case ' ', '\n', '\t', '\v', '\r':
			continue
		default:
			return s + i
		}
	}
	return s + len(b)
}

func skipString(b []byte, s int) (int, error) {
	if b[s] != '"' {
		return s, NewError(s, ErrUnexpectedChar)
	}
	esc := false
	for i, c := range b[s+1:] {
		if c == '\\' && !esc {
			esc = true
			continue
		}
		if c == '"' && !esc {
			return s + i + 2, nil
		}
		esc = false
	}
	return s + len(b), NewError(s+len(b), ErrUnexpectedEnd)
}

func skipArray(b []byte, i int) (int, error) {
	if b[i] != '[' {
		return i, NewError(i, ErrUnexpectedChar)
	}
	i++
	var err error
	first := true
	for {
		i = skipSpaces(b, i)
		if i == len(b) {
			return i, NewError(i, ErrUnexpectedEnd)
		}
		if b[i] == ']' {
			return i + 1, nil
		}
		if !first {
			if b[i] != ',' {
				return i, NewError(i, ErrUnexpectedChar)
			}
			i++
		} else {
			first = false
		}
		i, err = skipValue(b, i)
		if err != nil {
			return i, err
		}
	}
}

func skipObject(b []byte, i int) (int, error) {
	if b[i] != '{' {
		return i, NewError(i, ErrUnexpectedChar)
	}
	i++
	var err error
	first := true
	for {
		i = skipSpaces(b, i)
		if i == len(b) {
			return i, NewError(i, ErrUnexpectedEnd)
		}
		if b[i] == '}' {
			return i + 1, nil
		}
		if !first {
			if b[i] != ',' {
				return i, NewError(i, ErrUnexpectedChar)
			}
			i++
			i = skipSpaces(b, i)
		} else {
			first = false
		}
		i, err = skipString(b, i)
		if err != nil {
			return i, err
		}
		i = skipSpaces(b, i)
		if i == len(b) {
			return i, NewError(i, ErrUnexpectedEnd)
		}
		if b[i] != ':' {
			return i, NewError(i, ErrUnexpectedChar)
		}
		i++
		i, err = skipValue(b, i)
		if err != nil {
			return i, err
		}
	}
}

func skipNumber(b []byte, s int) (int, error) {
	var off int = len(b)
	expSign := true
	expPoint := true
	expE := false
	hadE := false
	expNum := true
	for i, c := range b[s:] {
		if expSign {
			if c == '+' || c == '-' {
				expSign = false
				continue
			}
		}
		if expPoint {
			if c == '.' {
				expPoint = false
				continue
			}
		}
		if expE {
			if c == 'e' || c == 'E' {
				hadE = true
				expE = false
				expSign = true
				expNum = true
				expPoint = false
				continue
			}
		}
		if c >= '0' && c <= '9' {
			if !hadE && !expE {
				expE = true
			}
			expNum = false
			continue
		}
		off = i
		break
	}
	if expNum {
		return s + off, NewError(s+off, ErrExpectedValue)
	}
	return s + off, nil
}

func skipValue(b []byte, i int) (int, error) {
	i = skipSpaces(b, i)
	if i == len(b) {
		return i, NewError(i, ErrExpectedValue)
	}
	switch b[i] {
	case '[':
		return skipArray(b, i)
	case '{':
		return skipObject(b, i)
	case '"':
		return skipString(b, i)
	case 't':
		if bytes.HasPrefix(b[i:], []byte("true")) {
			return i + 4, nil
		}
		return i, NewError(i, ErrUnexpectedChar)
	case 'f':
		if bytes.HasPrefix(b[i:], []byte("false")) {
			return i + 5, nil
		}
		return i, NewError(i, ErrUnexpectedChar)
	case 'n':
		if bytes.HasPrefix(b[i:], []byte("null")) {
			return i + 4, nil
		}
		return i, NewError(i, ErrUnexpectedChar)
	default:
		return skipNumber(b, i)
	}
}

func (v *Value) Buffer() []byte {
	return v.buf
}

func (v *Value) String() string {
	b := v.buf
	if !v.parsed {
		_, err := skipValue(b, 0)
		if err != nil {
			return string(b)
		}
		v.parsed = true
	}
	if b[0] == '"' {
		return string(b[1 : len(b)-1])
	}
	return string(b)
}

func (v *Value) Int() (int, error) {
	r := 0
	n := false
	buf := v.buf
	if buf[0] == '-' {
		n = true
		buf = buf[1:]
	} else if buf[0] == '+' {
		buf = buf[1:]
	}
	for _, c := range v.buf {
		if c < '0' || c > '9' {
			return r, ErrUnexpectedChar
		}
		r = r*10 + (int)(c-'0')
		if r < 0 {
			return r, ErrOverflow
		}
	}

	if n {
		r = -r
	}

	return r, nil
}

func (v *Value) MustInt() int {
	r, err := v.Int()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) Float64() (float64, error) {
	return strconv.ParseFloat(string(v.buf), 64)
}

func (v *Value) MustFloat64() float64 {
	r, err := v.Float64()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) Bool() (bool, error) {
	if bytes.Equal(v.buf, []byte("true")) {
		return true, nil
	}
	if bytes.Equal(v.buf, []byte("false")) {
		return false, nil
	}
	return false, ErrConversion
}

func (v *Value) MustBool() bool {
	r, err := v.Bool()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) CastBool() (bool, error) {
	b := v.buf
	if !v.parsed {
		_, err := skipValue(b, 0)
		if err != nil {
			return false, err
		}
		v.parsed = true
	}
	switch b[0] {
	case '[':
		return true, nil
	case '{':
		return true, nil
	case '"':
		return true, nil
	case 't':
		return true, nil
	case 'f':
		return false, nil
	case 'n':
		return false, nil
	default:
		return true, nil
	}
}
