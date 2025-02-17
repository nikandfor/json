package jq

import "nikand.dev/go/json2"

type (
	Key   string
	Index int
)

func (f Key) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	return keyIndexNext(w, r, st, string(f))
}

func (f Index) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	return keyIndexNext(w, r, st, int(f))
}

func keyIndexNext(w, r []byte, st int, f any) (_ []byte, i int, _ State, err error) {
	var d json2.Iterator

	st = d.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	i, err = d.Seek(r, st, f)
	if err != nil {
		return w, i, nil, err
	}

	raw, i, err := d.Raw(r, i)
	if err != nil {
		return w, i, nil, err
	}

	i, err = d.Break(r, i, 1)
	if err != nil {
		return w, i, nil, err
	}

	w = append(w, raw...)

	return w, i, nil, nil
}
