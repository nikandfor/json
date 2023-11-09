package jq

import (
	"fmt"

	"github.com/nikandfor/errors"
	"github.com/nikandfor/json"
)

type (
	Filter interface {
		Next(w, r []byte, st int, state State) ([]byte, int, State, error)
	}

	MapFilter interface {
		Next(w, r []byte, st int, state State) ([]byte, []byte, int, State, error)
	}

	State interface{}

	Dot struct{}

	Empty struct{}

	Literal []byte

	Index []interface{}

	Comma struct {
		Filters []Filter
		pool    []*commaState
	}

	commaState struct {
		fi  int
		sub State
	}

	Pipe struct {
		Filters []Filter
		pool    []*pipeState
	}

	pipeState struct {
		b   []byte
		sub []pipeSub
	}

	pipeSub struct {
		St    int
		State State
		Bst   int
	}

	First struct{}

	Func func(w, r []byte, st int, state State) ([]byte, int, State, error)

	Dumper func(w, r []byte, st, end int)

	ParseError struct {
		Err error
		Pos int
	}
)

func ApplyToAll(f Filter, w, r []byte, st int, sep []byte) ([]byte, error) {
	var err error

	for i := json.SkipSpaces(r, st); i < len(r); i = json.SkipSpaces(r, i) {
		w, i, err = NextAll(f, w, r, i, sep)
		if err != nil {
			return w, err
		}
	}

	return w, nil
}

func NextAll(f Filter, w, r []byte, st int, sep []byte) ([]byte, int, error) {
	var err error
	var state State

	if sep == nil {
		sep = []byte{'\n'}
	}

	wst := len(w)

	for {
		if wst != len(w) {
			w = append(w, sep...)
		}

		wst = len(w)

		w, st, state, err = f.Next(w, r, st, state)
		if err != nil || state == nil {
			return w, st, err
		}
	}
}

func (f Dot) Next(w, r []byte, st int, state State) ([]byte, int, State, error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	raw, i, err := p.Raw(r, st)
	if err != nil {
		return w, i, state, pe(err, i)
	}

	w = append(w, raw...)
	//	w = append(w, '\n')

	return w, i, nil, nil
}

func (f Index) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	if len(f) == 0 {
		raw, i, err := p.Raw(r, st)
		if err != nil {
			return w, i, state, pe(err, i)
		}

		w = append(w, raw...)

		return w, i, nil, nil
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
		return nil, i, state, errors.New("unsupported selector type: %T", f[0])
	}

	i, err = p.Enter(r, st, typ)
	if err != nil {
		return w, i, state, pe(err, i)
	}

	var k []byte

	for p.ForMore(r, &i, typ, &err) {
		if typ == json.Object {
			k, i, err = p.Key(r, i)
			if err != nil {
				return w, i, state, pe(err, i)
			}

			if string(k) != key {
				i, err = p.Skip(r, i)
				if err != nil {
					return w, i, state, pe(err, i)
				}

				continue
			}
		} else if index != 0 {
			index--

			i, err = p.Skip(r, i)
			if err != nil {
				return w, i, state, pe(err, i)
			}

			continue
		}

		w, i, state, err = f[1:].Next(w, r, i, state) // TODO: use sub state if .[] support added
		if err != nil {
			return w, i, state, err
		}

		i, err = p.Break(r, i, 1)
		if err != nil {
			return w, i, state, pe(err, i)
		}

		return w, i, nil, nil
	}
	if err != nil {
		return w, i, state, pe(err, i)
	}

	w = append(w, "null"...) // \n

	return w, i, nil, nil
}

func NewPipe(fs ...Filter) *Pipe {
	return &Pipe{
		Filters: fs,
	}
}

