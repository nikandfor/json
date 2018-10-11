package json

import (
	"bytes"
	"errors"
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
	for _, k := range ks {
		switch k := k.(type) {
		case int:
			b, err = getFromArray(b, k)
		case string:
			req := []byte(k)
			b, err = getFromObject(b, req)
		case []byte:
			b, err = getFromObject(b, k)
		default:
			panic(k)
		}
		if err != nil {
			return nil, err
		}
	}

	return &Value{buf: b, parsed: true}, nil
}

func (v *Value) MustGet(ks ...interface{}) *Value {
	v, err := v.Get(ks...)
	if err != nil {
		panic(err)
	}
	return v
}

func getFromArray(b []byte, k int) ([]byte, error) {
	if b[0] != '[' {
		return nil, ErrUnexpectedChar
	}
	i := 1
	for j := 0; ; j++ {
		if i == len(b) {
			return nil, ErrUnexpectedEnd
		}
		if b[i] == ']' {
			return nil, ErrOutOfRange
		}
		if i != 1 {
			if b[i] != ',' {
				return nil, ErrUnexpectedChar
			}
			i++
		}
		si, err := skipValue(b[i:])
		i += si
		if err != nil {
			return nil, err
		}
		if j == k {
			b = b[i-si : i]
			break
		}
	}
	return b, nil
}

func getFromObject(b []byte, k []byte) ([]byte, error) {
	if b[0] != '{' {
		return nil, ErrUnexpectedChar
	}
	i := 1
	for {
		if i == len(b) {
			return nil, ErrUnexpectedEnd
		}
		if b[i] == '}' {
			return nil, ErrNoSuchKey
		}
		if i != 1 {
			if b[i] != ',' {
				return nil, ErrUnexpectedChar
			}
			i++
		}
		eq, si, err := checkKey(b[i:], k)
		i += si
		if err != nil {
			return nil, err
		}
		if i == len(b) {
			return nil, ErrUnexpectedEnd
		}
		if b[i] != ':' {
			return nil, ErrUnexpectedChar
		}
		i++
		si, err = skipValue(b[i:])
		i += si
		if err != nil {
			return nil, err
		}
		if eq {
			b = b[i-si : i]
			break
		}
	}
	return b, nil
}

func checkKey(b, k []byte) (bool, int, error) {
	if len(b) == 0 {
		return false, 0, ErrExpectedValue
	}
	if b[0] != '"' {
		return false, 0, ErrUnexpectedChar
	}
	esc := false
	var off, j int
	var eq bool = len(b) > 1
	for i, c := range b[1:] {
		if c == '\\' && !esc {
			esc = true
			continue
		}
		if c == '"' && !esc {
			off = 2 + i
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
	return eq, off, nil
}

func skipString(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, ErrExpectedValue
	}
	if b[0] != '"' {
		return 0, ErrUnexpectedChar
	}
	esc := false
	for i, c := range b[1:] {
		if c == '\\' && !esc {
			esc = true
			continue
		}
		if c == '"' && !esc {
			return i + 2, nil
		}
		esc = false
	}
	return len(b), ErrUnexpectedEnd
}

func skipArray(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, ErrExpectedValue
	}
	if b[0] != '[' {
		return 0, ErrUnexpectedChar
	}
	i := 1
	for {
		if i == len(b) {
			return i, ErrUnexpectedEnd
		}
		if b[i] == ']' {
			return i + 1, nil
		}
		if i != 1 {
			if b[i] != ',' {
				return i, ErrUnexpectedChar
			}
			i++
		}
		off, err := skipValue(b[i:])
		i += off
		if err != nil {
			return i, err
		}
	}
}

func skipObject(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, ErrExpectedValue
	}
	if b[0] != '{' {
		return 0, ErrUnexpectedChar
	}
	i := 1
	for {
		if i == len(b) {
			return 0, ErrUnexpectedEnd
		}
		if b[i] == '}' {
			return i + 1, nil
		}
		if i != 1 {
			if b[i] != ',' {
				return i, ErrUnexpectedChar
			}
			i++
		}
		off, err := skipString(b[i:])
		i += off
		if err != nil {
			return i, err
		}
		if i == len(b) {
			return 0, ErrUnexpectedEnd
		}
		if b[i] != ':' {
			return i, ErrUnexpectedChar
		}
		i++
		off, err = skipValue(b[i:])
		i += off
		if err != nil {
			return i, err
		}
	}
}

func skipNumber(b []byte) (int, error) {
	var off int = len(b)
	expSign := true
	expPoint := true
	expE := false
	hadE := false
	expNum := true
	for i, c := range b {
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
		return off, ErrExpectedValue
	}
	return off, nil
}

func skipValue(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, ErrExpectedValue
	}
	switch b[0] {
	case '[':
		return skipArray(b)
	case '{':
		return skipObject(b)
	case '"':
		return skipString(b)
	case 't':
		if bytes.HasPrefix(b, []byte("true")) {
			return 4, nil
		}
		return 0, ErrUnexpectedChar
	case 'f':
		if bytes.HasPrefix(b, []byte("false")) {
			return 5, nil
		}
		return 0, ErrUnexpectedChar
	case 'n':
		if bytes.HasPrefix(b, []byte("null")) {
			return 4, nil
		}
		return 0, ErrUnexpectedChar
	default:
		return skipNumber(b)
	}
}

func (v *Value) Buffer() []byte {
	return v.buf
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
		_, err := skipValue(b)
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
