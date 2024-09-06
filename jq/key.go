package jq

import "nikand.dev/go/json"

type (
	Key   string
	Index int
)

func (f Key) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var d json.Decoder

	st = d.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	i, err = d.Seek(r, st, string(f))
	if err != nil {
		return w, i, nil, err
	}

	raw, i, err := d.Raw(r, i)
	if err != nil {
		return w, i, nil, err
	}

	w = append(w, raw...)

	return w, i, nil, nil
}

func (f Index) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var d json.Decoder

	st = d.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	i, err = d.Seek(r, st, int(f))
	if err != nil {
		return w, i, nil, err
	}

	raw, i, err := d.Raw(r, i)
	if err != nil {
		return w, i, nil, err
	}

	w = append(w, raw...)

	return w, i, nil, nil
}
