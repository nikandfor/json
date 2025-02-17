package json

// IterFunc iterates over array or object calling f for each element.
// k is always nil while iterating Array.
// k is decoded as Iterator.Key, which means it doesn't decode escape sequences.
// It reads array or object to the end unless f returned an error.
func (d *Iterator) IterFunc(b []byte, st int, tp byte, f func(k, v []byte) error) (i int, err error) {
	var k, v []byte

	i, err = d.Enter(b, st, tp)
	if err != nil {
		return i, err
	}

	for d.ForMore(b, &i, tp, &err) {
		if tp == Object {
			k, i, err = d.Key(b, i)
			if err != nil {
				return i, err
			}
		}

		v, i, err = d.Raw(b, i)
		if err != nil {
			return i, err
		}

		err = f(k, v)
		if err != nil {
			return i, err
		}
	}
	if err != nil {
		return i, err
	}

	return i, nil
}

// IterFunc iterates over array or object calling f for each element.
// k is nil while iterating Array.
// Refer Iterator.IterFunc for more details.
func (r *Reader) IterFunc(tp byte, f func(k, v []byte) error) (err error) {
	var k, v []byte

	err = r.Enter(tp)
	if err != nil {
		return err
	}

	for r.ForMore(tp, &err) {
		if tp == Object {
			k, err = r.Key()
			if err != nil {
				return err
			}
		}

		v, err = r.Raw()
		if err != nil {
			return err
		}

		err = f(k, v)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	return nil
}
