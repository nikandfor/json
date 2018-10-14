package json

import (
	"strconv"
)

func (it *Iter) CheckString() (string, error) {
	s := it.NextString()
	if it.err != nil {
		return "", it.Err()
	}
	return string(s), nil
}

func (it *Iter) Int64() (int64, error) {
	r := int64(0)
	n := false
	if c := it.b[it.i]; c == '-' {
		n = true
		it.i++
	} else if c == '+' {
		it.i++
	}

	for ; it.i < it.end; it.i++ {
		c := it.b[it.i]
		if c < '0' || c > '9' {
			break
		}
		r = r*10 + (int64)(c-'0')
		if r < 0 {
			it.err = ErrOverflow
			return 0, it.Err()
		}
	}

	if n {
		r = -r
	}

	return r, nil
}

func (it *Iter) Uint64() (uint64, error) {
	r := uint64(0)
	if c := it.b[it.i]; c == '-' {
		it.err = ErrConversion
		return 0, it.Err()
	} else if c == '+' {
		it.i++
	}

	for ; it.i < it.end; it.i++ {
		c := it.b[it.i]
		if c < '0' || c > '9' {
			break
		}
		old := r
		r = r*10 + (uint64)(c-'0')
		if r < old {
			it.err = ErrOverflow
			return 0, it.Err()
		}
	}

	return r, nil
}

func (it *Iter) Float64() (float64, error) {
	v := it.NextNumber()
	return strconv.ParseFloat(string(v), 64)
}

func (it *Iter) Bool() (bool, error) {
	v := it.NextBool()
	if it.err != nil {
		return false, it.Err()
	}
	return v, nil
}
