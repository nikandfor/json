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

func (f Length) Apply(w, r []byte, st int) ([]byte, int, error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil
	}

	n, i, err := p.Length(r, st)
	if err != nil {
		return w, i, pe(err, i)
	}

	w = strconv.AppendInt(w, int64(n), 10)
	//	w = append(w, '\n')

	return w, i, nil
}

func (f *Slice) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil
	}

	tp, i, err := p.Type(r, st)
	if err != nil {
		return w, i, pe(err, i)
	}

	switch tp {
	case json.String:
		return f.applyString(w, r, st)
	case json.Array:
	default:
		return w, i, pe(json.ErrType, i)
	}

	n, i, err := p.Length(r, st)
	if err != nil {
		return w, i, pe(err, i)
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

	w = append(w, ']') //, '\n')

	return w, i, nil
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

func (f *Array) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil
	}

	var vals []byte

	if f.Filter != nil {
		f.Buf = f.Buf[:0]
		f.Buf, i, err = f.Filter.Apply(f.Buf, r, st)

		vals = f.Buf
	} else {
		vals, i, err = p.Raw(r, st)
	}
	if err != nil {
		return w, i, err
	}

	var raw []byte

	w = append(w, '[')

	for j := p.SkipSpaces(vals, 0); j < len(vals); j = p.SkipSpaces(vals, j) {
		if j != 0 {
			w = append(w, ',')
		}

		raw, j, err = p.Raw(vals, j)
		if err != nil {
			return w, i, pe(err, i)
		}

		w = append(w, raw...)
	}

	w = append(w, ']') //, '\n')

	return w, i, nil
}