func (f *Pipe) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	ss, _ := state.(*pipeState)
	if ss == nil {
		ss = f.state()
	}

	last := len(f.Filters) - 1
	fi := last

	for fi >= 0 && ss.sub[fi].State == nil {
		fi--
	}

	if fi == -1 && state == nil {
		fi++
	}

	if fi == -1 {
		return w, st, nil, nil
	}

	for ; fi < len(f.Filters); fi++ {
		ff := f.Filters[fi]

		if ss.sub[fi].State == nil {
			ss.sub[fi].Bst = len(ss.b)
		}

		wbst := ss.sub[fi].Bst
		ss.b = ss.b[:wbst]

		wb := cbuf(fi == last, w, ss.b)
		rb := cbuf(fi == 0, r, ss.b)
		rbst := cint(fi == 0, st, ss.sub[fi].St)

		if fi < last {
			ss.sub[fi+1].St = len(ss.b)
		}

		//	log.Printf("pipe args   #%d  %3v  %-30q  %q  sub %+v", fi, rbst, rb, ss.b, ss.sub)

		wb, ss.sub[fi].St, ss.sub[fi].State, err = ff.Next(wb, rb, rbst, ss.sub[fi].State)
		//	log.Printf("pipe filter #%d  %3v  %-30q  %q  err %v", fi, ss.sub[fi].St, wb, ss.b, err)
		if err != nil {
			return w, i, state, err
		}

		w = cbuf(fi == last, wb, w)
		ss.b = cbuf(fi != last, wb, ss.b)
	}

	i = ss.sub[0].St
	state = nil

	for _, sub := range ss.sub {
		if sub.State != nil {
			state = ss
			break
		}
	}

	if state == nil {
		f.pool = append(f.pool, ss)
	}

	return w, i, state, nil
}

func (f *Pipe) state() (ss *pipeState) {
	l := len(f.pool)

	if l == 0 {
		ss = &pipeState{}
	} else {
		ss = f.pool[l-1]
		f.pool = f.pool[:l-1]

		ss.b = ss.b[:0]
	}

	if cap(ss.sub) < len(f.Filters) {
		ss.sub = make([]pipeSub, len(f.Filters))
	}

	ss.sub = ss.sub[:len(f.Filters)]

	return ss
}

func NewComma(fs ...Filter) *Comma {
	return &Comma{
		Filters: fs,
	}
}

func (f *Comma) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	st = json.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	ss, _ := state.(*commaState)
	if ss == nil {
		ss = f.state()
		state = ss
	}

	w, i, ss.sub, err = f.Filters[ss.fi].Next(w, r, st, ss.sub)
	if err != nil {
		return w, i, state, err
	}

	if ss.sub == nil {
		ss.fi++
	}

	if ss.fi == len(f.Filters) {
		state = nil
		f.pool = append(f.pool, ss)
	}

	return w, cint(state != nil, st, i), state, nil
}

func (f *Comma) state() (ss *commaState) {
	l := len(f.pool)
	if l == 0 {
		ss = &commaState{}
	} else {
		ss = f.pool[l-1]
		f.pool = f.pool[:l-1]

		ss.fi = 0
	}

	return ss
}

func (f Empty) Next(w, r []byte, st int, state *State) (_ []byte, i int, _ *State, err error) {
	var p json.Parser

	i, err = p.Skip(r, st)

	return w, i, nil, err
}

func (f Literal) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var p json.Parser

	i, err = p.Skip(r, st)
	if err != nil {
		return w, i, state, err
	}

	w = append(w, f...)

	return w, i, nil, nil
}

func (f First) Next(w, r []byte, st int, state State) ([]byte, int, State, error) {
	var p json.Parser

	st = p.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	raw, i, err := p.Raw(r, st)
	if err != nil {
		return w, i, state, pe(err, i)
	}

	for i = p.SkipSpaces(r, i); i < len(r); i = p.SkipSpaces(r, i) {
		i, err = p.Skip(r, i)
		if err != nil {
			return w, i, state, pe(err, i)
		}
	}

	w = append(w, raw...)

	return w, i, nil, nil
}

func (f Func) Next(w, r []byte, st int, state State) ([]byte, int, State, error) {
	return f(w, r, st, state)
}

func (f Dumper) Next(w, r []byte, st int, state State) ([]byte, int, State, error) {
	st = json.SkipSpaces(r, st)
	if st == len(r) {
		return w, st, nil, nil
	}

	i, err := (&json.Parser{}).Skip(r, st)
	if err != nil {
		return w, st, state, err
	}

	f(w, r, st, i)

	return w, i, nil, nil
}

func pe(err error, i int) error {
	//	tlog.Printw("parse error", "i", i, "err", err, "from", loc.Callers(1, 3))
	return ParseError{Err: err, Pos: i}
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse input: %v (at pos %d)", e.Err.Error(), e.Pos)
}

func (e ParseError) Unwrap() error { return e.Err }

func cint(c bool, x, y int) int {
	if c {
		return x
	}

	return y
}

func cbuf(c bool, x, y []byte) []byte {
	if c {
		return x
	}

	return y
}

func cfilter(f0, f1 Filter) Filter {
	if f0 != nil {
		return f0
	}

	return f1
}
