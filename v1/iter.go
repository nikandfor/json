package json

type ArrayIter struct {
	v    Value
	last int
}

func (v *Value) ArrayIter() (*ArrayIter, error) {
	tp, err := v.Type()
	if err != nil {
		return nil, err
	}
	if tp != Array {
		return nil, NewError(v.buf, v.i, ErrConversion)
	}
	if !v.parsed {
		end, err := skipValue(v.buf, v.i)
		if err != nil {
			return nil, err
		}
		v.end = end
		v.parsed = true
	}

	return &ArrayIter{v: *v, last: v.i}, nil
}

func (it *ArrayIter) HasNext() bool {
	i := it.last
	b := it.v.buf

	i = skipSpaces(b, i)
	if b[i] == ']' {
		it.last = i
		return false
	}
	i++ // , or [
	i = skipSpaces(b, i)

	it.last = i

	if b[i] == ']' {
		return false
	}

	return true
}

func (it *ArrayIter) Next() *Value {
	i := it.last
	b := it.v.buf
	end, _ := skipValue(b, i)
	it.last = end
	return &Value{buf: b, i: i, end: end}
}

type ObjectIter struct {
	v    Value
	last int
}

func (v *Value) ObjectIter() (*ObjectIter, error) {
	tp, err := v.Type()
	if err != nil {
		return nil, err
	}
	if tp != Object {
		return nil, NewError(v.buf, v.i, ErrConversion)
	}
	if !v.parsed {
		end, err := skipValue(v.buf, v.i)
		if err != nil {
			return nil, err
		}
		v.end = end
		v.parsed = true
	}

	return &ObjectIter{v: *v, last: v.i}, nil
}

func (it *ObjectIter) HasNext() bool {
	i := it.last
	b := it.v.buf

	i = skipSpaces(b, i)
	if b[i] == '}' {
		it.last = i
		return false
	}
	i++ // , or {
	i = skipSpaces(b, i)

	it.last = i

	if b[i] == '}' {
		return false
	}

	return true
}

func (it *ObjectIter) Next() (k, v *Value) {
	i := it.last
	b := it.v.buf

	end, _ := skipValue(b, i)

	k = &Value{buf: b, i: i, end: end}

	i = end
	i = skipSpaces(b, i)
	i++ // ,
	i = skipSpaces(b, i)

	end, _ = skipValue(b, i)

	v = &Value{buf: b, i: i, end: end}

	i = end
	it.last = i

	return
}
