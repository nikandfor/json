package json

import (
	"bytes"
)

type Value struct {
	buf    []byte
	i, end int
	parsed bool
}

func Wrap(b []byte) *Value {
	return &Value{buf: b, end: len(b)}
}

func WrapString(s string) *Value {
	return Wrap([]byte(s))
}

func (v *Value) Get(ks ...interface{}) (*Value, error) {
	if len(ks) == 0 {
		return v, nil
	}
	var err error
	var i, end int = v.i, v.end
	b := v.buf[i:end]
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

	return &Value{buf: b, i: i, end: end, parsed: true}, nil
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
		return i, len(b), NewError(b, i, ErrUnexpectedChar)
	}
	i++
	var err error
	var start int
	for j := 0; ; j++ {
		i = skipSpaces(b, i)
		if i == len(b) {
			return i, len(b), NewError(b, i, ErrUnexpectedEnd)
		}
		if b[i] == ']' {
			return i, len(b), NewError(b, i, ErrOutOfRange)
		}
		if j != 0 {
			if b[i] != ',' {
				return i, len(b), NewError(b, i, ErrUnexpectedChar)
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
		return i, len(b), NewError(b, i, ErrUnexpectedChar)
	}
	i++
	var err error
	var eq bool
	var start int
	first := true
	for {
		i = skipSpaces(b, i)
		if i == len(b) {
			return i, len(b), NewError(b, i, ErrUnexpectedEnd)
		}
		if b[i] == '}' {
			return i, len(b), NewError(b, i, ErrNoSuchKey)
		}
		if !first {
			if b[i] != ',' {
				return i, len(b), NewError(b, i, ErrUnexpectedChar)
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
			return i, len(b), NewError(b, i, ErrUnexpectedEnd)
		}
		if b[i] != ':' {
			return i, len(b), NewError(b, i, ErrUnexpectedChar)
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
		return false, i, NewError(b, i, ErrExpectedValue)
	}
	if b[i] != '"' {
		return false, i, NewError(b, i, ErrUnexpectedChar)
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
		return s, NewError(b, s, ErrUnexpectedChar)
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
	return s + len(b), NewError(b, s+len(b), ErrUnexpectedEnd)
}

func skipArray(b []byte, i int) (int, error) {
	if b[i] != '[' {
		return i, NewError(b, i, ErrUnexpectedChar)
	}
	i++
	var err error
	first := true
	for {
		i = skipSpaces(b, i)
		if i == len(b) {
			return i, NewError(b, i, ErrUnexpectedEnd)
		}
		if b[i] == ']' {
			return i + 1, nil
		}
		if !first {
			if b[i] != ',' {
				return i, NewError(b, i, ErrUnexpectedChar)
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
		return i, NewError(b, i, ErrUnexpectedChar)
	}
	i++
	var err error
	first := true
	for {
		i = skipSpaces(b, i)
		if i == len(b) {
			return i, NewError(b, i, ErrUnexpectedEnd)
		}
		if b[i] == '}' {
			return i + 1, nil
		}
		if !first {
			if b[i] != ',' {
				return i, NewError(b, i, ErrUnexpectedChar)
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
			return i, NewError(b, i, ErrUnexpectedEnd)
		}
		if b[i] != ':' {
			return i, NewError(b, i, ErrUnexpectedChar)
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
		return s + off, NewError(b, s+off, ErrExpectedValue)
	}
	return s + off, nil
}

func skipValue(b []byte, i int) (int, error) {
	i = skipSpaces(b, i)
	if i == len(b) {
		return i, NewError(b, i, ErrExpectedValue)
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
		return i, NewError(b, i, ErrUnexpectedChar)
	case 'f':
		if bytes.HasPrefix(b[i:], []byte("false")) {
			return i + 5, nil
		}
		return i, NewError(b, i, ErrUnexpectedChar)
	case 'n':
		if bytes.HasPrefix(b[i:], []byte("null")) {
			return i + 4, nil
		}
		return i, NewError(b, i, ErrUnexpectedChar)
	default:
		return skipNumber(b, i)
	}
}
