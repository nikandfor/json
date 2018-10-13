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
	i := skipSpaces(b, 0)
	return &Value{buf: b, i: i, end: len(b), parsed: false}
}

func WrapString(s string) *Value {
	return Wrap([]byte(s))
}

func (v *Value) Get(ks ...interface{}) (*Value, error) {
	if len(ks) == 0 {
		return v, nil
	}
	var err error
	i := v.i
	for _, k := range ks {
		i = skipSpaces(v.buf, i)
		switch k := k.(type) {
		case int:
			i, err = getFromArray(v.buf, k, i)
		case string:
			i, err = getFromObject(v.buf, []byte(k), i)
		case []byte:
			i, err = getFromObject(v.buf, k, i)
		default:
			panic(k)
		}
		if err != nil {
			return nil, err
		}
	}

	end, err := skipValue(v.buf, i)
	if err != nil {
		return nil, err
	}

	return &Value{buf: v.buf, i: i, end: end, parsed: true}, nil
}

func (v *Value) MustGet(ks ...interface{}) *Value {
	v, err := v.Get(ks...)
	if err != nil {
		panic(err)
	}
	return v
}

func getFromArray(b []byte, k int, i int) (i_ int, e_ error) {
	if b[i] != '[' {
		return i, NewError(b, i, ErrUnexpectedChar)
	}
	i++
	var err error
	for j := 0; ; j++ {
		i = skipSpaces(b, i)
		if i == len(b) {
			return i, NewError(b, i, ErrUnexpectedEnd)
		}
		if b[i] == ']' {
			return i, NewError(b, i, ErrOutOfRange)
		}
		if j != 0 {
			if b[i] != ',' {
				return i, NewError(b, i, ErrUnexpectedChar)
			}
			i++
			i = skipSpaces(b, i)
		}
		if j == k {
			break
		}
		i, err = skipValue(b, i)
		if err != nil {
			return i, err
		}
	}
	return i, nil
}

func getFromObject(b []byte, k []byte, i int) (i_ int, e_ error) {
	if b[i] != '{' {
		return i, NewError(b, i, ErrUnexpectedChar)
	}
	i++
	var err error
	var eq bool
	first := true
	for {
		i = skipSpaces(b, i)
		if i == len(b) {
			return i, NewError(b, i, ErrUnexpectedEnd)
		}
		if b[i] == '}' {
			return i, NewError(b, i, ErrNoSuchKey)
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
		eq, i, err = checkKey(b, k, i)
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
		i = skipSpaces(b, i)
		if eq {
			break
		}
		i, err = skipValue(b, i)
		if err != nil {
			return i, err
		}
	}
	return i, nil
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
