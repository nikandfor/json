package jq

import "github.com/nikandfor/json"

type (
	Iter struct{}
)

func (f Iter) Apply(w, r []byte, st int) ([]byte, int, error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil
	}

	tp, i, err := p.Type(r, st)
	if err != nil {
		return w, i, pe(err, i)
	}

	if tp != json.Array && tp != json.Object {
		return w, i, pe(json.ErrType, i)
	}

	i, err = p.Enter(r, st, tp)
	if err != nil {
		return w, i, pe(err, i)
	}

	var raw []byte

	for p.ForMore(r, &i, tp, &err) {
		if tp == json.Object {
			i, err = p.Skip(r, i)
			if err != nil {
				return w, i, err
			}
		}

		raw, i, err = p.Raw(r, i)
		if err != nil {
			return w, i, err
		}

		w = append(w, raw...)
		w = append(w, '\n')
	}
	if err != nil {
		return w, i, err
	}

	return w, i, nil
}
