package jq

import (
	"fmt"

	"github.com/nikandfor/errors"
	"github.com/nikandfor/json"
	"github.com/nikandfor/loc"
	"tlog.app/go/tlog"
)

type (
	Filter interface {
		Apply(w, r []byte, st int) ([]byte, int, error)
	}

	Dot struct{}

	Selector []interface{}

	Comma []Filter

	Pipe struct {
		Filters []Filter
		Bufs    [2][]byte
	}

	ParseError struct {
		Err error
		Pos int
	}
)

func (f Dot) Apply(w, r []byte, st int) ([]byte, int, error) {
	var p json.Parser

	raw, i, err := p.Raw(r, st)
	if err != nil {
		return w, i, pe(err, i)
	}

	w = append(w, raw...)
	w = append(w, '\n')

	return w, i, nil
}

func (f Selector) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	var p json.Parser

	if len(f) == 0 {
		raw, i, err := p.Raw(r, st)
		if err != nil {
			return w, i, pe(err, i)
		}

		w = append(w, raw...)
		w = append(w, '\n')

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

	w = append(w, "null\n"...)

	return w, i, nil
}

func NewPipe(fs ...Filter) *Pipe {
	return &Pipe{
		Filters: fs,
	}
}

func (f *Pipe) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	var j int

	ri := r
	sti := st

	for fi, ff := range f.Filters {
		var wi []byte

		if fi == len(f.Filters)-1 {
			wi = w
		} else {
			wi = f.Bufs[fi&1][:0]
		}

		wi, j, err = ff.Apply(wi, ri, sti)

		if fi == 0 {
			i = j
		}

		if fi == len(f.Filters)-1 {
			w = wi
		} else {
			f.Bufs[fi&1] = wi
		}

		if err != nil {
			return w, i, err
		}

		ri = wi
		sti = 0
	}

	return w, i, nil
}

func (f Comma) Apply(w, r []byte, st int) (_ []byte, i int, err error) {
	for _, ff := range f {
		w, i, err = ff.Apply(w, r, st)
		if err != nil {
			return w, i, err
		}
	}

	return w, i, nil
}

func pe(err error, i int) error {
	tlog.Printw("parse error", "i", i, "err", err, "from", loc.Callers(1, 3))
	return ParseError{Err: err, Pos: i}
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse input: %v (at pos %d)", e.Err.Error(), e.Pos)
}

func (e ParseError) Unwrap() error { return e.Err }
