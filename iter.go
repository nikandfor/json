package json

type ArrayIter struct {
	v    Value
	last int
	err  error
}

func (v *Value) ArrayIter() (*ArrayIter, error) {
	tp, err := v.Type()
	if err != nil {
		return nil, err
	}
	if tp != Array {
		return nil, NewError(v.buf, v.i, ErrConversion)
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
	end, err := skipValue(b, i)
	if err != nil {
		it.err = err
		return nil
	}
	it.last = end
	return &Value{buf: b, i: i, end: end}
}

func (it *ArrayIter) Err() error {
	return it.err
}

type ObjectIter struct {
	v    Value
	last int
	err  error
}

func (v *Value) ObjectIter() (*ObjectIter, error) {
	tp, err := v.Type()
	if err != nil {
		return nil, err
	}
	if tp != Object {
		return nil, NewError(v.buf, v.i, ErrConversion)
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

	end, err := skipValue(b, i)
	if err != nil {
		it.err = err
		return
	}

	k = &Value{buf: b, i: i, end: end}

	i = end
	i = skipSpaces(b, i)
	i++ // ,
	i = skipSpaces(b, i)

	end, err = skipValue(b, i)
	if err != nil {
		it.err = err
		return
	}

	v = &Value{buf: b, i: i, end: end}

	i = end
	it.last = i

	return
}

func (it *ObjectIter) Err() error {
	return it.err
}
