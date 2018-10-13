package json

import (
	"bytes"
	"strconv"
)

func (v *Value) Buffer() []byte {
	return v.buf[v.i:v.end]
}

func (v *Value) String() string {
	if !v.parsed {
		end, err := skipValue(v.buf, v.i)
		if err != nil {
			return string(v.Buffer())
		}
		v.end = end
		v.parsed = true
	}
	if v.buf[v.i] != '"' {
		return string(v.Buffer())
	}
	return string(v.buf[v.i+1 : v.end-1])
}

func (v *Value) CheckString() (string, error) {
	if !v.parsed {
		end, err := skipValue(v.buf, v.i)
		if err != nil {
			return "", err
		}
		v.end = end
		v.parsed = true
	}
	if v.buf[v.i] != '"' {
		return "", ErrConversion
	}
	return string(v.buf[v.i+1 : v.end-1]), nil
}

func (v *Value) MustCheckString() string {
	r, err := v.CheckString()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) Bytes() []byte {
	if !v.parsed {
		end, err := skipValue(v.buf, v.i)
		if err != nil {
			return v.Buffer()
		}
		v.end = end
		v.parsed = true
	}
	if v.buf[v.i] != '"' {
		return v.Buffer()
	}
	return v.buf[v.i+1 : v.end-1]
}

func (v *Value) CheckBytes() ([]byte, error) {
	if !v.parsed {
		end, err := skipValue(v.buf, v.i)
		if err != nil {
			return nil, err
		}
		v.end = end
		v.parsed = true
	}
	if v.buf[v.i] != '"' {
		return nil, ErrConversion
	}
	return v.buf[v.i+1 : v.end-1], nil
}

func (v *Value) MustCheckBytes() []byte {
	r, err := v.CheckBytes()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) Int() (int, error) {
	r, err := v.Int64()
	return int(r), err
}

func (v *Value) MustInt() int {
	r, err := v.Int()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) Uint() (uint, error) {
	r, err := v.Uint64()
	return uint(r), err
}

func (v *Value) MustUint() uint {
	r, err := v.Uint()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) Int64() (int64, error) {
	r := int64(0)
	n := false
	buf := v.buf[v.i:v.end]
	if buf[0] == '-' {
		n = true
		buf = buf[1:]
	} else if buf[0] == '+' {
		buf = buf[1:]
	}
	for _, c := range buf {
		if c < '0' || c > '9' {
			return r, NewError(v.buf, v.i, ErrUnexpectedChar)
		}
		r = r*10 + (int64)(c-'0')
		if r < 0 {
			return r, NewError(v.buf, v.i, ErrOverflow)
		}
	}

	if n {
		r = -r
	}

	return r, nil
}

func (v *Value) MustInt64() int64 {
	r, err := v.Int64()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) Uint64() (uint64, error) {
	r := uint64(0)
	buf := v.buf[v.i:v.end]
	if buf[0] == '-' {
		return 0, NewError(v.buf, v.i, ErrConversion)
	} else if buf[0] == '+' {
		buf = buf[1:]
	}
	for _, c := range buf {
		if c < '0' || c > '9' {
			return r, NewError(v.buf, v.i, ErrUnexpectedChar)
		}
		rp := r
		r = r*10 + (uint64)(c-'0')
		if r < rp {
			return r, NewError(v.buf, v.i, ErrOverflow)
		}
	}

	return r, nil
}

func (v *Value) MustUint64() uint64 {
	r, err := v.Uint64()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) Float64() (float64, error) {
	return strconv.ParseFloat(string(v.buf[v.i:v.end]), 64)
}

func (v *Value) MustFloat64() float64 {
	r, err := v.Float64()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) Bool() (bool, error) {
	if bytes.Equal(v.buf[v.i:v.end], []byte("true")) {
		return true, nil
	}
	if bytes.Equal(v.buf[v.i:v.end], []byte("false")) {
		return false, nil
	}
	return false, NewError(v.buf, v.i, ErrConversion)
}

func (v *Value) IsNull() (bool, error) {
	if bytes.Equal(v.buf[v.i:v.end], []byte("null")) {
		return true, nil
	}
	return false, NewError(v.buf, v.i, ErrConversion)
}

func (v *Value) MustBool() bool {
	r, err := v.Bool()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *Value) CastBool() (bool, error) {
	if !v.parsed {
		end, err := skipValue(v.buf, v.i)
		if err != nil {
			return false, err
		}
		v.end = end
		v.parsed = true
	}
	switch v.buf[v.i] {
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
