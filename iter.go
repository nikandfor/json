package json

// IterFunc is a little helper on top of Enter and ForMore methonds.
// It iterates over object or array and calls f for each value.
// If it iterates over array k will be nil.
// If it iterates over object k is decoded using Key which doesn't decode escape sequences.
// It reads object or array to the end unless f returned an error.
func (d *Decoder) IterFunc(b []byte, st int, tp byte, f func(k, v []byte) error) (i int, err error) {
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

// IterFunc is a little helper on top of Enter and ForMore methonds.
// See Decoder.IterFunc for more details.
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
