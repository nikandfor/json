package jq

import (
	"strconv"
	"unicode/utf8"

	"github.com/nikandfor/json"
)

type (
	Length struct{}

	Slice struct {
		L, R   int
		Circle bool
		Buf    []byte
	}

	Array struct {
		Filter Filter
		Buf    []byte
	}
)

func (f Length) Next(w, r []byte, st int, state State) ([]byte, int, State, error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	n, i, err := p.Length(r, st)
	if err != nil {
		return w, i, state, pe(err, i)
	}

	w = strconv.AppendInt(w, int64(n), 10)

	return w, i, nil, nil
}

func (f *Slice) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	tp, i, err := p.Type(r, st)
	if err != nil {
		return w, i, nil, pe(err, i)
	}

	switch tp {
	case json.String:
		w, i, err = f.applyString(w, r, st)
		return w, i, nil, err
	case json.Array:
	default:
		return w, i, state, pe(json.ErrType, i)
	}

	n, i, err := p.Length(r, st)
	if err != nil {
		return w, i, nil, pe(err, i)
	}

	left, right := f.leftRight(n)

	w = append(w, '[')

	switch {
	case left == right:
	case left < right:
		w, err = f.applyPart(w, r, st, left, right, true)
		if err != nil {
			return
		}
	case f.Circle:
		w, err = f.applyPart(w, r, st, left, n, true)
		if err != nil {
			return
		}

		w, err = f.applyPart(w, r, st, 0, right, false)
		if err != nil {
			return
		}
	}

	w = append(w, ']')

	return w, i, nil, nil
}

func (f *Slice) applyPart(w, r []byte, st, left, right int, first bool) ([]byte, error) {
	var p json.Parser
	var raw []byte

	i, err := p.Enter(r, st, json.Array)
	if err != nil {
		return w, pe(err, i)
	}

	for n := 0; n < right && p.ForMore(r, &i, json.Array, &err); n++ {
		raw, i, err = p.Raw(r, i)
		if err != nil {
			return w, pe(err, i)
		}

		if n < left {
			continue
		}

		if !first || n != left {
			w = append(w, ',')
		}

		w = append(w, raw...)
	}
	if err != nil {
		return w, pe(err, i)
	}

	return w, nil
}

func (f *Slice) applyString(w, r []byte, st int) (_ []byte, i int, err error) {
	var p json.Parser

	f.Buf, i, err = p.DecodeString(r, st, f.Buf[:0])
	if err != nil {
		return w, i, err
	}

	n := utf8.RuneCount(f.Buf)

	left, right := f.leftRight(n)

	w = append(w, '"')

	switch {
	case left == right:
	case left < right:
		w, err = f.applyStringPart(w, f.Buf, left, right)
		if err != nil {
			return
		}
	case f.Circle:
		w, err = f.applyStringPart(w, f.Buf, left, n)
		if err != nil {
			return
		}

		w, err = f.applyStringPart(w, f.Buf, 0, right)
		if err != nil {
			return
		}
	}

	w = append(w, '"') //, '\n')

	return w, i, nil
}

func (f *Slice) applyStringPart(w, s []byte, l, r int) (_ []byte, err error) {
	var i, st int

	for n := 0; n < r; n++ {
		if n == l {
			st = i
		}

		_, size := utf8.DecodeRune(s[i:])
		i += size
	}

	// TODO: encode string
	w = append(w, s[st:i]...)

	return w, nil
}

func (f *Slice) leftRight(n int) (l, r int) {
	l, r = f.L, f.R

	if l < 0 {
		l = n + l
	}

	if r < 0 {
		r = n + r
	}

	return
}

func (f *Array) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	ff := cfilter(f.Filter, Dot{})

	w = append(w, '[')

	var sub State
	wst := len(w)

	for i = st; ; {
		if wst != len(w) {
			w = append(w, ',')
		}

		wst = len(w)

		w, i, sub, err = ff.Next(w, r, i, sub)
		//	log.Printf("array next %q  i %d  state %v  err %v", w[wst:], i, sub, err)
		if err != nil {
			return w, i, state, err
		}
		if sub == nil {
			break
		}
	}

	if l := len(w) - 1; w[l] == ',' {
		w[l] = ']'
	} else {
		w = append(w, ']')
	}

	return w, i, nil, nil
}
