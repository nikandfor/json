package jq

import (
	"fmt"

	"github.com/nikandfor/errors"
	"github.com/nikandfor/json"
)

type (
	Filter interface {
		Apply(w, r []byte, st int) ([]byte, int, error)
	}

	Dot struct{}

	Empty struct{}

	Literal []byte

	Index []interface{}

	Comma []Filter

	Pipe struct {
		Filters []Filter
		Bufs    [2][]byte
	}

	First struct{}

	Func func(w, r []byte, st int) ([]byte, int, error)

	Dumper func(w, r []byte, st, end int)

	ParseError struct {
		Err error
		Pos int
	}
)

func ApplyToAll(f Filter, w, r []byte, st int) ([]byte, error) {
	var err error
	newline := false

	for i := json.SkipSpaces(r, st); i < len(r); i = json.SkipSpaces(r, i) {
		if newline {
			w = append(w, '\n')
		}

		was := len(w)

		w, i, err = f.Apply(w, r, i)
		if err != nil {
			return w, err
		}

		newline = len(w) != was
	}

	return w, nil
}

func (f Dot) Apply(w, r []byte, st int) ([]byte, int, error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil
	}

	raw, i, err := p.Raw(r, st)
	if err != nil {
		return w, i, pe(err, i)
	}

	w = append(w, raw...)
	//	w = append(w, '\n')

	return w, i, nil
}

func (f Index) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil
	}

	if len(f) == 0 {
		raw, i, err := p.Raw(r, st)
		if err != nil {
			return w, i, pe(err, i)
		}

		w = append(w, raw...)
		//	w = append(w, '\n')

		return w, i, nil
	}

	var typ byte
	var index int
	var key string

	switch x := f[0].(type) {
	case int:
		typ = json.Array
		index = x
	case string:
		typ = json.Object
		key = x
	default:
		return nil, i, errors.New("unsupported selector type: %T", f[0])
	}

	i, err = p.Enter(r, st, typ)
	if err != nil {
		return w, i, pe(err, i)
	}

	var k []byte

	for p.ForMore(r, &i, typ, &err) {
		if typ == json.Object {
			k, i, err = p.Key(r, i)
			if err != nil {
				return w, i, pe(err, i)
			}

			if string(k) != key {
				i, err = p.Skip(r, i)
				if err != nil {
					return w, i, pe(err, i)
				}

				continue
			}
		} else if index != 0 {
			index--

			i, err = p.Skip(r, i)
			if err != nil {
				return w, i, pe(err, i)
			}

			continue
		}

		w, i, err = f[1:].Apply(w, r, i)
		if err != nil {
			return
		}

		i, err = p.Break(r, i, 1)
		if err != nil {
			return w, i, pe(err, i)
		}

		return w, i, nil
	}
	if err != nil {
		return w, i, pe(err, i)
	}

	w = append(w, "null"...) // \n

	return w, i, nil
}

func NewPipe(fs ...Filter) *Pipe {
	return &Pipe{
		Filters: fs,
	}
}

func (f *Pipe) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	st = json.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil
	}

	switch len(f.Filters) {
	case 0:
		// TODO: what to do here?
		return Dot{}.Apply(w, r, st)
	case 1:
		return f.Filters[0].Apply(w, r, st)
	}

	fi := 0
	l := len(f.Filters) - 1

	f.Bufs[fi&1], i, err = f.Filters[0].Apply(f.Bufs[fi&1][:0], r, st)
	if err != nil {
		return w, i, err
	}

	for fi = 1; fi < len(f.Filters); fi++ {
		bw, br := fi&1, 1-fi&1

		var wb []byte

		if fi == l {
			wb = w
		} else {
			wb = f.Bufs[bw][:0]
		}

		wb, err = ApplyToAll(f.Filters[fi], wb, f.Bufs[br], 0)
		if err != nil {
			return w, i, err
		}

		if fi == l {
			w = wb
		} else {
			f.Bufs[bw] = wb
		}
	}

	return w, i, nil
}

func (f Comma) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	st = json.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil
	}

	for fi, ff := range f {
		if fi != 0 {
			w = append(w, '\n')
		}

		w, i, err = ff.Apply(w, r, st)
		if err != nil {
			return w, i, err
		}
	}

	return w, i, nil
}

func (f Empty) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	var p json.Parser

	i, err = p.Skip(r, st)

	return w, i, err
}

func (f Literal) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	var p json.Parser

	i, err = p.Skip(r, st)
	if err != nil {
		return w, i, err
	}

	w = append(w, f...)

	return w, i, nil
}

func (f First) Apply(w, r []byte, st int) ([]byte, int, error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil
	}

	raw, i, err := p.Raw(r, st)
	if err != nil {
		return w, i, pe(err, i)
	}

	for i = p.SkipSpaces(r, i); i < len(r); i = p.SkipSpaces(r, i) {
		i, err = p.Skip(r, i)
		if err != nil {
			return w, i, pe(err, i)
		}
	}

	w = append(w, raw...)
	//	w = append(w, '\n')

	return w, i, nil
}

func (f Func) Apply(w, r []byte, st int) ([]byte, int, error) {
	return f(w, r, st)
}

func (f Dumper) Apply(w, r []byte, st int) ([]byte, int, error) {
	st = json.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil
	}

	raw, i, err := (&json.Parser{}).Raw(r, st)
	if err != nil {
		return w, i, err
	}

	f(w, r, st, i)

	return append(w, raw...), i, nil
}

func pe(err error, i int) error {
	//	tlog.Printw("parse error", "i", i, "err", err, "from", loc.Callers(1, 3))
	return ParseError{Err: err, Pos: i}
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse input: %v (at pos %d)", e.Err.Error(), e.Pos)
}

func (e ParseError) Unwrap() error { return e.Err }
