package jq

import "nikand.dev/go/json"

type (
	Cat struct {
		Separator []byte
	}
)

func (f Cat) Next(w, r []byte, st int, _ State) (_ []byte, i int, _ State, err error) {
	var p json.Decoder

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	w = append(w, '"')

	var s []byte
	i = st
	wst := len(w)

	for i < len(r) {
		s, i, err = p.Key(r, i)
		if err != nil {
			return w, i, nil, err
		}

		if wst != len(w) {
			w = append(w, f.Separator...)
		}

		wst = len(w)

		w = append(w, s...)
	}

	w = append(w, '"')

	return w, i, nil, nil
}
