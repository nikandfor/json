package jq

import (
	"nikand.dev/go/json"
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

func (f *Select) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var p json.Decoder

	raw, i, err := p.Raw(r, st)
	if err != nil {
		return w, i, state, err
	}

	ff := cfilter(f.Filter, Dot{})

	var sub State
	var ok bool

	for {
		f.Buf, i, sub, err = ff.Next(f.Buf[:0], r, st, sub)
		if err != nil {
			return w, i, state, err
		}

		if len(f.Buf) == 0 && sub == nil {
			break
		}
		if len(f.Buf) == 0 {
			continue
		}

		ok, _, err = IsTrue(f.Buf, 0)
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

	if ok {
		w = append(w, raw...)
	}

	return w, i, nil, nil
}

func (f *Map) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var p json.Decoder

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

func IsTrue(val []byte, st int) (bool, int, error) {
	var p json.Decoder

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
