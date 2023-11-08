package jq

import (
	"github.com/nikandfor/json"
)

type (
	Select struct {
		Filter Filter
		Buf    []byte
	}

	Map struct {
		Filter Filter
		Buf    []byte
		Values bool
	}
)

func (f *Select) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	var p json.Parser

	raw, i, err := p.Raw(r, st)
	if err != nil {
		return w, i, err
	}

	val := raw

	if f.Filter != nil {
		f.Buf, i, err = f.Filter.Apply(f.Buf[:0], r, st)
		if err != nil {
			return w, i, err
		}

		val = f.Buf
	}

	ok, _, err := IsTrue(val, 0)
	if err != nil {
		return w, st, err
	}

	if ok {
		w = append(w, raw...)
	}

	return w, i, nil
}

func (f *Map) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	var p json.Parser

	ff := f.Filter
	if ff == nil {
		ff = Dot{}
	}

	tp, i, err := p.Type(r, st)
	if err != nil {
		return w, i, err
	}

	i, err = p.Enter(r, i, tp)
	if err != nil {
		return w, i, err
	}

	var key []byte
	comma := false

	restp := byte(json.Array)

	if tp == json.Object && f.Values {
		restp = json.Object
	}

	w = append(w, restp)

	for p.ForMore(r, &i, tp, &err) {
		if comma {
			w = append(w, ',')
		}

		comma = true

		wkey := len(w)

		if tp == json.Object {
			key, i, err = p.Key(r, i)
			if err != nil {
				return w, i, err
			}
		}

		if tp == json.Object && f.Values {
			w = append(w, key...)
			w = append(w, ':')
		}

		wst := len(w)

		w, i, err = ff.Apply(w, r, i)
		if err != nil {
			return w, i, err
		}

		//	log.Printf("map f: %s", w[wst:])

		switch {
		case p.SkipSpaces(w, wst) == len(w):
			w = w[:wkey]

			comma = false
		case f.Values:
			wend, err := p.Skip(w, wst)
			if err != nil {
				return w, i, err
			}

			w = w[:wend]
		default:
			// TODO
		}
	}
	if err != nil {
		return w, i, err
	}

	w = append(w, restp+2)

	return w, i, nil
}

func IsTrue(val []byte, st int) (bool, int, error) {
	var p json.Parser

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
