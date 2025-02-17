package jq

import (
	"bytes"

	"nikand.dev/go/json"
)

type (
	Select struct {
		Filter Filter
	}

	Map struct {
		Filter Filter
		Buf    []byte
		Values bool
	}

	Equal struct {
		L, R Filter
		Not  bool
		Buf  []byte
	}
)

func NewSelect(f Filter) *Select { return &Select{Filter: f} }

func (f *Select) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var p json.Iterator

	wreset := len(w)

	raw, i, err := p.Raw(r, st)
	if err != nil {
		return w, i, state, err
	}

	if f.Filter == nil {
		ok, _, err := IsTrue(raw, 0)
		if err != nil {
			return w, st, state, err
		}

		if ok {
			w = append(w, raw...)
		}

		return w, i, nil, nil
	}

	ff := cfilter(f.Filter, Dot{})

	var sub State
	var ok bool

	for {
		w, i, sub, err = ff.Next(w[:wreset], r, st, sub)
		if err != nil {
			return w, i, state, err
		}

		if len(w[wreset:]) == 0 && sub == nil {
			break
		}
		if len(w[wreset:]) == 0 {
			continue
		}

		ok, _, err = IsTrue(w[wreset:], 0)
		//	log.Printf("select istrue %v %v <- %q", ok, err, f.Buf)
		if err != nil {
			return w, i, state, err
		}
		if ok {
			break
		}

		if sub == nil {
			break
		}
	}

	w = w[:wreset]

	if ok {
		w = append(w, raw...)
	}

	return w, i, nil, nil
}

func NewMap(f Filter) *Map       { return &Map{Filter: f} }
func NewMapValues(f Filter) *Map { return &Map{Filter: f, Values: true} }

func (f *Map) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var p json.Iterator

	ff := f.Filter
	if ff == nil {
		ff = Dot{}
	}

	tp, i, err := p.Type(r, st)
	if err != nil {
		return w, i, state, err
	}

	i, err = p.Enter(r, i, tp)
	if err != nil {
		return w, i, state, err
	}

	var key []byte

	restp := byte(json.Array)

	if tp == json.Object && f.Values {
		restp = json.Object
	}

	w = append(w, restp)
	wkey := len(w)

	for p.ForMore(r, &i, tp, &err) {
		if wkey != len(w) {
			w = append(w, ',')
		}

		wkey = len(w)

		if tp == json.Object {
			key, i, err = p.Key(r, i)
			if err != nil {
				return w, i, state, err
			}
		}

		if tp == json.Object && f.Values {
			w = append(w, key...)
			w = append(w, ':')
		}

		var sub State
		wst := len(w)

		w, i, sub, err = ff.Next(w, r, i, sub)
		if err != nil {
			return w, i, state, err
		}

		//	log.Printf("map f: %s", w[wst:])

		switch {
		case wst == len(w):
			w = w[:wkey-1]
			continue
		case f.Values:
			continue
		}

		for sub != nil {
			if wst != len(w) {
				w = append(w, ',')
			}

			wst = len(w)

			w, i, sub, err = ff.Next(w, r, i, sub)
			if err != nil {
				return w, i, state, err
			}
		}
	}
	if err != nil {
		return w, i, state, err
	}

	w = append(w, restp+2)

	return w, i, nil, nil
}

func NewEqual(l, r Filter) *Equal    { return &Equal{L: l, R: r} }
func NewNotEqual(l, r Filter) *Equal { return &Equal{L: l, R: r, Not: true} }

func (f *Equal) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	f.Buf, i, _, err = f.L.Next(f.Buf[:0], r, st, nil)
	if err != nil {
		return w, st, nil, err
	}

	roff := len(f.Buf)

	f.Buf, _, _, err = f.R.Next(f.Buf, r, st, nil)
	if err != nil {
		return w, st, nil, err
	}

	ok := bytes.Equal(f.Buf[:roff], f.Buf[roff:])

	if ok == !f.Not {
		w = append(w, "true"...)
	} else {
		w = append(w, "false"...)
	}

	return w, i, nil, nil
}

func IsTrue(val []byte, st int) (bool, int, error) {
	var p json.Iterator

	st = p.SkipSpaces(val, st)
	if st == len(val) {
		return false, st, nil
	}

	tp, i, err := p.Type(val, st)
	if err != nil {
		return false, i, err
	}

	raw, i, err := p.Raw(val, i)
	if err != nil {
		return false, i, err
	}

	switch tp {
	case json.Number, json.String, json.Array, json.Object:
		return true, i, nil
	case json.Null:
		return false, i, nil
	case json.Bool:
		return string(raw) == "true", i, nil
	default:
		panic(tp)
	}
}
