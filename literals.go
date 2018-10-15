package json

import (
	"fmt"
	"strconv"
)

func (r *Reader) IsNull() bool {
	return r.Type() == Null
}

func (r *Reader) CheckString() (string, error) {
	if r.Type() != String {
		return "", r.Err()
	}
	d := r.NextString()
	if r.err != nil {
		return "", r.Err()
	}
	return string(d), nil
}

func (r *Reader) Int64() (int64, error) {
	if r.Type() != Number {
		return 0, r.Err()
	}
	buf := r.NextNumber()
	if r.err != nil {
		return 0, r.Err()
	}

	res := int64(0)
	n := false
	if buf[0] == '-' {
		n = true
		buf = buf[1:]
	} else if buf[0] == '+' {
		buf = buf[1:]
	}
	for _, c := range buf {
		if c < '0' || c > '9' {
			r.err = fmt.Errorf("expected number")
			return 0, r.Err()
		}
		res = res*10 + (int64)(c-'0')
		if res < 0 {
			r.err = fmt.Errorf("type overflow")
			return 0, r.Err()
		}
	}

	if n {
		res = -res
	}

	return res, nil
}

func (r *Reader) Uint64() (uint64, error) {
	if r.Type() != Number {
		return 0, r.Err()
	}
	buf := r.NextNumber()
	if r.err != nil {
		return 0, r.Err()
	}

	res := uint64(0)
	if buf[0] == '-' {
		r.err = fmt.Errorf("negative number")
		return 0, r.Err()
	} else if buf[0] == '+' {
		buf = buf[1:]
	}
	for _, c := range buf {
		if c < '0' || c > '9' {
			r.err = fmt.Errorf("expected number")
			return 0, r.Err()
		}
		res = res*10 + (uint64)(c-'0')
		if res < 0 {
			r.err = fmt.Errorf("type overflow")
			return 0, r.Err()
		}
	}

	return res, nil
}

func (r *Reader) Float64() (float64, error) {
	if r.Type() != Number {
		return 0, r.Err()
	}
	buf := r.NextNumber()
	if r.err != nil {
		return 0, r.Err()
	}

	return strconv.ParseFloat(string(buf), 64)
}

func (r *Reader) Bool() (bool, error) {
	if r.Type() != Bool {
		return false, r.Err()
	}

	c := r.b[r.i]
	if c == 't' {
		r.skip3('r', 'u', 'e')
	} else {
		r.skip4('a', 'l', 's', 'e')
	}

	return c == 't', r.Err()
}

func (r *Reader) Int() (int, error) {
	i, err := r.Int64()
	return (int)(i), err
}

func (r *Reader) Uint() (uint, error) {
	i, err := r.Uint64()
	return (uint)(i), err
}

func (r *Reader) MustInt() int {
	i, err := r.Int()
	if err != nil {
		panic(err)
	}
	return i
}

func (r *Reader) MustUint() uint {
	i, err := r.Uint()
	if err != nil {
		panic(err)
	}
	return i
}

func (r *Reader) MustInt64() int64 {
	i, err := r.Int64()
	if err != nil {
		panic(err)
	}
	return i
}

func (r *Reader) MustUint64() uint64 {
	i, err := r.Uint64()
	if err != nil {
		panic(err)
	}
	return i
}
