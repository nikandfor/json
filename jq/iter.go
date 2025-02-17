package jq

import "nikand.dev/go/json"

type (
	Iter struct{}
)

func (f Iter) Next(w, r []byte, st int, state State) ([]byte, int, State, error) {
	var p json.Iterator

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	var err error
	i := st
	tp, _ := state.(byte)

	if state == nil {
		tp, i, err = p.Type(r, i)
		if err != nil {
			return w, i, state, pe(err, i)
		}

		if tp != json.Array && tp != json.Object {
			return w, i, state, pe(json.ErrType, i)
		}

		i, err = p.Enter(r, i, tp)
		if err != nil {
			return w, i, state, pe(err, i)
		}

		state = tp
	}

	var raw []byte

	for p.ForMore(r, &i, tp, &err) {
		if tp == json.Object {
			i, err = p.Skip(r, i)
			if err != nil {
				return w, i, state, err
			}
		}

		raw, i, err = p.Raw(r, i)
		if err != nil {
			return w, i, state, err
		}

		w = append(w, raw...)

		return w, i, state, nil //nolint:staticcheck
	}
	if err != nil {
		return w, i, state, err
	}

	return w, i, nil, nil
}
